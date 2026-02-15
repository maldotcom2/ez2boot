package worker

import (
	"ez2boot/internal/config"
	"ez2boot/internal/notification"
	"ez2boot/internal/server"
	"ez2boot/internal/session"
	"ez2boot/internal/user"
	"ez2boot/internal/util"
	"log/slog"
)

type Worker struct {
	ServerService       *server.Service
	SessionService      *session.Service
	UserService         *user.Service
	NotificationService *notification.Service
	UtilService         *util.Service
	Config              *config.Config
	Logger              *slog.Logger
}
