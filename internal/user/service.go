package user

import (
	"database/sql"
	"errors"
	"ez2boot/internal/shared"
	"ez2boot/internal/util"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/alexedwards/argon2id"
)

func (s *Service) loginUser(u UserLogin) (string, error) {
	// Authenticate
	userID, ok, err := s.AuthenticateUser(u.Email, u.Password)
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

	hash := util.HashToken(str)

	sessionExpiry := time.Now().Add(s.Config.UserSessionDuration).Unix()

	// Store it
	if err = s.Repo.createUserSession(hash, sessionExpiry, userID); err != nil {
		return "", err
	}

	return str, nil
}

func (s *Service) logoutUser(token string) error {
	hash := util.HashToken(token)

	if err := s.Repo.deleteUserSession(hash); err != nil {
		return err
	}

	return nil
}

func (s *Service) createUser(u UserLogin) error {
	// Validate password requirements
	if err := validatePassword(u.Email, u.Password); err != nil {
		return err
	}

	// Hash password here
	passwordHash, err := util.HashPassword(u.Password)
	if err != nil {
		return err
	}

	if err := s.Repo.createUser(u.Email, passwordHash); err != nil {
		return err
	}

	return nil
}

// Change a password for authenticated user
func (s *Service) changePassword(req ChangePasswordRequest) error {
	// Get email of authenticated user
	email, err := s.FindEmailFromUserID(req.UserID)
	if err != nil {
		return err
	}

	req.Email = email

	if err = s.validateChangePassword(req); err != nil {
		return err
	}

	// Check current password
	_, isCurrentPassword, err := s.AuthenticateUser(email, req.OldPassword)
	if err != nil {
		return err
	}

	if !isCurrentPassword {
		return shared.ErrAuthenticationFailed
	}

	//Validate complexity
	if err := validatePassword(email, req.NewPassword); err != nil {
		return fmt.Errorf("%w: %v", shared.ErrInvalidPassword, err)
	}

	// Hash new password and change
	newHash, err := util.HashPassword(req.NewPassword)
	if err != nil {
		return err
	}

	if err = s.Repo.changePassword(req.Email, newHash); err != nil {
		return err
	}

	return nil
}

func validatePassword(email string, password string) error {
	length := utf8.RuneCountInString(password)

	if length < 14 {
		return errors.New("Password must be 14 characters or more")
	}

	if strings.Contains(password, email) {
		return errors.New("Password cannot contain the email")
	}

	if strings.Contains(password, email) {
		return errors.New("email cannot contain the password")
	}

	return nil
}

// Authenticate user and return userID and match bool
func (s *Service) AuthenticateUser(email string, password string) (int64, bool, error) {
	id, hash, err := s.Repo.findUserIDHashByEmail(email)
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

func (s *Service) GetSessionStatus(token string) (UserSession, error) {
	// Hash token from cookie
	hash := util.HashToken(token)

	u, err := s.Repo.findSessionStatus(hash)
	if err != nil {
		if err == sql.ErrNoRows {
			return UserSession{}, shared.ErrSessionNotFound
		}
		return UserSession{}, err
	}

	if u.SessionExpiry < time.Now().Unix() {
		return UserSession{}, shared.ErrSessionExpired
	}

	return u, nil
}

func (s *Service) GetUserAuthorisation(email string) (UserAuth, error) {
	u, err := s.Repo.findUserAuthorisation(email)
	if err != nil {
		return UserAuth{}, nil
	}

	return u, nil
}

func (s *Service) CleanupExpiredSessions() (sql.Result, error) {
	now := time.Now().Unix()

	result, err := s.Repo.deleteExpiredSessions(now)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *Service) FindEmailFromUserID(userID int64) (string, error) {
	email, err := s.Repo.findEmailFromUserID(userID)
	if err != nil {
		return "", err
	}

	return email, nil
}
