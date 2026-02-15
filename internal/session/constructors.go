package session

import (
	"ez2boot/internal/audit"
	"ez2boot/internal/db"
	"ez2boot/internal/notification"
	"ez2boot/internal/user"
	"log/slog"
)

func NewHandler(sessionService *Service, logger *slog.Logger) *Handler {
	return &Handler{
		Service: sessionService,
		Logger:  logger,
	}
}

func NewService(sessionRepo *Repository, notificationService *notification.Service, userService *user.Service, audit *audit.Service, logger *slog.Logger) *Service {
	return &Service{
		Repo:                sessionRepo,
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
