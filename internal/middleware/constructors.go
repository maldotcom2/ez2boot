package middleware

import (
	"ez2boot/internal/audit"
	"ez2boot/internal/config"
	"ez2boot/internal/user"
	"log/slog"
)

func NewMiddleware(userService *user.Service, cfg *config.Config, audit *audit.Service, logger *slog.Logger) *Middleware {
	return &Middleware{
		UserService: userService,
		Config:      cfg,
		Audit:       audit,
		Logger:      logger,
	}
}
