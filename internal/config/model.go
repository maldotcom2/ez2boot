package config

import (
	"log/slog"
	"time"
)

type Config struct {
	SetupMode           bool
	TrustProxyHeaders   bool
	CloudProvider       string
	Port                string
	ScrapeInterval      time.Duration
	InternalClock       time.Duration
	TagKey              string
	AWSRegion           string
	UserNotifications   string
	UserSessionDuration time.Duration
	LogLevel            slog.Level
	EncryptionKey       string
	// Add more fields as needed
}
