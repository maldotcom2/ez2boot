package server

import (
	"ez2boot/internal/audit"
	"ez2boot/internal/db"
	"log/slog"
)

type Repository struct {
	Base *db.Repository
}

type Service struct {
	Repo   *Repository
	Logger *slog.Logger
}

type Handler struct {
	Service *Service
	Audit   *audit.Service
	Logger  *slog.Logger
}

type Server struct {
	UniqueID    string `json:"unique_id"`
	Name        string `json:"name"`
	State       string `json:"state"`
	ServerGroup string `json:"server_group"`
	TimeAdded   int64  `json:"time_added"`
}
