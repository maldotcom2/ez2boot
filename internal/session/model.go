package session

import (
	"ez2boot/internal/audit"
	"ez2boot/internal/db"
	"ez2boot/internal/notification"
	"ez2boot/internal/user"
	"log/slog"
	"time"
)

type Repository struct {
	Base *db.Repository
}

type Service struct {
	Repo                *Repository
	NotificationService *notification.Service
	UserService         *user.Service
	Audit               *audit.Service
	Logger              *slog.Logger
}

type Handler struct {
	Service *Service
	Logger  *slog.Logger
}

type ServerSession struct {
	Id          int64     `json:"-"`
	UserID      int64     `json:"-"`
	Email       string    `json:"-"`
	ServerGroup string    `json:"server_group"`
	Duration    string    `json:"duration"`
	Expiry      time.Time `json:"expiry"`
}

type ServerSessionRequest struct {
	UserID      int64  `json:"-"`
	ServerGroup string `json:"server_group"`
	Duration    string `json:"duration"`
	Expiry      int64  `json:"-"`
}

type ServerSessionResponse struct {
	ServerGroup string    `json:"server_group"`
	Duration    string    `json:"duration"`
	Expiry      time.Time `json:"expiry"`
}

type ServerSessionSummary struct {
	ServerGroup string  `json:"server_group"`
	ServerCount int64   `json:"server_count"`
	ServerNames string  `json:"server_names"`
	CurrentUser *string `json:"current_user"` // Can be null
	Expiry      *int64  `json:"expiry"`       // Can be null
}
