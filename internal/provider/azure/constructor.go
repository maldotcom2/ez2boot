package azure

import (
	"ez2boot/internal/config"
	"ez2boot/internal/db"
	"ez2boot/internal/server"
	"log/slog"
)

func NewService(azureRepo *Repository, cfg *config.Config, serverService *server.Service, logger *slog.Logger) *Service {
	return &Service{
		Repo:          azureRepo,
		Config:        cfg,
		ServerService: serverService,
		Logger:        logger,
	}
}

func NewRepository(base *db.Repository) *Repository {
	return &Repository{
		Base: base,
	}
}
