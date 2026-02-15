package email

import (
	"ez2boot/internal/db"
	"log/slog"
)

func NewHandler(emailService *Service, logger *slog.Logger) *Handler {
	return &Handler{
		Service: emailService,
		Logger:  logger,
	}
}

func NewService(emailRepo *Repository, logger *slog.Logger) *Service {
	return &Service{
		Repo:   emailRepo,
		Logger: logger,
	}
}

func NewRepository(base *db.Repository) *Repository {
	return &Repository{
		Base: base,
	}
}
