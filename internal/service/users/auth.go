package users

import (
	"ez2boot/internal/repository"
	"log/slog"

	"github.com/alexedwards/argon2id"
)

func comparePassword(repo *repository.Repository, username string, password string, logger *slog.Logger) (bool, error) {
	hash, err := repo.FindHashByUsername(username)
	if err != nil {
		return false, err
	}

	match, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		return false, err
	}

	return match, nil
}
