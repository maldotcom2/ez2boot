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

func NewWorker(
	serverService *server.Service,
	sessionService *session.Service,
	userService *user.Service,
	notificationService *notification.Service,
	utilService *util.Service,
	cfg *config.Config,
	logger *slog.Logger,
) *Worker {
	return &Worker{
		ServerService:       serverService,
		SessionService:      sessionService,
		UserService:         userService,
		NotificationService: notificationService,
		UtilService:         utilService,
		Config:              cfg,
		Logger:              logger,
	}
}
