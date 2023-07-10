package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/caarlos0/env/v8"
	log "github.com/inconshreveable/log15"
	"github.com/joho/godotenv"
	"github.com/sylvrs/fuse"
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
	Token          string `env:"TOKEN,required"`
	DatabaseString string `env:"DATABASE_STRING,required"`
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
	mng, err := fuse.NewManager(logger, &fuse.Config{
		Token:          envConfig.Token,
		DatabaseString: envConfig.DatabaseString,
	})
	if err != nil {
		logger.Crit("Failed to initialize bot", "error", err)
		os.Exit(1)
	}
	// register services here
	// e.g., mng.RegisterService(&bot.PingService{})

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
