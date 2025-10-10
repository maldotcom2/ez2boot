package config

import (
	"log/slog"
	"strings"
	"time"
)

func GetDurationFromString(strValue string) (time.Duration, error) {
	scrapeInterval, err := time.ParseDuration(strValue)
	if err != nil {
		return 0, err
	}

	return scrapeInterval, nil
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
