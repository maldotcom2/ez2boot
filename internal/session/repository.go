package session

import (
	"database/sql"
	"ez2boot/internal/util"
	"fmt"
	"time"
)

// Return currently active sessions
func (r *Repository) getServerSessions() ([]ServerSession, error) {
	rows, err := r.Base.DB.Query("SELECT user_id, server_group, expiry FROM server_sessions")
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	sessions := []ServerSession{}

	for rows.Next() {
		var s ServerSession
		var expiryInt int64
		err = rows.Scan(&s.UserID, &s.ServerGroup, &expiryInt)
		if err != nil {
			return nil, err
		}

		// Convert epoch to time
		s.Expiry = time.Unix(expiryInt, 0).UTC()

		sessions = append(sessions, s)
	}

	return sessions, nil
}

func (r *Repository) getAgingServerSessions() ([]ServerSession, error) {
	now := time.Now().UTC()
	threshold := now.Add(15 * time.Minute)

	rows, err := r.Base.DB.Query("SELECT user_id, server_group FROM server_sessions WHERE warning_notified = $1 AND expiry BETWEEN $2 AND $3", 0, now.Unix(), threshold.Unix())

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	sessions := []ServerSession{}

	for rows.Next() {
		var s ServerSession
		err = rows.Scan(&s.UserID, &s.ServerGroup)
		if err != nil {
			return nil, err
		}

		sessions = append(sessions, s)
	}

	return sessions, nil
}

func (r *Repository) getExpiredServerSessions() ([]ServerSession, error) {
	rows, err := r.Base.DB.Query("SELECT user_id, server_group FROM server_sessions WHERE expiry < $1 AND to_cleanup = $2", time.Now().Unix(), 0)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	sessions := []ServerSession{}

	for rows.Next() {
		var s ServerSession
		err = rows.Scan(&s.UserID, &s.ServerGroup)
		if err != nil {
			return nil, err
		}

		sessions = append(sessions, s)
	}

	return sessions, nil
}

// Create a new session
func (r *Repository) newServerSession(session ServerSession) (ServerSession, error) {
	tx, err := r.Base.DB.Begin()
	if err != nil {
		return ServerSession{}, err
	}

	// Set server table for state worker
	result, err := tx.Exec("UPDATE servers SET next_state = $1, time_last_on = $2, last_user_id = $3 WHERE server_group = $4", "on", time.Now().Unix(), session.UserID, session.ServerGroup)
	if err != nil {
		tx.Rollback()
		return ServerSession{}, err
	}

	// Impact check
	rows, err := result.RowsAffected()
	if err != nil {
		tx.Rollback()
		return ServerSession{}, err
	}

	if rows == 0 {
		tx.Rollback()
		return ServerSession{}, fmt.Errorf("No servers found for server_group: %s", session.ServerGroup)
	}

	sessionExpiry, err := util.GetExpiryFromDuration(0, session.Duration)
	if err != nil {
		tx.Rollback()
		return ServerSession{}, err
	}

	// Convert epoch to time and add to struct
	session.Expiry = time.Unix(sessionExpiry, 0).UTC()

	if _, err = tx.Exec("INSERT INTO server_sessions (user_id, server_group, expiry, warning_notified, on_notified) VALUES ($1, $2, $3, $4, $5)", session.UserID, session.ServerGroup, sessionExpiry, 0, 0); err != nil {
		tx.Rollback()
		return ServerSession{}, err
	}

	if err = tx.Commit(); err != nil {
		return ServerSession{}, err
	}

	return session, nil
}

// Update existing session
func (r *Repository) updateServerSession(session ServerSession) (bool, ServerSession, error) {
	newExpiry, err := util.GetExpiryFromDuration(0, session.Duration)
	if err != nil {
		return false, ServerSession{}, err
	}

	result, err := r.Base.DB.Exec("UPDATE server_sessions SET expiry = $1, warning_notified = $2 WHERE expiry > $3", newExpiry, 0, time.Now().Unix())
	if err != nil {
		return false, ServerSession{}, err
	}

	// Impact check
	rows, err := result.RowsAffected()
	if err != nil {
		return false, ServerSession{}, err
	}

	if rows == 0 {
		return false, ServerSession{}, nil
	}

	// Convert epoch to time and add to struct
	session.Expiry = time.Unix(newExpiry, 0).UTC()

	return true, session, nil
}

// Set servers next_state off and mark session for cleanup
func (r *Repository) endServerSession(serverGroup string) error {
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
	if _, err = tx.Exec("UPDATE server_sessions SET to_cleanup = $1 WHERE server_group = $2", 1, serverGroup); err != nil {
		tx.Rollback()
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

// Cleanup worker task
func (r *Repository) cleanupServerSessions(sessionsForCleanup []ServerSession) {
	for _, session := range sessionsForCleanup {
		r.Base.Logger.Debug("Cleanup Session", "session", session.Email)
		tx, err := r.Base.DB.Begin()
		if err != nil {
			r.Base.Logger.Error("Failed to begin transaction", "email", session.Email, "server_group", session.ServerGroup, "error", err)
			continue
		}

		// Delete session
		_, err = tx.Exec("DELETE from server_sessions where server_group = $1", session.ServerGroup)
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

// Find session states no longer matching server state
func (r *Repository) findPendingOffServerSessions() ([]ServerSession, error) {
	query := `SELECT s.id, u.email, s.server_group, s.expiry
			FROM server_sessions s
			JOIN users u ON s.user_id = u.id
			WHERE s.to_cleanup = $1 AND s.off_notified = $2
			AND NOT EXISTS (
			SELECT 1
			FROM servers srv
			WHERE srv.server_group = s.server_group
			AND (srv.state != $4 OR srv.next_state != $3))`

	rows, err := r.Base.DB.Query(query, 1, 1, "off")
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	sessionsForAction := []ServerSession{}

	for rows.Next() {
		var userID int64
		var email string
		var serverGroup string
		var expiryInt int64

		if err = rows.Scan(&userID, &email, &serverGroup, &expiryInt); err != nil {
			return nil, err
		}

		s := ServerSession{
			UserID:      userID,
			Email:       email,
			ServerGroup: serverGroup,
			Expiry:      time.Unix(expiryInt, 0).UTC(),
		}

		sessionsForAction = append(sessionsForAction, s)
	}

	return sessionsForAction, nil
}

func (r *Repository) findPendingOnServerSessions() ([]ServerSession, error) {
	query := `SELECT s.id, u.email, s.server_group, s.expiry
			FROM server_sessions s
			JOIN users u ON s.user_id = u.id
			WHERE s.to_cleanup = $1 AND s.on_notified = $2
			AND NOT EXISTS (
			SELECT 1
			FROM servers srv
			WHERE srv.server_group = s.server_group
			AND (srv.state != $3 OR srv.next_state != $3))`

	rows, err := r.Base.DB.Query(query, 0, 0, "on")
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	sessionsForAction := []ServerSession{}

	for rows.Next() {
		var userID int64
		var email string
		var serverGroup string
		var expiryInt int64

		if err = rows.Scan(&userID, &email, &serverGroup, &expiryInt); err != nil {
			return nil, err
		}

		s := ServerSession{
			UserID:      userID,
			Email:       email,
			ServerGroup: serverGroup,
			Expiry:      time.Unix(expiryInt, 0).UTC(),
		}

		sessionsForAction = append(sessionsForAction, s)
	}

	return sessionsForAction, nil
}

// Set warning notified flag - called with notification queuing so runs as a transaction
func (r *Repository) setWarningNotifiedFlag(tx *sql.Tx, flagValue int, serverGroup string) error {
	_, err := tx.Exec("UPDATE server_sessions SET warning_notified = $1 WHERE server_group = $2", flagValue, serverGroup)
	if err != nil {
		return err
	}

	return nil
}

// Set on notified flag - called with notification queuing so runs as a transaction
func (r *Repository) setOnNotifiedFlag(tx *sql.Tx, flagValue int, serverGroup string) error {
	_, err := tx.Exec("UPDATE server_sessions SET on_notified = $1 WHERE server_group = $2", flagValue, serverGroup)
	if err != nil {
		return err
	}

	return nil
}
