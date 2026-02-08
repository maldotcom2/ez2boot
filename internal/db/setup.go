package db

import (
	_ "github.com/mattn/go-sqlite3"
)

// First time table setup
func (r *Repository) SetupDB() error {
	// Create audit table
	if _, err := r.DB.Exec("CREATE TABLE IF NOT EXISTS audit_log (id INTEGER PRIMARY KEY AUTOINCREMENT, actor_user_id INTEGER NOT NULL, actor_email TEXT NOT NULL, target_user_id INTEGER, target_email TEXT, action TEXT NOT NULL, resource TEXT NOT NULL, success BOOLEAN NOT NULL CHECK (success IN (0, 1)), reason TEXT, metadata TEXT, time_stamp INTEGER NOT NULL)"); err != nil {
		return err
	}

	// Create servers table
	if _, err := r.DB.Exec("CREATE TABLE IF NOT EXISTS servers (unique_id TEXT PRIMARY KEY, name TEXT NOT NULL, state TEXT NOT NULL, next_state TEXT, server_group TEXT NOT NULL, time_added INTEGER NOT NULL, time_last_on INTEGER, time_last_off INTEGER, last_user_id INTEGER)"); err != nil {
		return err
	}

	// Create table for server sessions
	if _, err := r.DB.Exec("CREATE TABLE IF NOT EXISTS server_sessions (id INTEGER PRIMARY KEY AUTOINCREMENT, user_id INTEGER NOT NULL REFERENCES users(id), server_group TEXT UNIQUE NOT NULL, expiry INTEGER NOT NULL, to_cleanup INTEGER NOT NULL DEFAULT 0 CHECK (to_cleanup IN (0, 1)), warning_notified INTEGER NOT NULL DEFAULT 0 CHECK (warning_notified IN (0, 1)), on_notified INTEGER NOT NULL DEFAULT 0 CHECK (on_notified IN (0, 1)), off_notified INTEGER NOT NULL DEFAULT 0 CHECK (off_notified IN (0, 1)))"); err != nil {
		return err
	}

	// Create user table
	if _, err := r.DB.Exec("CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY AUTOINCREMENT, email TEXT UNIQUE NOT NULL, password_hash TEXT, is_active INTEGER NOT NULL DEFAULT 1 CHECK (is_active IN (0, 1)), is_admin INTEGER NOT NULL DEFAULT 0 CHECK (is_admin IN (0, 1)), api_enabled INTEGER NOT NULL DEFAULT 0 CHECK (api_enabled IN (0, 1)), ui_enabled INTEGER NOT NULL DEFAULT 1 CHECK (ui_enabled IN (0, 1)), identity_provider TEXT NOT NULL, last_login INTEGER)"); err != nil {
		return err
	}

	// create table for user sessions
	if _, err := r.DB.Exec("CREATE TABLE IF NOT EXISTS user_sessions (token_hash TEXT PRIMARY KEY, session_expiry INTEGER NOT NULL, user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE)"); err != nil {
		return err
	}

	// create table for notification queue
	if _, err := r.DB.Exec("CREATE TABLE IF NOT EXISTS notification_queue (id INTEGER PRIMARY KEY AUTOINCREMENT, user_id INTEGER REFERENCES users(id) ON DELETE CASCADE, message TEXT NOT NULL, title TEXT NOT NULL, time_added INTEGER NOT NULL)"); err != nil {
		return err
	}

	// create table for user notification
	if _, err := r.DB.Exec("CREATE TABLE IF NOT EXISTS user_notifications (user_id INTEGER PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE, type TEXT NOT NULL, config BLOB NOT NULL)"); err != nil {
		return err
	}

	// create table for version
	if _, err := r.DB.Exec("CREATE TABLE IF NOT EXISTS version (id INTEGER PRIMARY KEY, latest_version TEXT, checked_at INTEGER, release_url TEXT)"); err != nil {
		return err
	}

	return nil
}
