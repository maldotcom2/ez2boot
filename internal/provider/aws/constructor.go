package aws

import (
	"ez2boot/internal/config"
	"ez2boot/internal/db"
	"ez2boot/internal/server"
	"log/slog"
)

func NewService(awsRepo *Repository, cfg *config.Config, serverService *server.Service, logger *slog.Logger) *Service {
	return &Service{
		Repo:          awsRepo,
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
