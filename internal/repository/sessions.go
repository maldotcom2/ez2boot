package repository

import (
	"ez2boot/internal/model"
	"fmt"
	"time"
)

// Return currently active sessions
func (r *Repository) GetSessions() ([]model.Session, error) {
	rows, err := r.DB.Query("SELECT email, server_group, expiry FROM sessions")
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	sessions := []model.Session{}

	for rows.Next() {
		var email string
		var serverGroup string
		var expiryInt int64

		err = rows.Scan(&email, &serverGroup, &expiryInt)
		if err != nil {
			return nil, err
		}

		s := model.Session{
			Email:       email,
			ServerGroup: serverGroup,
			Expiry:      time.Unix(expiryInt, 0).UTC(),
		}
		sessions = append(sessions, s)
	}

	return sessions, nil
}

// Create a new session
func (r *Repository) NewSession(session model.Session) (model.Session, error) {
	tx, err := r.DB.Begin()
	if err != nil {
		return model.Session{}, err
	}

	// Set server table for state worker
	result, err := tx.Exec("UPDATE servers SET next_state = $1, time_last_on = $2, last_user = $3 WHERE server_group = $4", "on", time.Now().Unix(), session.Email, session.ServerGroup)
	if err != nil {
		tx.Rollback()
		return model.Session{}, err
	}

	// Impact check
	rows, err := result.RowsAffected()
	if err != nil {
		tx.Rollback()
		return model.Session{}, err
	}

	if rows == 0 {
		tx.Rollback()
		return model.Session{}, fmt.Errorf("No servers found for server_group: %s", session.ServerGroup)
	}

	sessionExpiry, err := GetExpiryFromDuration(0, session.Duration)
	if err != nil {
		tx.Rollback()
		return model.Session{}, err
	}

	// Convert epoch to time and add to struct
	session.Expiry = time.Unix(sessionExpiry, 0).UTC()

	if _, err = tx.Exec("INSERT INTO sessions (token, email, server_group, expiry, to_notify) VALUES ($1, $2, $3, $4, $5)", session.Token, session.Email, session.ServerGroup, sessionExpiry, 1); err != nil {
		tx.Rollback()
		return model.Session{}, err
	}

	if err = tx.Commit(); err != nil {
		return model.Session{}, err
	}

	return session, nil
}

// Update existing session
func (r *Repository) UpdateSession(session model.Session) (bool, model.Session, error) {
	newExpiry, err := GetExpiryFromDuration(0, session.Duration)
	if err != nil {
		return false, session, err
	}

	// Convert epoch to time and add to struct
	session.Expiry = time.Unix(newExpiry, 0).UTC()

	result, err := r.DB.Exec("UPDATE sessions SET expiry = $1 WHERE token = $2", newExpiry, session.Token)
	if err != nil {
		// TO DO: Add error for non-unique where server group already has a session
		return false, session, err
	}

	// Impact check
	rows, err := result.RowsAffected()
	if err != nil {
		return false, model.Session{}, err
	}

	if rows == 0 {
		session.Message = "Session not found"
		return false, model.Session{}, nil
	}

	session.Message = "Successfully updated session"
	return true, session, nil
}

// Full tx for end of session to maintain ref integrity
func (r *Repository) EndSession(serverGroup string) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}

	// Set server next state
	if _, err = tx.Exec("UPDATE servers SET next_state = $1, time_last_off = $2 WHERE server_group = $3", "off", time.Now().Unix(), serverGroup); err != nil {
		tx.Rollback()
		return err
	}

	// Set notify flag on session
	if _, err = tx.Exec("UPDATE sessions SET to_notify = $1 WHERE server_group = $2", 1, serverGroup); err != nil {
		tx.Rollback()
		return err
	}

	// Delete token to invalidate session and prevent changes
	if _, err = tx.Exec("UPDATE sessions SET token = NULL WHERE server_group = $1", serverGroup); err != nil {
		tx.Rollback()
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}
