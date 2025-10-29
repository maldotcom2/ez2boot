package session

import (
	"ez2boot/internal/db"
	"log/slog"
	"time"
)

type Repository struct {
	Base *db.Repository
}

type Service struct {
	Repo   *Repository
	Logger *slog.Logger
}

type Handler struct {
	Service *Service
	Logger  *slog.Logger
}

type ServerSession struct {
	Email       string    `json:"email"`
	ServerGroup string    `json:"server_group"`
	Token       string    `json:"token"`
	Duration    string    `json:"duration"`
	Expiry      time.Time `json:"expiry"`
	Message     string    `json:"message"`
}
