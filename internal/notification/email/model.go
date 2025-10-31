package email

import (
	"ez2boot/internal/db"
	"log/slog"
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

type Config struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	To       string `json:"to"`
	From     string `json:"from"`
	User     string `json:"user"`
	Password string `json:"password"`
	Auth     bool   `json:"auth"`
}

type EmailNotification struct{}
