package config

import (
	"ez2boot/internal/model"
	"ez2boot/internal/utils"
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
		cloudProvider = "aws" // default
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000" //default
	}

	scrapeIntervalStr := os.Getenv("SCRAPE_INTERVAL")
	if scrapeIntervalStr == "" {
		scrapeIntervalStr = "30s" //default
	}

	scrapeInterval, err := utils.GetDurationFromString(scrapeIntervalStr)
	if err != nil {
		return model.Config{}, err
	}

	tagKey := os.Getenv("TAG_KEY")
	if tagKey == "" {
		tagKey = "ez2boot" //default
	}

	awsRegion := os.Getenv("AWS_REGION")
	if awsRegion == "" {
		awsRegion = "ap-southeast-2" //default
	}

	return model.Config{
		CloudProvider:  cloudProvider,
		Port:           port,
		ScrapeInterval: scrapeInterval,
		TagKey:         tagKey,
		AWSRegion:      awsRegion,
	}, nil
}
