package user

import (
	"database/sql"
	"errors"
	"ez2boot/internal/model"
	"ez2boot/internal/shared"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/alexedwards/argon2id"
)

func (s *Service) validateAndCreateUser(u model.User) error {
	// Validate password requirements
	if err := validatePassword(u.Username, u.Password); err != nil {
		return err
	}

	// Hash password here
	passwordHash, err := hashString(u.Password)
	if err != nil {
		return err
	}

	if err := s.Repo.createUser(u.Username, passwordHash); err != nil {
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

// Change a password for authenticated user
func (s *Service) changePasswordByUser(req model.ChangePasswordRequest) error {
	// Check current password
	isCurrentPassword, err := s.ComparePassword(req.Username, req.OldPassword)
	if err != nil {
		return err
	}

	if !isCurrentPassword {
		return shared.ErrAuthenticationFailed
	}

	//Validate complexity
	if err := validatePassword(req.Username, req.NewPassword); err != nil {
		return fmt.Errorf("%w: %v", shared.ErrInvalidPassword, err)
	}

	// Hash new password and change
	newHash, err := hashString(req.NewPassword)
	if err != nil {
		return err
	}

	if err = s.Repo.changePassword(req.Username, newHash); err != nil {
		return err
	}

	return nil
}

func validatePassword(username string, password string) error {
	length := utf8.RuneCountInString(password)

	if length < 14 {
		return errors.New("Password must be 14 characters or more")
	}

	if strings.Contains(password, username) {
		return errors.New("Password cannot contain the username")
	}

	if strings.Contains(password, username) {
		return errors.New("Username cannot contain the password")
	}

	return nil
}

func (s *Service) ComparePassword(username string, password string) (bool, error) {
	hash, err := s.Repo.findHashByUsername(username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, shared.ErrUserNotFound
		}
		return false, err
	}

	match, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		return false, err
	}

	return match, nil
}

func (s *Service) GetSessionInfo(token string) (model.UserSession, error) {
	u, err := s.Repo.findUserInfoByToken(token)
	if err != nil {
		if err == sql.ErrNoRows {
			return model.UserSession{}, shared.ErrSessionNotFound
		}
		return model.UserSession{}, err
	}

	if u.SessionExpiry < time.Now().Unix() {
		return model.UserSession{}, shared.ErrSessionExpired
	}

	return u, nil
}

func (s *Service) GetBasicAuthInfo(username string) (int64, error) {
	userID, err := s.Repo.findBasicAuthUserID(username)
	if err != nil {
		return 0, err
	}

	return userID, nil
}
