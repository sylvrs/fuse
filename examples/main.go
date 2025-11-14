package main

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/caarlos0/env/v8"
	"github.com/glebarez/sqlite"
	log "github.com/inconshreveable/log15"
	"github.com/joho/godotenv"
	"github.com/sylvrs/fuse"
	"github.com/sylvrs/fuse/command"
	"github.com/sylvrs/fuse/utils"
)

const (
	// The path to write the log file to
	logPath = "bot.log"
)

// Global variables
var (
	// The logger used for logging to the console and Discord
	logger log.Logger
)

// environmentConfig is the struct used to parse the environment variables
type environmentConfig struct {
	Token        string `env:"TOKEN,required"`
	DatabaseName string `env:"DATABASE_NAME,required"`
}

// devEnvironmentConfig is the struct used to parse the development environment variables
type devEnvironmentConfig struct {
	Config environmentConfig `envPrefix:"DEV_"`
}

// initialize logging
func init() {
	// set up initial logging suite w/ file handler
	logger = log.New()
	fileHandler, _ := log.FileHandler(logPath, log.LogfmtFormat())
	logger.SetHandler(log.MultiHandler(
		log.StderrHandler,
		fileHandler,
	))
}

// initialize env variables
func init() {
	if err := godotenv.Load(); err != nil {
		logger.Info("No .env file found. Environment variables will be used instead.")
	}
}

func main() {
	// get if we are in production or not
	production := os.Getenv("PRODUCTION") == "true"

	// get the environment variables based on if we are in production or not
	var envConfig environmentConfig
	if production {
		if err := env.Parse(&envConfig); err != nil {
			logger.Crit("Failed to parse production environment variables", "error", err)
			os.Exit(1)
		}
	} else {
		var devConfig devEnvironmentConfig
		if err := env.Parse(&devConfig); err != nil {
			logger.Crit("Failed to parse development environment variables", "error", err)
			os.Exit(1)
		}
		envConfig = devConfig.Config
	}

	dialector := sqlite.Open(envConfig.DatabaseName)
	mng, err := fuse.NewManager(dialector, logger, &fuse.Config{
		Token: envConfig.Token,
	})
	if err != nil {
		logger.Crit("Failed to initialize bot", "error", err)
		os.Exit(1)
	}
	// register services here
	mng.RegisterService(&PingService{})

	if err := mng.Start(); err != nil {
		logger.Crit("Failed to start service", "error", err)
		os.Exit(1)
	}

	// wait for a signal to exit
	exit := make(chan os.Signal, 1)
	signal.Notify(exit, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-exit

	logger.Info("Received shutdown signal. Closing Discord session...")
	mng.Stop()
}

// PingService is a service that responds to the ping command
type PingService struct {
	config PingServiceConfiguration
}

// PingServiceConfiguration represents a table in the database that holds the configuration for the ping service in each guild
type PingServiceConfiguration struct {
	fuse.ServiceConfiguration
	RandomNumber int
	CachedInput  fuse.StringArray `gorm:"type:TEXT"`
}

func (s *PingService) Create(mng *fuse.GuildManager) (fuse.Service, error) {
	var config PingServiceConfiguration

	number := rand.Intn(128)
	logger.Info("Generated number", "number", number)
	// fetch the service config from the database and assign a random number to it
	if err := mng.FetchServiceConfig(&config, PingServiceConfiguration{RandomNumber: number}); err != nil {
		return nil, err
	}

	logger.Info("Fetched number", "number", config.RandomNumber)
	return &PingService{config: config}, nil
}

func (s *PingService) Start(mng *fuse.GuildManager) error {

	permission := int64(discordgo.PermissionAdministrator)
	mng.CommandHandler().Register(&command.Command{
		Name:               "cache",
		Description:        "Adds a given input to the database for later listing",
		DefaultPermissions: &permission,
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "value",
				Description: "The value to cache",
				Required:    true,
			},
		},
		Handler: func(_ *discordgo.Session, i *discordgo.InteractionCreate) (*discordgo.InteractionResponse, error) {
			return s.HandleCacheCommand(mng, i)
		},
	})
	mng.CommandHandler().Register(&command.Command{
		Name:               "viewcache",
		Description:        "Lists the values in the cache",
		DefaultPermissions: &permission,
		Options:            []*discordgo.ApplicationCommandOption{},
		Handler: func(_ *discordgo.Session, i *discordgo.InteractionCreate) (*discordgo.InteractionResponse, error) {
			return s.HandleViewCacheCommand(mng, i)
		},
	})

	// ...
	return nil
}

func (s *PingService) HandleCacheCommand(mng *fuse.GuildManager, i *discordgo.InteractionCreate) (*discordgo.InteractionResponse, error) {
	value := i.ApplicationCommandData().Options[0].StringValue()
	if value == "" {
		return nil, errors.New("Empty input given")
	}

	s.config.CachedInput = append(s.config.CachedInput, value)
	mng.SaveServiceConfig(s.config)

	return &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				utils.SuccessAsEmbed(fmt.Sprintf("Added %s to cache list", value)),
			},
		},
	}, nil
}

func (s *PingService) HandleViewCacheCommand(mng *fuse.GuildManager, i *discordgo.InteractionCreate) (*discordgo.InteractionResponse, error) {
	data := ""
	for _, line := range s.config.CachedInput {
		data += fmt.Sprintf("- %s\n", line)
	}

	return &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				utils.SuccessAsEmbed(data),
			},
		},
	}, nil

}

func (s *PingService) Stop(mng *fuse.GuildManager) error {
	// ...
	return nil
}
