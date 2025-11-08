package main

import (
	"database/sql"
	"ez2boot/internal/db"
	"log/slog"
	"os"
)

func initDatabase(logger *slog.Logger) (*sql.DB, *db.Repository) {
	// connect to DB
	conn, err := db.Connect()
	if err != nil {
		logger.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}

	// Shared base repo
	repo := db.NewRepository(conn, logger)

	return conn, repo
}
