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

	// Check server group exists
	var count int
	err = tx.QueryRow("SELECT COUNT(*) FROM servers WHERE server_group = $1", session.ServerGroup).Scan(&count)
	if err != nil {
		tx.Rollback()
		return model.Session{}, err
	}

	if count == 0 {
		tx.Rollback()
		return model.Session{}, fmt.Errorf("No servers found in server_group: %s", session.ServerGroup) // TODO Make this detectable by handler
	}

	newExpiry, err := GetExpiryFromDuration(0, session.Duration)
	if err != nil {
		return model.Session{}, err
	}

	// Convert epoch to time and add to struct
	session.Expiry = time.Unix(newExpiry, 0).UTC()

	_, err = tx.Exec("INSERT INTO sessions (token, email, server_group, expiry) VALUES ($1, $2, $3, $4)", session.Token, session.Email, session.ServerGroup, newExpiry)
	if err != nil {
		tx.Rollback()
		return model.Session{}, err
	}

	err = tx.Commit()
	if err != nil {
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

	// Check number of rows affected
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
