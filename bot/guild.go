package bot

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	log "github.com/inconshreveable/log15"
	"github.com/sylvrs/fuse/bot/command"
	"github.com/sylvrs/fuse/bot/component"
	"github.com/sylvrs/fuse/bot/modal"
	"github.com/sylvrs/fuse/bot/utils"
	"gorm.io/gorm"
)

// GuildConfiguration is the structure that holds the configuration for a single guild
type GuildConfiguration struct {
	// GuildID is the ID of the guild and is used as the primary key
	GuildID string `gorm:"primarykey"`
}

// GuildManager is the structure that manages all of the services for a single guild
// It holds the guild's ID, its configuration, the database connection, and the Discord session
type GuildManager struct {
	logger             log.Logger
	config             *GuildConfiguration
	manager            *Manager
	connection         *gorm.DB
	session            *discordgo.Session
	guild              *discordgo.Guild
	commandHandler     *command.CommandHandler
	modalHandler       *modal.ModalHandler
	listenedComponents map[string]component.ComponentHandlerFunc
	services           []Service
}

func CreateGuildManager(manager *Manager, config *GuildConfiguration) (*GuildManager, error) {
	guild, err := manager.session.State.Guild(config.GuildID)
	if err != nil {
		return nil, fmt.Errorf("failed to get guild by id %s", config.GuildID)
	}

	commandHandler, err := command.NewCommandHandler(manager.session, guild)
	if err != nil {
		return nil, err
	}
	guildManager := &GuildManager{
		logger:             log.New("guild", config.GuildID),
		config:             config,
		manager:            manager,
		connection:         manager.connection,
		session:            manager.session,
		guild:              guild,
		commandHandler:     commandHandler,
		modalHandler:       modal.NewModalHandler(manager.session, guild),
		listenedComponents: make(map[string]component.ComponentHandlerFunc),
		services:           make([]Service, 0),
	}
	// register services to guild manager
	guildManager.services = manager.CreateServices(guildManager)
	return guildManager, nil
}

func (mng *GuildManager) Start() error {
	for _, service := range mng.services {
		if err := service.Start(mng); err != nil {
			return err
		}
	}

	err := mng.commandHandler.Init()
	if err != nil {
		return err
	}
	mng.AddHandler(mng.handleListenedComponents)
	return nil
}

func (mng *GuildManager) Stop() error {
	for _, service := range mng.services {
		if err := service.Stop(mng); err != nil {
			return err
		}
	}

	mng.commandHandler.Deinit()
	return nil
}

func (mng *GuildManager) Save() error {
	return mng.Connection().Save(mng.config).Error
}

func (mng *GuildManager) Logger() log.Logger {
	return mng.logger
}

func (mng *GuildManager) GlobalManager() *Manager {
	return mng.manager
}

func (mng *GuildManager) Connection() *gorm.DB {
	return mng.connection
}

func (mng *GuildManager) Session() *discordgo.Session {
	return mng.session
}

func (mng *GuildManager) BotUser() *discordgo.User {
	return mng.GlobalManager().BotUser()
}

func (mng *GuildManager) Guild() *discordgo.Guild {
	return mng.guild
}

func (mng *GuildManager) CommandHandler() *command.CommandHandler {
	return mng.commandHandler
}

func (mng *GuildManager) ModalHandler() *modal.ModalHandler {
	return mng.modalHandler
}

// AddHandler is a wrapper for the session handler but limits the handler to only the guild
// This may seem excessive but it is a good practice to prevent accidental checking of the wrong guild
func (mng *GuildManager) AddHandler(handler interface{}) {
	switch handler := handler.(type) {
	case func(s *discordgo.Session, i *discordgo.MessageCreate):
		mng.session.AddHandler(func(s *discordgo.Session, i *discordgo.MessageCreate) {
			if i.GuildID != mng.guild.ID {
				return
			}
			handler(s, i)
		})
	case func(s *discordgo.Session, i *discordgo.InteractionCreate):
		mng.session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if i.GuildID != mng.guild.ID {
				return
			}
			handler(s, i)
		})
	case func(s *discordgo.Session, i *discordgo.GuildCreate):
		mng.session.AddHandler(func(s *discordgo.Session, i *discordgo.GuildCreate) {
			if i.Guild.ID != mng.guild.ID {
				return
			}
			handler(s, i)
		})
	case func(s *discordgo.Session, i *discordgo.GuildDelete):
		mng.session.AddHandler(func(s *discordgo.Session, i *discordgo.GuildDelete) {
			if i.Guild.ID != mng.guild.ID {
				return
			}
			handler(s, i)
		})
	case func(s *discordgo.Session, i *discordgo.GuildRoleUpdate):
		mng.session.AddHandler(func(s *discordgo.Session, i *discordgo.GuildRoleUpdate) {
			if i.GuildID != mng.guild.ID {
				return
			}
			handler(s, i)
		})
	case func(s *discordgo.Session, i *discordgo.ChannelDelete):
		mng.session.AddHandler(func(s *discordgo.Session, i *discordgo.ChannelDelete) {
			if i.GuildID != mng.guild.ID {
				return
			}
			handler(s, i)
		})
	case func(s *discordgo.Session, i *discordgo.GuildMemberAdd):
		mng.session.AddHandler(func(s *discordgo.Session, i *discordgo.GuildMemberAdd) {
			if i.GuildID != mng.guild.ID {
				return
			}
			handler(s, i)
		})
	default:
		mng.logger.Warn("guild handler will not check for guild id", "handler", handler)
		mng.session.AddHandler(handler)
	}
}

func (mng *GuildManager) handleListenedComponents(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionMessageComponent {
		return
	}
	data := i.MessageComponentData()
	handler, ok := mng.listenedComponents[data.CustomID]
	if !ok {
		return
	}
	resp, err := handler(i, &data)
	if err != nil {
		mng.logger.Error("failed to handle component", "error", err)
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:  discordgo.MessageFlagsEphemeral,
				Embeds: []*discordgo.MessageEmbed{utils.ErrorAsEmbed(err.Error())},
			},
		})
		return
	}
	if resp != nil {
		if err := s.InteractionRespond(i.Interaction, resp); err != nil {
			mng.logger.Error("failed to respond to interaction", "error", err)
		}
	}
}

func (mng *GuildManager) ListenForComponent(customId string, handler component.ComponentHandlerFunc) {
	mng.listenedComponents[customId] = handler
}
