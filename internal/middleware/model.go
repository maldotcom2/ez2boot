package middleware

import (
	"ez2boot/internal/config"
	"ez2boot/internal/user"
	"log/slog"
)

type Middleware struct {
	UserService *user.Service
	Config      *config.Config
	Logger      *slog.Logger
}
