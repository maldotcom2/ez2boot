package users

import (
	"ez2boot/internal/model"
	"ez2boot/internal/repository"

	"github.com/alexedwards/argon2id"
)

func ValidateAndCreateUser(repo *repository.Repository, user model.User) error {
	// Validate password requirements
	if err := validatePassword(user.Username, user.Password); err != nil {
		return err
	}

	// Hash password here
	passwordHash, err := hashString(user.Password)
	if err != nil {
		return err
	}

	if err := repo.CreateUser(user.Username, passwordHash); err != nil {
		return err
	}

	return nil
}

func hashString(secret string) (string, error) {

	params := &argon2id.Params{
		Memory:      128 * 1024,
		Iterations:  4,
		Parallelism: 1,
		SaltLength:  16,
		KeyLength:   32,
	}

	hash, err := argon2id.CreateHash(secret, params)
	if err != nil {
		return "", err
	}

	return hash, nil
}
