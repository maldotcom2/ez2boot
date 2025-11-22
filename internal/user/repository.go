package user

import (
	"ez2boot/internal/shared"
	"fmt"
	"time"

	"github.com/mattn/go-sqlite3"
)

func (r *Repository) getUsers() ([]User, error) {
	rows, err := r.Base.DB.Query("SELECT id, email, is_active, is_admin, api_enabled, ui_enabled, last_login FROM users")
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	users := []User{}

	for rows.Next() {
		var u User
		if err = rows.Scan(&u.UserID, &u.Email, &u.IsActive, &u.IsAdmin, &u.APIEnabled, &u.UIEnabled, &u.LastLogin); err != nil {
			return nil, err
		}

		users = append(users, u)
	}

	return users, nil
}

// Bulk update from admin panel
func (r *Repository) updateUserAuthorisation(users []UpdateUserRequest) error {
	tx, err := r.Base.DB.Begin()
	if err != nil {
		return err
	}

	for _, u := range users {
		if _, err := tx.Exec("UPDATE users SET is_active = $1, is_admin = $2, api_enabled = $3, ui_enabled = $4 WHERE id = $5", u.IsActive, u.IsAdmin, u.APIEnabled, u.UIEnabled, u.UserID); err != nil {
			tx.Rollback()
			return err
		}
	}

	tx.Commit()

	return nil
}

// Login UI user
func (r *Repository) createUserSession(tokenHash string, expiry int64, userID int64) error {
	// Write hash, expiry and user ID
	if _, err := r.Base.DB.Exec("INSERT INTO user_sessions (token_hash, session_expiry, user_id) VALUES ($1, $2, $3)", tokenHash, expiry, userID); err != nil {
		return err
	}

	return nil
}

func (r *Repository) updateLastLogin(userID int64) error {
	if _, err := r.Base.DB.Exec("UPDATE users SET last_login = $1 WHERE id = $2", time.Now().Unix(), userID); err != nil {
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

// Database contains users true or false
func (r *Repository) hasUsers() (bool, error) {
	var count int64
	if err := r.Base.DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&count); err != nil {
		return false, err
	}

	return count > 0, nil
}

// Create new user
func (r *Repository) createUser(u CreateUser) error {
	if _, err := r.Base.DB.Exec("INSERT INTO users (email, password_hash, is_active, is_admin, api_enabled, ui_enabled) VALUES ($1, $2, $3, $4, $5, $6)", u.Email, u.PasswordHash, u.IsActive, u.IsAdmin, u.APIEnabled, u.UIEnabled); err != nil {
		if sqliteErr, ok := err.(sqlite3.Error); ok {
			if sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
				return shared.ErrUserAlreadyExists
			}
		}
		return err
	}
	return nil
}

func (r *Repository) deleteUser(userID int64) error {
	if _, err := r.Base.DB.Exec("DELETE FROM users WHERE id = $1", userID); err != nil {
		return err
	}

	return nil
}

// Find password hash and ID by email
func (r *Repository) getUserIDHashByEmail(email string) (int64, string, error) {
	var passwordHash string
	var id int64
	err := r.Base.DB.QueryRow("SELECT id, password_hash FROM users WHERE email = $1", email).Scan(&id, &passwordHash)
	if err != nil {
		return 0, "", err
	}
	return id, passwordHash, nil
}

func (r *Repository) getEmailFromUserID(userID int64) (string, error) {
	var email string
	if err := r.Base.DB.QueryRow("SELECT email from users WHERE id = $1", userID).Scan(&email); err != nil {
		return "", err
	}

	return email, nil
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
func (r *Repository) getSessionStatus(hash string) (UserSession, error) {
	query := `SELECT us.session_expiry, u.id, u.email
        	FROM user_sessions AS us
        	JOIN users u ON us.user_id = u.id
        	WHERE us.token_hash = $1`

	var u UserSession
	err := r.Base.DB.QueryRow(query, hash).Scan(&u.SessionExpiry, &u.UserID, &u.Email)
	if err != nil {
		return UserSession{}, err
	}

	return u, nil
}

// Delete expired sessons and return rows affected
func (r *Repository) deleteExpiredUserSessions() (int64, error) {
	result, err := r.Base.DB.Exec("DELETE FROM user_sessions WHERE session_expiry < $1", time.Now().Unix())
	if err != nil {
		return 0, err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return rows, nil
}

func (r *Repository) getUserAuthorisation(userID int64) (UserAuthRequest, error) {
	query := `SELECT id, email, is_active, is_admin, api_enabled, ui_enabled FROM users WHERE id = $1`

	var u UserAuthRequest
	if err := r.Base.DB.QueryRow(query, userID).Scan(&u.UserID, &u.Email, &u.IsActive, &u.IsAdmin, &u.APIEnabled, &u.UIEnabled); err != nil {
		return UserAuthRequest{}, err
	}

	return u, nil
}
