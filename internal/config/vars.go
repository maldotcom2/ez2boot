package config

import (
	"ez2boot/internal/model"
	"os"

	"github.com/joho/godotenv"
)

func GetEnvVars() (*model.Config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, err
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
		return nil, err
	}

	internalClockStr := os.Getenv("INTERNAL_CLOCK")
	if scrapeIntervalStr == "" {
		scrapeIntervalStr = "10s" //default
	}

	internalClock, err := GetDurationFromString(internalClockStr)
	if err != nil {
		return nil, err
	}

	tagKey := os.Getenv("TAG_KEY")
	if tagKey == "" {
		tagKey = "ez2boot" //default
	}

	awsRegion := os.Getenv("AWS_REGION")
	if awsRegion == "" {
		awsRegion = "ap-southeast-2" //default
	}

	userNotifications := os.Getenv("USER_NOTIFICATIONS")
	if userNotifications == "" {
		userNotifications = "disabled" //default
	}

	logLevelStr := os.Getenv("LOG_LEVEL")
	if logLevelStr == "" {
		logLevelStr = "info" //default
	}

	logLevel := ParseLogLevel(logLevelStr)

	userSessionDurationStr := os.Getenv("USER_SESSION_DURATION")
	if userSessionDurationStr == "" {
		userSessionDurationStr = "6h" //default
	}

	userSessionDuration, err := GetDurationFromString(userSessionDurationStr)
	if err != nil {
		return nil, err
	}

	cfg := &model.Config{
		CloudProvider:       cloudProvider,
		Port:                port,
		ScrapeInterval:      scrapeInterval,
		InternalClock:       internalClock,
		TagKey:              tagKey,
		AWSRegion:           awsRegion,
		UserNotifications:   userNotifications,
		UserSessionDuration: userSessionDuration,
		LogLevel:            logLevel,
	}

	return cfg, nil
}
