package repository

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3" // Requires gcc compiler on path
)

// Connect to DB and return pointer to connection pool
func Connect() (*sql.DB, error) {
	dbPath := "../../data/ez2boot.sqlite"
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
