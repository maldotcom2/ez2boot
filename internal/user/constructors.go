package user

import (
	"ez2boot/internal/audit"
	"ez2boot/internal/config"
	"ez2boot/internal/db"
	"log/slog"
)

func NewHandler(userService *Service, cfg *config.Config, logger *slog.Logger) *Handler {
	return &Handler{
		Service: userService,
		Config:  cfg,
		Logger:  logger,
	}
}

func NewService(userRepo *Repository, cfg *config.Config, audit *audit.Service, logger *slog.Logger) *Service {
	return &Service{
		Repo:   userRepo,
		Config: cfg,
		Audit:  audit,
		Logger: logger,
	}
}

func NewRepository(base *db.Repository, logger *slog.Logger) *Repository {
	return &Repository{
		Base:   base,
		Logger: logger,
	}
}
