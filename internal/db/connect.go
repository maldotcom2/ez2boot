package db

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3" // Requires gcc compiler on path
)

// Connect to DB and return pointer to connection pool
func Connect() (*sql.DB, error) {
	dbDir := "/data"
	dbPath := filepath.Join(dbDir, "ez2boot.sqlite")

	// Make sure the directory exists
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	// Test the connection
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	// Enable foreign keys
	_, err = db.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		return nil, err
	}

	return db, nil
}
