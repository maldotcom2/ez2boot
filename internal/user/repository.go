package user

import (
	"database/sql"
	"fmt"
)

// Login UI user
func (r *Repository) createUserSession(tokenHash string, expiry int64, userID int64) error {
	// Write hash, expiry and user ID
	if _, err := r.Base.DB.Exec("INSERT INTO user_sessions (token_hash, session_expiry, user_id) VALUES ($1, $2, $3)", tokenHash, expiry, userID); err != nil {
		return err
	}

	return nil
}

// Delete provided session
func (r *Repository) deleteUserSession(tokenHash string) error {
	_, err := r.Base.DB.Exec("DELETE FROM user_sessions WHERE token_hash = $1", tokenHash)
	if err != nil {
		return err
	}
	// No check for 0 rows because logout is a protected route, auth is implicit
	return nil
}

// Create new user
func (r *Repository) createUser(email string, passwordHash string) error {
	if _, err := r.Base.DB.Exec("INSERT INTO users (email, password_hash, is_active) VALUES ($1, $2, $3)", email, passwordHash, 1); err != nil {
		return err
	}
	return nil
}

// Find password hash and ID by email
func (r *Repository) findUserIDHashByEmail(email string) (int64, string, error) {
	var passwordHash string
	var id int64
	err := r.Base.DB.QueryRow("SELECT id, password_hash FROM users WHERE email = $1", email).Scan(&id, &passwordHash)
	if err != nil {
		return 0, "", err
	}
	return id, passwordHash, nil
}

// Change password for user
func (r *Repository) changePassword(email string, newHash string) error {
	result, err := r.Base.DB.Exec("UPDATE users SET password_hash = $1 WHERE email = $2", newHash, email)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return fmt.Errorf("Password was not updated for user: %s", email)
	}

	return nil
}

// Get user session info by session token hash
func (r *Repository) findSessionStatus(hash string) (UserSession, error) {
	query := `SELECT user_sessions.session_expiry, users.email
        	FROM user_sessions
        	JOIN users ON user_sessions.user_id = users.id
        	WHERE user_sessions.token_hash = $1`

	var u UserSession
	err := r.Base.DB.QueryRow(query, hash).Scan(&u.SessionExpiry, &u.Email)
	if err != nil {
		return UserSession{}, err
	}

	return u, nil
}

func (r *Repository) deleteExpiredSessions(now int64) (sql.Result, error) {
	result, err := r.Base.DB.Exec("DELETE FROM user_sessions WHERE session_expiry < $1", now)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *Repository) findUserAuthorisation(email string) (UserAuth, error) {
	query := `SELECT id, email, is_active, is_admin, api_enabled, ui_enabled FROM users WHERE email = $1`

	var u UserAuth
	if err := r.Base.DB.QueryRow(query, email).Scan(&u.UserID, &u.Email, &u.IsActive, &u.IsAdmin, &u.APIEnabled, &u.UIEnabled); err != nil {
		return UserAuth{}, err
	}

	return u, nil
}
