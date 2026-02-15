package db

import (
	"database/sql"
	"log/slog"
)

// Constructor returns a new Repository that uses the provided DB handle
func NewRepository(db *sql.DB, logger *slog.Logger) *Repository {
	return &Repository{
		DB:     db,
		Logger: logger,
	}
}
