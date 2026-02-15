package main

import (
	"ez2boot/internal/config"
	"log/slog"
	"os"
)

func initLogger(cfg *config.Config) *slog.Logger {
	// Create log handler
	logHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     cfg.LogLevel,
		AddSource: true,
	})

	// create logger
	logger := slog.New(logHandler)
	logger.Info("Start app")
	logger.Info("Log Level", "level", cfg.LogLevel)

	return logger
}
