package config

import (
	"ez2boot/internal/model"
	"os"

	"github.com/joho/godotenv"
)

func GetEnvVars() (model.Config, error) {
	err := godotenv.Load()
	if err != nil {
		return model.Config{}, err
	}

	cloudProvider := os.Getenv("CLOUD_PROVIDER")
	if cloudProvider == "" {
		cloudProvider = "aws"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	scrapeInterval := os.Getenv("SCRAPE_INTERVAL")
	if scrapeInterval == "" {
		scrapeInterval = "30"
	}

	return model.Config{
		CloudProvider:  cloudProvider,
		Port:           port,
		ScrapeInterval: scrapeInterval,
	}, nil
}
