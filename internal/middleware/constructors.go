package middleware

import (
	"ez2boot/internal/config"
	"ez2boot/internal/user"
	"log/slog"
	"time"

	"golang.org/x/time/rate"
)

func NewMiddleware(userService *user.Service, cfg *config.Config, logger *slog.Logger) *Middleware {
	return &Middleware{
		UserService: userService,
		Config:      cfg,
		PublicRateLimiter: NewRateLimiter(RateLimitConfig{
			Rate:         rate.Limit(cfg.PublicRateLimit),
			Burst:        cfg.PublicRateLimit * 2,
			CleanupEvery: time.Minute,
			ExpiresAfter: 10 * time.Minute,
		}),
		PrivateRateLimiter: NewRateLimiter(RateLimitConfig{
			Rate:         rate.Limit(cfg.PrivateRateLimit),
			Burst:        cfg.PrivateRateLimit * 2,
			CleanupEvery: time.Minute,
			ExpiresAfter: 3 * time.Minute,
		}),
		Logger: logger,
	}
}
