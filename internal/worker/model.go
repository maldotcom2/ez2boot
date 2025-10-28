package worker

import (
	"ez2boot/internal/model"
	"ez2boot/internal/server"
	"ez2boot/internal/session"
	"ez2boot/internal/user"
	"log/slog"
)

type Worker struct {
	ServerService  *server.Service
	SessionService *session.Service
	UserService    *user.Service
	Config         *model.Config
	Logger         *slog.Logger
}
