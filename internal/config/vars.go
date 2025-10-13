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

	scrapeInterval, err := GetDurationFromString(scrapeIntervalStr)
	if err != nil {
		return model.Config{}, err
	}

	internalClockStr := os.Getenv("INTERNAL_CLOCK")
	if scrapeIntervalStr == "" {
		scrapeIntervalStr = "10s" //default
	}

	internalClock, err := GetDurationFromString(internalClockStr)
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

	logLevelStr := os.Getenv("LOG_LEVEL")
	if logLevelStr == "" {
		logLevelStr = "info" //default
	}

	logLevel := ParseLogLevel(logLevelStr)

	return model.Config{
		CloudProvider:  cloudProvider,
		Port:           port,
		ScrapeInterval: scrapeInterval,
		InternalClock:  internalClock,
		TagKey:         tagKey,
		AWSRegion:      awsRegion,
		LogLevel:       logLevel,
	}, nil
}
