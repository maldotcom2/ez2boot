package aws

import (
	"ez2boot/internal/config"
	"ez2boot/internal/db"
	"ez2boot/internal/server"
	"log/slog"
)

type Repository struct {
	Base *db.Repository
}

type Service struct {
	Repo          *Repository
	Config        *config.Config
	ServerService *server.Service
	Logger        *slog.Logger
}
