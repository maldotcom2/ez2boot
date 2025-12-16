package app

import "ez2boot/internal/db"

func hasUsers(repo *db.Repository) (bool, error) {
	var count int64
	if err := repo.DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&count); err != nil {
		return false, err
	}

	return count > 0, nil
}
