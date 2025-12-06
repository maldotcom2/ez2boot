package middleware

import (
	"ez2boot/internal/config"
	"ez2boot/internal/user"
	"log/slog"
)

func NewMiddleware(userService *user.Service, cfg *config.Config, logger *slog.Logger) *Middleware {
	return &Middleware{
		UserService: userService,
		Config:      cfg,
		Logger:      logger,
	}
}
