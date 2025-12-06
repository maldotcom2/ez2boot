package user

import (
	"ez2boot/internal/config"
	"ez2boot/internal/db"
	"log/slog"
)

func NewHandler(userService *Service, logger *slog.Logger) *Handler {
	return &Handler{
		Service: userService,
		Logger:  logger,
	}
}

func NewService(userRepo *Repository, cfg *config.Config, logger *slog.Logger) *Service {
	return &Service{
		Repo:   userRepo,
		Config: cfg,
		Logger: logger,
	}
}

func NewRepository(base *db.Repository, logger *slog.Logger) *Repository {
	return &Repository{
		Base:   base,
		Logger: logger,
	}
}
