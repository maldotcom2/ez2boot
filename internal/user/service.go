package user

import (
	"database/sql"
	"errors"
	"ez2boot/internal/model"
	"ez2boot/internal/shared"
	"ez2boot/internal/util"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/alexedwards/argon2id"
)

func (s *Service) LoginUser(u model.UserLogin) (string, error) {
	// Authenticate
	userID, ok, err := s.AuthenticateUser(u.Username, u.Password)
	if err != nil {
		return "", err
	}

	if !ok {
		return "", shared.ErrAuthenticationFailed
	}

	// Create token
	str, err := util.GenerateRandomString(32)
	if err != nil {
		return "", err
	}

	hash, err := util.HashString(str)
	if err != nil {
		return "", err
	}

	sessionExpiry := time.Now().Add(s.Config.UserSessionDuration).Unix()

	// Store it
	if err = s.Repo.saveUserSession(hash, sessionExpiry, userID); err != nil {
		return "", err
	}

	return hash, nil
}

func (s *Service) validateAndCreateUser(u model.UserLogin) error {
	// Validate password requirements
	if err := validatePassword(u.Username, u.Password); err != nil {
		return err
	}

	// Hash password here
	passwordHash, err := util.HashString(u.Password)
	if err != nil {
		return err
	}

	if err := s.Repo.createUser(u.Username, passwordHash); err != nil {
		return err
	}

	return nil
}

// Change a password for authenticated user
func (s *Service) changePasswordByUser(req model.ChangePasswordRequest) error {
	// Check current password
	_, isCurrentPassword, err := s.AuthenticateUser(req.Username, req.OldPassword)
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
	newHash, err := util.HashString(req.NewPassword)
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

func (s *Service) AuthenticateUser(username string, password string) (int64, bool, error) {
	id, hash, err := s.Repo.findIDHashByUsername(username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, false, shared.ErrUserNotFound
		}
		return 0, false, err
	}

	match, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		return 0, false, err
	}

	return id, match, nil
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
