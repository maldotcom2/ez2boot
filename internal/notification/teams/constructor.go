package teams

import (
	"ez2boot/internal/db"
	"log/slog"
)

func NewHandler(teamsService *Service, logger *slog.Logger) *Handler {
	return &Handler{
		Service: teamsService,
		Logger:  logger,
	}
}

func NewService(teamsRepo *Repository, logger *slog.Logger) *Service {
	return &Service{
		Repo:   teamsRepo,
		Logger: logger,
	}
}

func NewRepository(base *db.Repository) *Repository {
	return &Repository{
		Base: base,
	}
}
