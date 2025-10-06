package repository

import (
	"database/sql"
	"log/slog"
)

type Repository struct {
	DB     *sql.DB // Handle to the shared database connection pool
	Logger *slog.Logger
}
