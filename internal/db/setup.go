package db

import (
	_ "github.com/mattn/go-sqlite3"
)

func (r *Repository) SetupDB() error {
	// Create servers table
	if _, err := r.DB.Exec("CREATE TABLE IF NOT EXISTS servers (id INTEGER PRIMARY KEY AUTOINCREMENT, unique_id TEXT UNIQUE NOT NULL, name TEXT UNIQUE NOT NULL, state TEXT NOT NULL, next_state TEXT, server_group TEXT NOT NULL, time_added INTEGER NOT NULL, time_last_on INTEGER, time_last_off INTEGER, last_user TEXT)"); err != nil {
		return err
	}

	// Create sessions table
	if _, err := r.DB.Exec("CREATE TABLE IF NOT EXISTS server_sessions (token TEXT PRIMARY KEY, email TEXT NOT NULL, server_group TEXT UNIQUE NOT NULL, expiry INTEGER NOT NULL, to_cleanup INTEGER NOT NULL DEFAULT 0 CHECK (to_cleanup IN (0, 1)), to_notify INTEGER NOT NULL DEFAULT 0 CHECK (to_notify IN (0, 1)), warning_notified INTEGER NOT NULL DEFAULT 0 CHECK (warning_notified IN (0, 1)), on_notified INTEGER NOT NULL DEFAULT 0 CHECK (on_notified IN (0, 1)), off_notified INTEGER NOT NULL DEFAULT 0 CHECK (on_notified IN (0, 1)))"); err != nil {
		return err
	}

	// Create user table
	if _, err := r.DB.Exec("CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY AUTOINCREMENT, email TEXT UNIQUE NOT NULL, password_hash TEXT NOT NULL, is_active INTEGER NOT NULL DEFAULT 1 CHECK (is_active IN (0, 1)), is_admin INTEGER NOT NULL DEFAULT 0 CHECK (is_admin IN (0, 1)), api_enabled INTEGER NOT NULL DEFAULT 0 CHECK (api_enabled IN (0, 1)), ui_enabled INTEGER NOT NULL DEFAULT 1 CHECK (ui_enabled IN (0, 1)))"); err != nil {
		return err
	}

	// create table for user sessions
	if _, err := r.DB.Exec("CREATE TABLE IF NOT EXISTS user_sessions (token_hash TEXT PRIMARY KEY, session_expiry INTEGER NOT NULL, user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE)"); err != nil {
		return err
	}

	// create table for notification queue
	if _, err := r.DB.Exec("CREATE TABLE IF NOT EXISTS notification_queue (user_id PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE, message TEXT NOT NULL, title TEXT NOT NULL, time_added INTEGER NOT NULL)"); err != nil {
		return err
	}

	// create table for user notification
	if _, err := r.DB.Exec("CREATE TABLE IF NOT EXISTS user_notifications (user_id INTEGER PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE, type TEXT NOT NULL, config TEXT NOT NULL)"); err != nil {
		return err
	}

	return nil
}
