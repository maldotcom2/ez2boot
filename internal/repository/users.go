package repository

// Create new user
func (r *Repository) CreateUser(username string, passwordHash string) error {
	if _, err := r.DB.Exec("INSERT INTO users (username, password_hash, is_active) VALUES ($1, $2, $3)", username, passwordHash, 1); err != nil {
		return err
	}
	return nil
}

func (r *Repository) FindHash(username string) (string, error) {
	var passwordHash string
	err := r.DB.QueryRow("SELECT password_hash FROM users WHERE username = $1", username).Scan(&passwordHash)
	if err != nil {
		return "", err
	}
	return passwordHash, nil
}
