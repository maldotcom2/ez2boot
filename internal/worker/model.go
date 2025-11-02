package worker

import (
	"ez2boot/internal/config"
	"ez2boot/internal/notification"
	"ez2boot/internal/server"
	"ez2boot/internal/session"
	"ez2boot/internal/user"
	"log/slog"
)

type Worker struct {
	ServerService       *server.Service
	SessionService      *session.Service
	UserService         *user.Service
	NotificationService *notification.Service
	Config              *config.Config
	Logger              *slog.Logger
}
