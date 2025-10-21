package repository

import "fmt"

// Create new user
func (r *Repository) CreateUser(username string, passwordHash string) error {
	if _, err := r.DB.Exec("INSERT INTO users (username, password_hash, is_active) VALUES ($1, $2, $3)", username, passwordHash, 1); err != nil {
		return err
	}
	return nil
}

func (r *Repository) FindHashByUsername(username string) (string, error) {
	var passwordHash string
	err := r.DB.QueryRow("SELECT password_hash FROM users WHERE username = $1", username).Scan(&passwordHash)
	if err != nil {
		return "", err
	}
	return passwordHash, nil
}

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
