package user

import (
	"ez2boot/internal/model"
	"fmt"
)

// Login UI user
func (r *Repository) saveUserSession(tokenHash string, expiry int64, userID int64) error {
	// Write hash, expiry and user ID
	if _, err := r.Base.DB.Exec("INSERT INTO user_sessions (token_hash, session_expiry, user_id) VALUES ($1, $2, $3)", tokenHash, expiry, userID); err != nil {
		return err
	}

	return nil
}

// Create new user
func (r *Repository) createUser(username string, passwordHash string) error {
	if _, err := r.Base.DB.Exec("INSERT INTO users (username, password_hash, is_active) VALUES ($1, $2, $3)", username, passwordHash, 1); err != nil {
		return err
	}
	return nil
}

// Find password hash by username
func (r *Repository) findUserIDHashByUsername(username string) (int64, string, error) {
	var passwordHash string
	var id int64
	err := r.Base.DB.QueryRow("SELECT id, password_hash FROM users WHERE username = $1", username).Scan(&id, &passwordHash)
	if err != nil {
		return 0, "", err
	}
	return id, passwordHash, nil
}

// Change password for username
func (r *Repository) changePassword(username string, newHash string) error {
	result, err := r.Base.DB.Exec("UPDATE users SET password_hash = $1 WHERE username = $2", newHash, username)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return fmt.Errorf("Password was not updated for user: %s", username)
	}

	return nil
}

// Get user session info by session token
func (r *Repository) findUserInfoByToken(token string) (model.UserSession, error) {
	query := `SELECT user_sessions.session_expiry, user_sessions.user_id, users.username
        	FROM user_sessions
        	JOIN users ON user_sessions.user_id = users.id
        	WHERE user_sessions.token_hash = $1`

	var u model.UserSession
	err := r.Base.DB.QueryRow(query, token).Scan(&u.SessionExpiry, &u.UserID, &u.Username)
	if err != nil {
		r.Logger.Debug("Error", "error", err)
		return model.UserSession{}, err
	}
	r.Logger.Debug("Comparing", "token", token)
	return u, nil
}
