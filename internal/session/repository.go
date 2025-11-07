package session

import (
	"database/sql"
	"ez2boot/internal/shared"
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

// Get server sessions which will expire soon and user not yet notified
func (r *Repository) getAgingServerSessions() ([]ServerSession, error) {
	now := time.Now().UTC()
	threshold := now.Add(15 * time.Minute)

	rows, err := r.Base.DB.Query("SELECT user_id, server_group FROM server_sessions WHERE warning_notified = 0 AND expiry BETWEEN $1 AND $2", now.Unix(), threshold.Unix())

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

// Get expired server session which haven't been processed yet
func (r *Repository) getExpiredServerSessions() ([]ServerSession, error) {
	rows, err := r.Base.DB.Query("SELECT user_id, server_group FROM server_sessions WHERE expiry < $1 AND to_cleanup = 0", time.Now().Unix())

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

	// Get expiry
	sessionExpiry, err := util.GetExpiryFromDuration(session.Duration)
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
func (r *Repository) updateServerSession(session ServerSession) (ServerSession, error) {
	newExpiry, err := util.GetExpiryFromDuration(session.Duration)
	if err != nil {
		return ServerSession{}, err
	}

	result, err := r.Base.DB.Exec("UPDATE server_sessions SET expiry = $1, warning_notified = $2 WHERE server_group = $3 AND user_id = $4 AND expiry > $5", newExpiry, 0, session.ServerGroup, session.UserID, time.Now().Unix())
	if err != nil {
		return ServerSession{}, err
	}

	// Impact check
	rows, err := result.RowsAffected()
	if err != nil {
		return ServerSession{}, err
	}

	if rows == 0 {
		return ServerSession{}, shared.ErrNoRowsUpdated
	}

	// Convert epoch to time and add to struct
	session.Expiry = time.Unix(newExpiry, 0).UTC()

	return session, nil
}

// Set servers next_state off and mark session for cleanup
func (r *Repository) endServerSession(tx *sql.Tx, serverGroup string) error {
	// Set server next state
	if _, err := tx.Exec("UPDATE servers SET next_state = $1, time_last_off = $2 WHERE server_group = $3", "off", time.Now().Unix(), serverGroup); err != nil {
		return err
	}

	// Set cleanup flag on session
	if _, err := tx.Exec("UPDATE server_sessions SET to_cleanup = $1 WHERE server_group = $2", 1, serverGroup); err != nil {
		return err
	}

	return nil
}

// Delete the server session and set server next state to null
func (r *Repository) cleanupServerSession(tx *sql.Tx, session ServerSession) error {
	r.Base.Logger.Debug("Cleanup Session", "session", session.Email)

	// Delete session
	if _, err := tx.Exec("DELETE from server_sessions where server_group = $1", session.ServerGroup); err != nil {
		return err
	}

	// Null next_state for all servers in group
	if _, err := tx.Exec("UPDATE servers SET next_state = NULL where server_group = $1", session.ServerGroup); err != nil {
		return err
	}

	return nil
}

// Find sessions which are not marked for cleanup, and haven't been notified on yet
func (r *Repository) getPendingOnServerSessions() ([]ServerSession, error) {
	query := `SELECT u.id, u.email, s.server_group, s.expiry
			FROM server_sessions s
			JOIN users u ON s.user_id = u.id
			WHERE s.to_cleanup = 0 AND s.on_notified = 0
			AND NOT EXISTS (
			SELECT 1
			FROM servers srv
			WHERE srv.server_group = s.server_group
			AND (srv.state != 'on' OR srv.next_state != 'on'))`

	rows, err := r.Base.DB.Query(query)
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

// Find sessions which have been marked for cleanup and not yet notified
func (r *Repository) getTerminatedServerSessions() ([]ServerSession, error) {
	query := `SELECT u.id AS user_id, u.email, s.server_group, s.expiry
			FROM server_sessions s
			JOIN users u ON s.user_id = u.id
			WHERE s.to_cleanup = 1 AND s.off_notified = 0
			AND NOT EXISTS (
			SELECT 1
			FROM servers srv
			WHERE srv.server_group = s.server_group
			AND (srv.state != 'off' OR srv.next_state != 'off'))`

	rows, err := r.Base.DB.Query(query)
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

// Find sessions which are marked for cleanup and user has been notified of servers off state
func (r *Repository) getFinalisedServerSessions() ([]ServerSession, error) {
	query := `SELECT u.id, u.email, s.server_group, s.expiry
			FROM server_sessions s
			JOIN users u ON s.user_id = u.id
			WHERE s.to_cleanup = 1 AND s.off_notified = 1
			AND NOT EXISTS (
			SELECT 1
			FROM servers srv
			WHERE srv.server_group = s.server_group
			AND (srv.state != 'off' OR srv.next_state != 'off'))`

	rows, err := r.Base.DB.Query(query)
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

// Set off notified flag - called with notification queuing so runs as a transaction
func (r *Repository) setOffNotifiedFlag(tx *sql.Tx, flagValue int, serverGroup string) error {
	_, err := tx.Exec("UPDATE server_sessions SET off_notified = $1 WHERE server_group = $2", flagValue, serverGroup)
	if err != nil {
		return err
	}

	return nil
}
