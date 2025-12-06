package worker

import (
	"ez2boot/internal/config"
	"ez2boot/internal/notification"
	"ez2boot/internal/server"
	"ez2boot/internal/session"
	"ez2boot/internal/user"
	"log/slog"
)

func NewWorker(
	serverService *server.Service,
	sessionService *session.Service,
	userService *user.Service,
	notificationService *notification.Service,
	cfg *config.Config,
	logger *slog.Logger,
) *Worker {
	return &Worker{
		ServerService:       serverService,
		SessionService:      sessionService,
		UserService:         userService,
		NotificationService: notificationService,
		Config:              cfg,
		Logger:              logger,
	}
}
