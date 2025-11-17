package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

func GetEnvVars() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		log.Print("Could not load .env file, assuming env vars from other means")
	}

	val := os.Getenv("TRUST_PROXY_HEADERS")
	if val == "" {
		val = "true" // default
	}

	trustProxyHeaders, err := strconv.ParseBool(val)
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
	if internalClockStr == "" {
		internalClockStr = "10s" //default
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

	userSessionDurationStr := os.Getenv("USER_SESSION_DURATION")
	if userSessionDurationStr == "" {
		userSessionDurationStr = "6h" //default
	}

	userSessionDuration, err := GetDurationFromString(userSessionDurationStr)
	if err != nil {
		return nil, err
	}

	logLevelStr := os.Getenv("LOG_LEVEL")
	if logLevelStr == "" {
		logLevelStr = "info" //default
	}

	logLevel := ParseLogLevel(logLevelStr)

	encryptionKey := os.Getenv("ENCRYPTION_KEY") // optional

	rateLimitStr := os.Getenv("RATE_LIMIT")
	if rateLimitStr == "" {
		rateLimitStr = "20" //default
	}

	rateLimit, err := strconv.Atoi(rateLimitStr)
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		TrustProxyHeaders:   trustProxyHeaders,
		CloudProvider:       cloudProvider,
		Port:                port,
		ScrapeInterval:      scrapeInterval,
		InternalClock:       internalClock,
		TagKey:              tagKey,
		AWSRegion:           awsRegion,
		UserSessionDuration: userSessionDuration,
		LogLevel:            logLevel,
		EncryptionKey:       encryptionKey,
		RateLimit:           rateLimit,
	}

	return cfg, nil
}
