package server

import (
	"ez2boot/internal/audit"
	"ez2boot/internal/db"
	"log/slog"
)

func NewHandler(serverService *Service, audit *audit.Service, logger *slog.Logger) *Handler {
	return &Handler{
		Service: serverService,
		Audit:   audit,
		Logger:  logger,
	}
}

func NewService(serverRepo *Repository, logger *slog.Logger) *Service {
	return &Service{
		Repo:   serverRepo,
		Logger: logger,
	}
}

func NewRepository(base *db.Repository) *Repository {
	return &Repository{
		Base: base,
	}
}
