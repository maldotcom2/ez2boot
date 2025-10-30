package model

import (
	"log/slog"
	"time"
)

type Config struct {
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

type Server struct {
	UniqueID    string `json:"unique_id"`
	Name        string `json:"name"`
	State       string `json:"state"`
	ServerGroup string `json:"server_group"`
	TimeAdded   int64  `json:"time_added"`
}
