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
	// Add more fields as needed
}

type Server struct {
	UniqueID    string `json:"unique_id"`
	Name        string `json:"name"`
	State       string `json:"state"`
	ServerGroup string `json:"server_group"`
	TimeAdded   int64  `json:"time_added"`
}

type ServerSession struct {
	Email       string    `json:"email"`
	ServerGroup string    `json:"server_group"`
	Token       string    `json:"token"`
	Duration    string    `json:"duration"`
	Expiry      time.Time `json:"expiry"`
	Message     string    `json:"message"`
}

type UserLogin struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type ChangePasswordRequest struct {
	Email       string `json:"email"`
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

type UserSession struct {
	SessionExpiry int64
	UserID        string
	Email         string
	Password      string
	Token         string
}
