package session

import (
	"ez2boot/internal/audit"
	"ez2boot/internal/config"
	"ez2boot/internal/db"
	"ez2boot/internal/notification"
	"ez2boot/internal/user"
	"log/slog"
)

func NewHandler(sessionService *Service, cfg *config.Config, logger *slog.Logger) *Handler {
	return &Handler{
		Service: sessionService,
		Config:  cfg,
		Logger:  logger,
	}
}

func NewService(sessionRepo *Repository, cfg *config.Config, notificationService *notification.Service, userService *user.Service, audit *audit.Service, logger *slog.Logger) *Service {
	return &Service{
		Repo:                sessionRepo,
		Config:              cfg,
		NotificationService: notificationService,
		UserService:         userService,
		Audit:               audit,
		Logger:              logger,
	}
}

func NewRepository(base *db.Repository) *Repository {
	return &Repository{
		Base: base,
	}
}
