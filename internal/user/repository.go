package user

import (
	"database/sql"
	"ez2boot/internal/shared"
	"time"

	"github.com/mattn/go-sqlite3"
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

func (r *Repository) getUserAuthorisation(userID int64) (UserAuthResponse, error) {
	query := `SELECT id, email, is_active, is_admin, api_enabled, ui_enabled, identity_provider, mfa_confirmed FROM users WHERE id = $1`

	var u UserAuthResponse
	if err := r.Base.DB.QueryRow(query, userID).Scan(&u.UserID, &u.Email, &u.IsActive, &u.IsAdmin, &u.APIEnabled, &u.UIEnabled, &u.IdentityProvider, &u.MFAConfirmed); err != nil {
		return UserAuthResponse{}, err
	}

	return u, nil
}

// Update from admin panel
func (r *Repository) updateUserAuthorisation(tx *sql.Tx, u UpdateUserRequest) error {
	if _, err := tx.Exec("UPDATE users SET is_active = $1, is_admin = $2, api_enabled = $3, ui_enabled = $4 WHERE id = $5", u.IsActive, u.IsAdmin, u.APIEnabled, u.UIEnabled, u.UserID); err != nil {
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

// Create new user
func (r *Repository) createUser(u CreateUser) (int64, error) {
	var userID int64
	if err := r.Base.DB.QueryRow("INSERT INTO users (email, password_hash, is_active, is_admin, api_enabled, ui_enabled, identity_provider) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id", u.Email, u.PasswordHash, u.IsActive, u.IsAdmin, u.APIEnabled, u.UIEnabled, "local").Scan(&userID); err != nil {
		if sqliteErr, ok := err.(sqlite3.Error); ok {
			if sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
				return 0, shared.ErrUserAlreadyExists
			}
		}
		return 0, err
	}
	return userID, nil
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
func (r *Repository) changePassword(userID int64, newHash string) error {
	resReset, err := r.Base.DB.Exec("UPDATE users SET password_hash = $1 WHERE id = $2", newHash, userID)
	if err != nil {
		return err
	}

	rowsReset, err := resReset.RowsAffected()
	if err != nil {
		return err
	}

	if rowsReset == 0 {
		return shared.ErrNoRowsUpdated
	}

	// Clear existing sessions
	resPurge, err := r.Base.DB.Exec("DELETE FROM user_sessions WHERE user_id = $1", userID)
	if err != nil {
		return err
	}

	rowsPurge, err := resPurge.RowsAffected()
	if err != nil {
		return err
	}

	if rowsPurge == 0 {
		return shared.ErrNoRowsDeleted
	}

	return nil
}

// Get user session info by session token hash
func (r *Repository) getSessionStatus(hash string) (UserSessionResponse, error) {
	query := `SELECT us.session_expiry, u.id, u.email
        	FROM user_sessions AS us
        	JOIN users u ON us.user_id = u.id
        	WHERE us.token_hash = $1`

	var u UserSessionResponse
	err := r.Base.DB.QueryRow(query, hash).Scan(&u.SessionExpiry, &u.UserID, &u.Email)
	if err != nil {
		return UserSessionResponse{}, err
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

func (r *Repository) setMFASecret(secret *string, userID int64) (int64, error) {
	result, err := r.Base.DB.Exec("UPDATE users SET mfa_secret = $1, mfa_confirmed = 0 WHERE id = $2", secret, userID)
	if err != nil {
		return 0, err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return rows, nil
}

func (r *Repository) confirmMFA(userID int64) (int64, error) {
	result, err := r.Base.DB.Exec("UPDATE users SET mfa_confirmed = 1 WHERE mfa_confirmed = 0 AND id = $1", userID)
	if err != nil {
		return 0, err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return rows, nil
}

func (r *Repository) getMFASecret(userID int64) (*string, error) {
	var secret *string
	if err := r.Base.DB.QueryRow("SELECT mfa_secret FROM users WHERE id = $1", userID).Scan(&secret); err != nil {
		return nil, err
	}

	return secret, nil
}
