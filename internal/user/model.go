package user

import (
	"ez2boot/internal/db"
	"ez2boot/internal/model"
	"log/slog"
)

type Repository struct {
	Base   *db.Repository
	Logger *slog.Logger
}

type Service struct {
	Repo   *Repository
	Config *model.Config
	Logger *slog.Logger
}

type Handler struct {
	Service *Service
	Logger  *slog.Logger
}
