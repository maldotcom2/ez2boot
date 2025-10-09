package repository

import (
	_ "github.com/mattn/go-sqlite3"
)

func (r *Repository) SetupDB() error {
	// Create servers table
	_, err := r.DB.Exec("CREATE TABLE IF NOT EXISTS servers (id INTEGER PRIMARY KEY AUTOINCREMENT, unique_id TEXT UNIQUE, name TEXT UNIQUE, state TEXT, server_group TEXT, time_added INTEGER, time_last_on INTEGER, time_last_off INTEGER, last_user TEXT)")
	if err != nil {
		return err
	}

	// Create sessions table
	_, err = r.DB.Exec("CREATE TABLE IF NOT EXISTS sessions (token TEXT PRIMARY KEY, email TEXT, server_group TEXT UNIQUE, expiry INTEGER)")
	if err != nil {
		return err
	}

	return nil
}
