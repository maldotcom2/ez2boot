package main

import (
	"ez2boot/internal/config"
	"log"
	"os"
)

func loadVars() *config.Config {
	cfg, err := config.GetEnvVars()
	if err != nil {
		log.Print("Could not load environment variables, check that .env file is present or that env variables have been configured")
		os.Exit(1)
	}

	return cfg
}
