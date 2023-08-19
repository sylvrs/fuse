package fuse

import (
	"errors"
	"fmt"
	"reflect"

	log "github.com/inconshreveable/log15"
	"github.com/sylvrs/fuse/utils"
	mysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
	gorm_logger "gorm.io/gorm/logger"

	"github.com/bwmarrin/discordgo"
)

// Config is the structure that holds the configuration for the manager
// It holds the token used to authenticate with Discord and the string used to connect to the database
type Config struct {
	// The token used to authenticate with Discord
	Token string
	// The string used to connect to the database
	DatabaseString string
}

type ManagerStartFunc func(*Manager) error

// Manager is the overarching structure that manages all of the guild sub-services
// Moreover, it holds the database connection and handles all Discord events
type Manager struct {
	logger        log.Logger
	config        *Config
	connection    *gorm.DB
	session       *discordgo.Session
	guildManagers map[string]*GuildManager
	onStartFuncs  []ManagerStartFunc
	services      []Service
}

func NewManager(logger log.Logger, config *Config) (*Manager, error) {
	// initialize database
	database, err := gorm.Open(mysql.Open(config.DatabaseString), &gorm.Config{
		Logger: gorm_logger.Default.LogMode(gorm_logger.Silent),
	})
	if err != nil {
		return nil, err
	}
	// create discord session
	session, err := discordgo.New("Bot " + config.Token)
	if err != nil {
		return nil, err
	}
	session.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsAll)

	return &Manager{
		logger:        logger,
		config:        config,
		connection:    database,
		guildManagers: make(map[string]*GuildManager),
		session:       session,
		onStartFuncs:  make([]ManagerStartFunc, 0),
		services:      make([]Service, 0),
	}, nil
}

func (mng *Manager) OnStart(f ManagerStartFunc) {
	mng.onStartFuncs = append(mng.onStartFuncs, f)
}

// RegisterService registers a service to be created when the manager starts
// In most cases, an empty struct should be passed in as the argument
// This is because the actual guild services will be created using service.Create()
func (mng *Manager) RegisterService(s Service) {
	mng.services = append(mng.services, s)
	mng.logger.Info(fmt.Sprintf("Registered service '%s'", reflect.TypeOf(s).Name()))
}

// CreateServices creates a service for a provided guild manager
func (mng *Manager) CreateServices(guildManager *GuildManager) ([]Service, error) {
	services := make([]Service, len(mng.services))
	for i, s := range mng.services {
		service, err := s.Create(guildManager)
		if err != nil {
			mng.logger.Error("Failed to create service", "service", reflect.TypeOf(s).Name(), "error", err)
			return nil, err
		}
		services[i] = service
	}
	return services, nil
}

func (mng *Manager) Member(guildID string) (*discordgo.Member, error) {
	member, err := mng.Session().State.Member(guildID, mng.Session().State.User.ID)
	if err != nil {
		return nil, err
	}
	return member, nil
}

func (mng *Manager) Start() error {
	if err := mng.session.Open(); err != nil {
		return err
	}

	// load guilds before doing anything else
	if err := mng.loadGuilds(); err != nil {
		return err
	}

	// create handlers
	mng.setupHandlers()

	// run start functions
	for _, f := range mng.onStartFuncs {
		if err := f(mng); err != nil {
			return errors.New("failed to run start function: " + err.Error())
		}
	}

	mng.logger.Info(fmt.Sprintf("Logged in as %s#%s", mng.session.State.User.Username, mng.session.State.User.Discriminator))
	return nil
}

func (mng *Manager) onGuildLoad(event *discordgo.GuildCreate) {
	mng.logger.Info("Loaded guild", "guild", event.ID)
}

func (mng *Manager) onGuildJoin(event *discordgo.GuildCreate) {
	mng.Logger().Info("Joined guild", "guild", event.ID)
	guild := GuildConfiguration{
		GuildID: event.ID,
	}
	mng.connection.Limit(1).Find(&guild)
	// if guild already exists, don't do anything
	if mng.connection.RowsAffected != 0 {
		mng.Logger().Info("Guild already exists in database", "guild", event.ID)
		return
	}

	if _, err := mng.createGuild(event.Guild); err != nil {
		mng.Logger().Error("Failed to create guild", "guild", event.ID, "error", err)
		return
	}
}

func (mng *Manager) onGuildLeave(event *discordgo.GuildDelete) {
	mng.Logger().Info("Left guild", "guild", event.ID)
	if err := mng.deleteGuild(event.Guild); err != nil {
		mng.Logger().Error("Failed to delete guild", "guild", event.ID, "error", err)
	}
}

func (mng *Manager) onReceiveCommand(event *discordgo.InteractionCreate) {
	guildManager, ok := mng.guildManagers[event.GuildID]
	if !ok {
		mng.logger.Error("Failed to find guild manager for guild", "guild", event.GuildID)
		return
	}
	mng.logger.Debug(fmt.Sprintf("Received command for guild %s", guildManager.Guild().Name), "command", event.ApplicationCommandData().Name)
	res, err := guildManager.commandHandler.Handle(mng.session, event)
	if err != nil {
		mng.logger.Error("Failed to handle command", "command", event.ApplicationCommandData().Name, "error", err)
		mng.session.InteractionRespond(event.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags: discordgo.MessageFlagsEphemeral,
				Embeds: []*discordgo.MessageEmbed{
					utils.ErrorAsEmbed(err.Error()),
				},
			},
		})
		return
	}
	// if response is nil, don't respond
	if res == nil {
		return
	}
	if err = mng.session.InteractionRespond(event.Interaction, res); err != nil {
		mng.logger.Error("Failed to respond to interaction", "error", err)
	}
}

func (mng *Manager) onReceiveModal(event *discordgo.InteractionCreate) {
	guildManager, ok := mng.guildManagers[event.GuildID]
	if !ok {
		mng.logger.Error("Failed to find guild manager for guild", "guild", event.GuildID)
		return
	}
	mng.logger.Debug(fmt.Sprintf("Received modal for guild %s", guildManager.Guild().Name), "modal", event.ModalSubmitData().CustomID)
	res, err := guildManager.modalHandler.Handle(event)
	if err != nil {
		mng.logger.Error("Failed to handle modal", "error", err)
		mng.session.InteractionRespond(event.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags: discordgo.MessageFlagsEphemeral,
				Embeds: []*discordgo.MessageEmbed{
					utils.ErrorAsEmbed(err.Error()),
				},
			},
		})
		return
	}
	// if response is nil, don't respond
	if res == nil {
		return
	}
	if err = mng.session.InteractionRespond(event.Interaction, res); err != nil {
		mng.logger.Error("Failed to respond to interaction", "error", err)
	}
}

func (mng *Manager) setupHandlers() {
	mng.session.AddHandler(func(s *discordgo.Session, event *discordgo.GuildCreate) {
		if mng.GuildExists(event.Guild.ID) {
			mng.onGuildLoad(event)
			return
		}
		mng.onGuildJoin(event)
	})
	mng.session.AddHandler(func(s *discordgo.Session, event *discordgo.GuildDelete) { mng.onGuildLeave(event) })
	mng.session.AddHandler(func(s *discordgo.Session, event *discordgo.InteractionCreate) {
		switch event.Type {
		case discordgo.InteractionApplicationCommand:
			mng.onReceiveCommand(event)
		case discordgo.InteractionModalSubmit:
			mng.onReceiveModal(event)
		}
	})
	mng.logger.Info("Registered event handlers")
}

func (mng *Manager) loadGuilds() error {
	// load guilds from database
	var guilds []*GuildConfiguration
	if err := mng.connection.Find(&guilds).Error; err != nil {
		return err
	}

	// create guild managers
	for _, guild := range guilds {
		guildManager, err := mng.createGuildManager(guild)
		if err != nil {
			mng.logger.Error("Failed to create guild manager", "guild", guild.GuildID, "error", err)
			continue
		}

		if err := guildManager.Start(); err != nil {
			mng.logger.Error("Failed to start guild manager", "guild", guild.GuildID, "error", err)
		}
	}

	mng.logger.Info(fmt.Sprintf("Loaded %d %s", len(mng.guildManagers), utils.Pluralize(len(mng.guildManagers), "guild", "guilds")))
	return nil
}

// createGuildManager creates a guild manager for a provided guild configuration
func (mng *Manager) createGuildManager(guild *GuildConfiguration) (*GuildManager, error) {
	// create guild manager
	guildManager, err := CreateGuildManager(mng, guild)
	if err != nil {
		return nil, err
	}
	// add guild manager to map
	mng.guildManagers[guild.GuildID] = guildManager
	// setup guild manager
	return guildManager, nil
}

// createGuild creates a guild in the database and creates a guild manager for it
func (mng *Manager) createGuild(guild *discordgo.Guild) (*GuildManager, error) {
	guildManager, err := mng.createGuildManager(&GuildConfiguration{
		GuildID: guild.ID,
	})
	if err != nil {
		return nil, err
	}
	// create guild in database
	if err := mng.connection.Create(&guildManager.config).Error; err != nil {
		return nil, err
	}
	// start guild manager
	if err := guildManager.Start(); err != nil {
		return nil, err
	}
	mng.logger.Info(fmt.Sprintf("Created guild %s (id: %s)", guild.Name, guild.ID))
	return guildManager, nil
}

func (mng *Manager) deleteGuild(guild *discordgo.Guild) error {
	// delete guild from database
	if err := mng.connection.Where("guild_id = ?", guild.ID).Delete(&GuildConfiguration{}).Error; err != nil {
		return err
	}
	guildManager := mng.guildManagers[guild.ID]
	// stop guild manager
	guildManager.Stop()
	// delete guild manager from map
	delete(mng.guildManagers, guild.ID)
	mng.logger.Info(fmt.Sprintf("Deleted guild %s (id: %s)", guild.Name, guild.ID))
	return nil
}

func (mng *Manager) GuildExists(guildID string) bool {
	_, ok := mng.guildManagers[guildID]
	return ok
}

func (mng *Manager) GuildManager(guildID string) (*GuildManager, error) {
	guildManager, ok := mng.guildManagers[guildID]
	if !ok {
		return nil, fmt.Errorf("guild manager not found for guild %s", guildID)
	}
	return guildManager, nil
}

func (mng *Manager) Stop() {
	// handle stopping for guilds
	for _, guildManager := range mng.guildManagers {
		guildManager.Stop()
	}
	mng.session.Close()
}

func (mng *Manager) Logger() log.Logger {
	return mng.logger
}

func (mng *Manager) Config() *Config {
	return mng.config
}

func (mng *Manager) Session() *discordgo.Session {
	return mng.session
}

func (mng *Manager) Connection() *gorm.DB {
	return mng.connection
}

func (mng *Manager) BotUser() *discordgo.User {
	return mng.session.State.User
}
