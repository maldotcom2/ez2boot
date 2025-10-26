package session

import (
	"ez2boot/internal/model"
	"ez2boot/internal/util"
	"fmt"
	"time"
)

// Return currently active sessions
func (r *Repository) GetSessions() ([]model.Session, error) {
	rows, err := r.Base.DB.Query("SELECT email, server_group, expiry FROM sessions")
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
	tx, err := r.Base.DB.Begin()
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

	sessionExpiry, err := util.GetExpiryFromDuration(0, session.Duration)
	if err != nil {
		tx.Rollback()
		return model.Session{}, err
	}

	// Convert epoch to time and add to struct
	session.Expiry = time.Unix(sessionExpiry, 0).UTC()

	if _, err = tx.Exec("INSERT INTO sessions (token, email, server_group, expiry, to_notify, warning_notified, on_notified) VALUES ($1, $2, $3, $4, $5, $6, $7)", session.Token, session.Email, session.ServerGroup, sessionExpiry, 1, 0, 0); err != nil {
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
	newExpiry, err := util.GetExpiryFromDuration(0, session.Duration)
	if err != nil {
		return false, model.Session{}, err
	}

	result, err := r.Base.DB.Exec("UPDATE sessions SET expiry = $1, warning_notified = $2 WHERE token = $3 AND expiry > $4", newExpiry, 0, session.Token, time.Now().Unix())
	if err != nil {
		return false, model.Session{}, err
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

	// Convert epoch to time and add to struct
	session.Expiry = time.Unix(newExpiry, 0).UTC()

	session.Message = "Successfully updated session"
	return true, session, nil
}

// Set servers next_state off and mark session for cleanup
func (r *Repository) EndSession(serverGroup string) error {
	tx, err := r.Base.DB.Begin()
	if err != nil {
		return err
	}

	// Set server next state
	if _, err = tx.Exec("UPDATE servers SET next_state = $1, time_last_off = $2 WHERE server_group = $3", "off", time.Now().Unix(), serverGroup); err != nil {
		tx.Rollback()
		return err
	}

	// Set cleanup flag on session
	if _, err = tx.Exec("UPDATE sessions SET to_cleanup = $1 WHERE server_group = $2", 1, serverGroup); err != nil {
		tx.Rollback()
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

// Cleanup worker task
func (r *Repository) CleanupSessions(sessionsForCleanup []model.Session) {
	for _, session := range sessionsForCleanup {
		tx, err := r.Base.DB.Begin()
		if err != nil {
			r.Base.Logger.Error("Failed to begin transaction", "email", session.Email, "server_group", session.ServerGroup, "error", err)
			continue
		}

		// Delete session
		_, err = tx.Exec("DELETE from sessions where server_group = $1", session.ServerGroup)
		if err != nil {
			tx.Rollback()
			r.Base.Logger.Error("Failed to cleanup expired session", "email", session.Email, "server_group", session.ServerGroup, "error", err)
			continue
		}

		// Null next_state for all servers in group
		r.Base.Logger.Debug(session.ServerGroup)
		_, err = tx.Exec("UPDATE servers SET next_state = NULL where server_group = $1", session.ServerGroup)
		if err != nil {
			tx.Rollback()
			r.Base.Logger.Error("Failed to null server next_state", "server_group", session.ServerGroup, "error", err)
			continue
		}

		if err = tx.Commit(); err != nil {
			r.Base.Logger.Error("Failed to commit transaction", "server_group", session.ServerGroup, "error", err)
			continue
		}

		r.Base.Logger.Info("Session ended normally", "email", session.Email, "server_group", session.ServerGroup)
	}
}

// Find sessions where all relevant servers are in requested state (on or off)
func (r *Repository) FindSessionsForAction(toCleanup int, onNotified int, serverState string) ([]model.Session, error) {
	query := `SELECT s.email, s.server_group, s.expiry
			FROM sessions s
			WHERE s.to_cleanup = $1 AND s.on_notified = $2
			AND NOT EXISTS (
			SELECT 1
			FROM servers srv
			WHERE srv.server_group = s.server_group
			AND (srv.state != $3 OR srv.next_state != $3))`

	rows, err := r.Base.DB.Query(query, toCleanup, onNotified, serverState)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	sessionsForAction := []model.Session{}

	for rows.Next() {
		var email string
		var serverGroup string
		var expiryInt int64

		if err = rows.Scan(&email, &serverGroup, &expiryInt); err != nil {
			return nil, err
		}

		s := model.Session{
			Email:       email,
			ServerGroup: serverGroup,
			Expiry:      time.Unix(expiryInt, 0).UTC(),
		}

		sessionsForAction = append(sessionsForAction, s)
	}

	return sessionsForAction, nil
}

func (r *Repository) SetWarningNotifiedFlag(value int, serverGroup string) error {
	_, err := r.Base.DB.Exec("UPDATE sessions SET warning_notified = $1 WHERE server_group = $2", value, serverGroup)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) SetOnNotifiedFlag(value int, serverGroup string) error {
	_, err := r.Base.DB.Exec("UPDATE sessions SET on_notified = $1 WHERE server_group = $2", value, serverGroup)
	if err != nil {
		return err
	}

	return nil
}
