package middleware

import (
	"ez2boot/internal/user"
	"log/slog"
)

type Middleware struct {
	Service *user.Service
	Logger  *slog.Logger
}
