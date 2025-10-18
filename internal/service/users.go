package service

import (
	"errors"
	"ez2boot/internal/model"
	"ez2boot/internal/repository"
	"log/slog"
	"strings"
	"unicode/utf8"

	"github.com/alexedwards/argon2id"
)

func ValidateAndCreateUser(repo *repository.Repository, user model.User) error {
	// Validate password requirements
	isOK, err := ValidatePassword(user.Username, user.Password)
	if !isOK {
		return err
	}

	// Hash password here
	passwordHash, err := HashString(user.Password)
	if err != nil {
		return err
	}

	if err := repo.CreateUser(user.Username, passwordHash); err != nil {
		return err
	}

	return nil
}

func HashString(token string) (string, error) {

	params := &argon2id.Params{
		Memory:      128 * 1024,
		Iterations:  4,
		Parallelism: 1,
		SaltLength:  16,
		KeyLength:   32,
	}

	hash, err := argon2id.CreateHash(token, params)
	if err != nil {
		return "", err
	}

	return hash, nil
}

func ValidatePassword(username string, password string) (bool, error) {
	length := utf8.RuneCountInString(password)

	if length < 14 {
		return false, errors.New("Password must be 14 characters or more")
	}

	if strings.Contains(password, username) {
		return false, errors.New("Password cannot contain the username")
	}

	if strings.Contains(password, username) {
		return false, errors.New("Username cannot contain the password")
	}

	return true, nil
}

func ComparePassword(repo *repository.Repository, username string, password string, logger *slog.Logger) (bool, error) {
	hash, err := repo.FindHash(username)
	if err != nil {
		return false, err
	}

	match, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		return false, err
	}

	return match, nil
}
