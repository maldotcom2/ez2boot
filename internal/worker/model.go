package worker

import (
	"ez2boot/internal/model"
	"ez2boot/internal/server"
	"ez2boot/internal/session"
	"log/slog"
)

type Worker struct {
	ServerService  *server.Service
	SessionService *session.Service
	Config         *model.Config
	Logger         *slog.Logger
}
