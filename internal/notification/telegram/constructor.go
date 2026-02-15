package telegram

import (
	"ez2boot/internal/db"
	"log/slog"
)

func NewHandler(telegramService *Service, logger *slog.Logger) *Handler {
	return &Handler{
		Service: telegramService,
		Logger:  logger,
	}
}

func NewService(telegramRepo *Repository, logger *slog.Logger) *Service {
	return &Service{
		Repo:   telegramRepo,
		Logger: logger,
	}
}

func NewRepository(base *db.Repository) *Repository {
	return &Repository{
		Base: base,
	}
}
