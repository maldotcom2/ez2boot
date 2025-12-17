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

// Attempt user login using even-time
func (s *Service) login(u UserLogin) (string, error) {
	if u.Email == "" || u.Password == "" {
		return "", shared.ErrEmailOrPasswordMissing
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

	// Authenticate
	userID, authenticated, err := s.AuthenticateUser(u.Email, u.Password)
	if err != nil {
		return "", err
	}

	if !authenticated {
		return "", shared.ErrAuthenticationFailed
	}

	// Store hash
	if err = s.Repo.createUserSession(hash, sessionExpiry, userID); err != nil {
		return "", err
	}

	return token, nil
}

func (s *Service) logout(token string) error {
	// Hash supplied session token
	hash := util.HashToken(token)

	if err := s.Repo.deleteUserSession(hash); err != nil {
		return err
	}

	return nil
}

func (s *Service) getUsers() ([]User, error) {
	users, err := s.Repo.getUsers()
	if err != nil {
		return nil, err
	}

	return users, nil
}

// Get a user's authorisation, eg admin, API access, etc
func (s *Service) GetUserAuthorisation(userID int64) (UserAuthRequest, error) {
	user, err := s.Repo.getUserAuthorisation(userID)
	if err != nil {
		return UserAuthRequest{}, nil
	}

	return user, nil
}

func (s *Service) updateUserAuthorisation(users []UpdateUserRequest, currentUserID int64) error {
	for _, u := range users {
		if u.UserID == currentUserID {
			return shared.ErrCannotModifyOwnAuth
		}
	}

	if err := s.Repo.updateUserAuthorisation(users); err != nil {
		return err
	}

	return nil
}

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

func (s *Service) deleteUser(targetUserID int64, currentUserID int64) error {
	if targetUserID == currentUserID {
		return shared.ErrCannotDeleteOwnUser
	}

	if err := s.Repo.deleteUser(targetUserID); err != nil {
		return err
	}

	return nil
}

// Change a password for authenticated user
func (s *Service) changePassword(req ChangePasswordRequest) (string, error) {
	// Get email of authenticated user
	email, err := s.GetEmailFromUserID(req.UserID)
	if err != nil {
		return email, err
	}

	if req.OldPassword == "" || req.NewPassword == "" {
		return email, shared.ErrOldOrNewPasswordMissing
	}

	// Check current password
	_, isCurrentPassword, err := s.AuthenticateUser(email, req.OldPassword)
	if err != nil {
		return email, err
	}

	if !isCurrentPassword {
		return email, shared.ErrAuthenticationFailed
	}

	//Validate complexity
	if err := validatePassword(email, req.NewPassword); err != nil {
		return email, fmt.Errorf("%w: %v", shared.ErrInvalidPassword, err)
	}

	// Hash new password and change
	newHash, err := util.HashPassword(req.NewPassword)
	if err != nil {
		return email, err
	}

	if err = s.Repo.changePassword(email, newHash); err != nil {
		return email, err
	}

	return email, nil
}

// Authenticate user, return userID for use in context and match bool
func (s *Service) AuthenticateUser(email string, password string) (int64, bool, error) {
	id, hash, err := s.Repo.getUserIDHashByEmail(email)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return 0, false, shared.ErrUserNotFound
	}

	if errors.Is(err, sql.ErrNoRows) {
		hash = "$argon2id$v=19$m=131072,t=4,p=1$fCSLCAorTbr9UeFcmUW3Jg$q8wabA06xx+zN8j80pwmxTMk0b/T88R+M3ycbFWZPlc" // dummy
		id = 0
	}

	match, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		return 0, false, err
	}

	if id == 0 { // User doesn't exist
		return 0, false, shared.ErrUserNotFound
	}

	if match {
		if err = s.Repo.updateLastLogin(id); err != nil {
			return 0, false, err
		}
	}

	return id, match, nil
}

func (s *Service) GetSessionStatus(token string) (UserSession, error) {
	// Hash token from cookie
	hash := util.HashToken(token)

	userSession, err := s.Repo.getSessionStatus(hash)
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

func (s *Service) ProcessUserSessions() error {
	rows, err := s.Repo.deleteExpiredUserSessions()
	if err != nil {
		s.Logger.Error("Error while deleting expired user sessions", "error", err)
		return err
	}

	if rows == 0 {
		s.Logger.Debug("No expired user sessions to cleanup")
	}

	if rows > 0 {
		s.Logger.Debug("Deleted expired user sessions", "count", rows)
	}

	return nil
}

func (s *Service) GetEmailFromUserID(userID int64) (string, error) {
	email, err := s.Repo.getEmailFromUserID(userID)
	if err != nil {
		return "", err
	}

	return email, nil
}
