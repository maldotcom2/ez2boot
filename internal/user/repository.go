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
func (r *Repository) createUser(email string, passwordHash string) error {
	if _, err := r.Base.DB.Exec("INSERT INTO users (email, password_hash, is_active) VALUES ($1, $2, $3)", email, passwordHash, 1); err != nil {
		return err
	}
	return nil
}

// Find password hash by email
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

// Get user session info by session token
func (r *Repository) findUserInfoByToken(token string) (model.UserSession, error) {
	query := `SELECT user_sessions.session_expiry, user_sessions.user_id, users.email
        	FROM user_sessions
        	JOIN users ON user_sessions.user_id = users.id
        	WHERE user_sessions.token_hash = $1`

	var u model.UserSession
	err := r.Base.DB.QueryRow(query, token).Scan(&u.SessionExpiry, &u.UserID, &u.Email)
	if err != nil {
		r.Logger.Debug("Error", "error", err)
		return model.UserSession{}, err
	}
	r.Logger.Debug("Comparing", "token", token)
	return u, nil
}
