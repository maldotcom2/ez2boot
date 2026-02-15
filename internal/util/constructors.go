package util

import (
	"ez2boot/internal/config"
	"ez2boot/internal/db"
	"log/slog"
)

func NewHandler(utilService *Service, logger *slog.Logger) *Handler {
	return &Handler{
		Service: utilService,
		Logger:  logger,
	}
}

func NewService(utilRepo *Repository, config *config.Config, buildInfo BuildInfo, logger *slog.Logger) *Service {
	return &Service{
		Repo:      utilRepo,
		Config:    config,
		BuildInfo: buildInfo,
		Logger:    logger,
	}
}

func NewRepository(base *db.Repository) *Repository {
	return &Repository{
		Base: base,
	}
}
