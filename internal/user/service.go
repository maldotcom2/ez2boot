package user

import (
	"database/sql"
	"errors"
	"ez2boot/internal/shared"
	"ez2boot/internal/util"
	"fmt"
	"time"

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

	// Create session token
	token, err := util.GenerateRandomString(32)
	if err != nil {
		return "", err
	}

	// Hash it
	hash := util.HashToken(token)

	// Set expiry
	sessionExpiry := time.Now().Add(s.Config.UserSessionDuration).Unix()

	// Store it
	if err = s.Repo.createUserSession(hash, sessionExpiry, userID); err != nil {
		return "", err
	}

	return token, nil
}

func (s *Service) logoutUser(token string) error {
	// Hash supplied session token
	hash := util.HashToken(token)

	if err := s.Repo.deleteUserSession(hash); err != nil {
		return err
	}

	return nil
}

// Check if any users exist in DB
func (s *Service) HasUsers() (bool, error) {
	hasUsers, err := s.Repo.hasUsers()
	if err != nil {
		return false, err
	}

	return hasUsers, nil
}

// TODO improve validation here
func (s *Service) createUser(req CreateUserRequest) error {
	// Check email
	if err := s.validateEmail(req.Email); err != nil {
		return err
	}

	// Validate password requirements
	if err := validatePassword(req.Email, req.Password); err != nil {
		return err
	}

	// Hash password here
	passwordHash, err := util.HashPassword(req.Password)
	if err != nil {
		return err
	}

	// Don't transport password
	user := CreateUser{
		UserID:       req.UserID,
		Email:        req.Email,
		PasswordHash: passwordHash,
		IsActive:     req.IsActive,
		IsAdmin:      req.IsAdmin,
		APIEnabled:   req.APIEnabled,
		UIEnabled:    req.UIEnabled,
	}

	if err := s.Repo.createUser(user); err != nil {
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

	// Validate request
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

// Authenticate user, return userID for use in context and match bool
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

	userSession, err := s.Repo.findSessionStatus(hash)
	if err != nil {
		if err == sql.ErrNoRows {
			return UserSession{}, shared.ErrSessionNotFound
		}
		return UserSession{}, err
	}

	if userSession.SessionExpiry < time.Now().Unix() {
		return UserSession{}, shared.ErrSessionExpired
	}

	return userSession, nil
}

// Get a user's authorisation, eg admin, API access, etc
func (s *Service) GetUserAuthorisation(email string) (User, error) {
	user, err := s.Repo.findUserAuthorisation(email)
	if err != nil {
		return User{}, nil
	}

	return user, nil
}

func (s *Service) DeleteExpiredUserSessions() (sql.Result, error) { // TODO should a result be returned here?
	now := time.Now().Unix()

	result, err := s.Repo.deleteExpiredUserSessions(now)
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
