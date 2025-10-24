package repository

import (
	"ez2boot/internal/model"
	"fmt"
)

// Create new user
func (r *Repository) CreateUser(username string, passwordHash string) error {
	if _, err := r.DB.Exec("INSERT INTO users (username, password_hash, is_active) VALUES ($1, $2, $3)", username, passwordHash, 1); err != nil {
		return err
	}
	return nil
}

// Find password hash by username
func (r *Repository) FindHashByUsername(username string) (string, error) {
	var passwordHash string
	err := r.DB.QueryRow("SELECT password_hash FROM users WHERE username = $1", username).Scan(&passwordHash)
	if err != nil {
		return "", err
	}
	return passwordHash, nil
}

// Change password for username
func (r *Repository) ChangePassword(username string, newHash string) error {
	result, err := r.DB.Exec("UPDATE users SET password_hash = $1 WHERE username = $2", newHash, username)
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

// Get userID by name for basic auth
func (r *Repository) FindBasicAuthUserID(username string) (int64, error) {
	var id int64
	err := r.DB.QueryRow("SELECT id FROM users WHERE username = $1", username).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

// Get user session info by session token
func (r *Repository) FindUserInfoByToken(token string) (model.UserSession, error) {
	query := `SELECT user_sessions.session_expiry, user_sessions.user_id, users.username
        	FROM user_sessions
        	JOIN users ON user_sessions.user_id = users.id
        	WHERE user_sessions.session_id = $1`

	var u model.UserSession
	err := r.DB.QueryRow(query, token).Scan(&u.SessionExpiry, &u.UserID, &u.Username)
	if err != nil {
		return model.UserSession{}, err
	}

	return u, nil
}
