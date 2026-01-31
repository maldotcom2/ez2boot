package server

import (
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
	Logger  *slog.Logger
}

type ServerState string

const (
	ServerOn            ServerState = "on"
	ServerOff           ServerState = "off"
	ServerTransitioning ServerState = "transitioning"
)

type Server struct {
	UniqueID    string      `json:"unique_id"`
	Name        string      `json:"name"`
	State       ServerState `json:"state"`
	ServerGroup string      `json:"server_group"`
	TimeAdded   int64       `json:"time_added"`
}
