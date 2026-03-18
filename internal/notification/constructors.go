package notification

import (
	"ez2boot/internal/audit"
	"ez2boot/internal/db"
	"log/slog"
)

func NewHandler(notificationService *Service, logger *slog.Logger) *Handler {
	return &Handler{
		Service: notificationService,
		Logger:  logger,
	}
}

func NewService(notificationRepo *Repository, audit *audit.Service, encryptor Encryptor, logger *slog.Logger) *Service {
	return &Service{
		Repo:      notificationRepo,
		Audit:     audit,
		Encryptor: encryptor,
		Logger:    logger,
	}
}

func NewRepository(base *db.Repository) *Repository {
	return &Repository{
		Base: base,
	}
}
