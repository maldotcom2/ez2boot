package config

import (
	"log/slog"
	"net/http"
	"strings"
	"time"
)

// Get a duration type from a string expression eg "4h"
func GetDurationFromString(strValue string) (time.Duration, error) {
	duration, err := time.ParseDuration(strValue)
	if err != nil {
		return 0, err
	}

	return duration, nil
}

func ParseLogLevel(strValue string) slog.Level {
	switch strings.ToLower(strValue) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		// Fallback to info if unknown
		return slog.LevelInfo
	}
}

func ParseSameSiteMode(strValue string) http.SameSite {
	switch strings.ToLower(strValue) {
	case "lax":
		return http.SameSiteLaxMode
	case "strict":
		return http.SameSiteStrictMode
	case "none":
		return http.SameSiteNoneMode
	default:
		// Fallback to lax if unknown
		return http.SameSiteLaxMode
	}
}
