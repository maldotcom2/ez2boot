package middleware

import (
	"ez2boot/internal/user"
	"log/slog"
)

type Middleware struct {
	UserService *user.Service
	Logger      *slog.Logger
}
