package notification

import (
	"ez2boot/internal/audit"
	"ez2boot/internal/db"
	"log/slog"
)

func NewHandler(notificationService *Service, audit *audit.Service, logger *slog.Logger) *Handler {
	return &Handler{
		Service: notificationService,
		Audit:   audit,
		Logger:  logger,
	}
}

func NewService(notificationRepo *Repository, logger *slog.Logger) *Service {
	return &Service{
		Repo:   notificationRepo,
		Logger: logger,
	}
}

func NewRepository(base *db.Repository) *Repository {
	return &Repository{
		Base: base,
	}
}
