package repository

import (
	_ "github.com/mattn/go-sqlite3"
)

func (r *Repository) SetupDB() error {
	// Create servers table
	_, err := r.DB.Exec("CREATE TABLE IF NOT EXISTS servers (id INTEGER PRIMARY KEY AUTOINCREMENT, unique_id TEXT UNIQUE, name TEXT UNIQUE, state TEXT, next_state TEXT, server_group TEXT, time_added INTEGER, time_last_on INTEGER, time_last_off INTEGER, last_user TEXT, UNIQUE (unique_id, name))")
	if err != nil {
		return err
	}

	// Create sessions table
	_, err = r.DB.Exec("CREATE TABLE IF NOT EXISTS sessions (token TEXT PRIMARY KEY, email TEXT, server_group TEXT UNIQUE, expiry INTEGER, to_cleanup INTEGER NOT NULL DEFAULT 0 CHECK (to_cleanup IN (0, 1)), to_notify INTEGER NOT NULL DEFAULT 0 CHECK (to_notify IN (0, 1)), warning_notified INTEGER NOT NULL DEFAULT 0 CHECK (warning_notified IN (0, 1)), on_notified INTEGER NOT NULL DEFAULT 0 CHECK (on_notified IN (0, 1)))")
	if err != nil {
		return err
	}

	// Create user table
	_, err = r.DB.Exec("CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY AUTOINCREMENT, username TEXT, password_hash TEXT, is_active INTEGER)")
	if err != nil {
		return err
	}

	return nil
}
