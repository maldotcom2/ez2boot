package repository

import (
	"database/sql"
	"log/slog"
)

type Repository struct {
	DB     *sql.DB // Handle to the shared database connection pool
	Logger *slog.Logger
}

// Constructor returns a new Repository that uses the provided DB handle
func NewRepository(db *sql.DB, logger *slog.Logger) *Repository {
	return &Repository{
		DB:     db,
		Logger: logger,
	}
}
