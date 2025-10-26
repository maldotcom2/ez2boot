package aws

import (
	"ez2boot/internal/db"
	"ez2boot/internal/model"
	"ez2boot/internal/server"
	"log/slog"
)

type Repository struct {
	Base *db.Repository
}

type Service struct {
	Repo          *Repository
	Config        *model.Config
	ServerService *server.Service
	Logger        *slog.Logger
}
