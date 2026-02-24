package user

import (
	"context"
	"database/sql"
	"errors"
	"ez2boot/internal/audit"
	"ez2boot/internal/ctxutil"
	"ez2boot/internal/shared"
	"ez2boot/internal/util"
	"fmt"
	"time"

	"github.com/alexedwards/argon2id"
)

// Attempt user login using even-time
func (s *Service) login(u UserLogin) (token string, err error) {
	var userID int64

	defer func() {
		var reason string
		if err != nil {
			reason = err.Error()
		}

		s.Audit.Log(audit.Event{
			ActorUserID: userID,
			ActorEmail:  u.Email,
			Action:      "login",
			Resource:    "user",
			Success:     err == nil,
			Reason:      reason,
		})
	}()

	// Input validation
	if u.Email == "" || u.Password == "" {
		return "", shared.ErrEmailOrPasswordMissing
	}

	// Generate session token
	token, err = util.GenerateRandomString(32)
	if err != nil {
		return "", err
	}

	hash := util.HashToken(token)
	sessionExpiry := time.Now().Add(s.Config.UserSessionDuration).Unix()

	// Authenticate user
	userID, authenticated, authErr := s.AuthenticateUser(u.Email, u.Password)
	if authErr != nil && authErr != shared.ErrUserNotFound {
		return "", authErr
	}

	if !authenticated || authErr == shared.ErrUserNotFound {
		return "", shared.ErrAuthenticationFailed
	}

	// User exists and is authenticated

	user, err := s.GetUserAuthorisation(userID)
	if err != nil {
		return "", err
	}

	if !user.IsActive {
		return "", shared.ErrUserInactive
	}

	if !user.UIEnabled {
		return "", shared.ErrUserNotAuthorised
	}

	// Store session hash
	if err = s.Repo.createUserSession(hash, sessionExpiry, userID); err != nil {
		return "", err
	}

	return token, nil
}

func (s *Service) logout(token string, ctx context.Context) error {
	// Hash supplied session token
	hash := util.HashToken(token)

	if err := s.Repo.deleteUserSession(hash); err != nil {
		return err
	}

	userID, email := ctxutil.GetActor(ctx)
	s.Audit.Log(audit.Event{
		ActorUserID: userID,
		ActorEmail:  email,
		Action:      "logout",
		Resource:    "user",
		Success:     true,
	})

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
func (s *Service) GetUserAuthorisation(userID int64) (UserAuthResponse, error) {
	user, err := s.Repo.getUserAuthorisation(userID)
	if err != nil {
		return UserAuthResponse{}, nil
	}

	return user, nil
}

func (s *Service) updateUserAuthorisation(users []UpdateUserRequest, ctx context.Context) error {
	userID, email := ctxutil.GetActor(ctx)
	currentUserID := userID
	currentUserEmail := email

	// Open transaction - create atomicity for expected UI experience
	tx, err := s.Repo.Base.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	for _, u := range users {
		if u.UserID == currentUserID {
			tx.Rollback()
			return shared.ErrCannotModifyOwnAuth
		}

		if err := s.Repo.updateUserAuthorisation(tx, u); err != nil {
			tx.Rollback()
			return err
		}

		targetEmail, _ := s.GetEmailFromUserID(u.UserID)

		s.Audit.LogTx(tx, audit.Event{
			ActorUserID:  currentUserID,
			ActorEmail:   currentUserEmail,
			TargetUserID: u.UserID,
			TargetEmail:  targetEmail,
			Action:       "update authorisation",
			Resource:     "user",
			Success:      true,
			Metadata: map[string]any{
				"isActive":   u.IsActive,
				"isAdmin":    u.IsAdmin,
				"apiEnabled": u.APIEnabled,
				"uiEnabled":  u.UIEnabled,
			},
		})
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (s *Service) createUser(req CreateUserRequest, ctx context.Context) error {
	if err := s.validateEmail(req.Email); err != nil {
		return err
	}

	if err := validatePassword(req.Email, req.Password); err != nil {
		return err
	}

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

	targetUserID, err := s.Repo.createUser(user)
	if err != nil {
		return err
	}

	actorUserID, actorEmail := ctxutil.GetActor(ctx)
	s.Audit.Log(audit.Event{
		ActorUserID:  actorUserID,
		ActorEmail:   actorEmail,
		TargetUserID: targetUserID,
		TargetEmail:  req.Email,
		Action:       "create",
		Resource:     "user",
		Success:      true,
	})

	return nil
}

func (s *Service) deleteUser(targetUserID int64, ctx context.Context) error {
	actorUserID, actorEmail := ctxutil.GetActor(ctx)
	if targetUserID == actorUserID {
		return shared.ErrCannotDeleteOwnUser
	}

	targetEmail, _ := s.GetEmailFromUserID(targetUserID)

	if err := s.Repo.deleteUser(targetUserID); err != nil {
		return err
	}

	s.Audit.Log(audit.Event{
		ActorUserID:  actorUserID,
		ActorEmail:   actorEmail,
		TargetUserID: targetUserID,
		TargetEmail:  targetEmail,
		Action:       "delete",
		Resource:     "user",
		Success:      true,
	})

	return nil
}

// Change a password for authenticated user
func (s *Service) changePassword(req ChangePasswordRequest, ctx context.Context) (err error) {
	actorUserID, actorEmail := ctxutil.GetActor(ctx)

	defer func() {
		var reason string
		if err != nil {
			reason = err.Error()
		}

		s.Audit.Log(audit.Event{
			ActorUserID: actorUserID,
			ActorEmail:  actorEmail,
			Action:      "change password",
			Resource:    "user",
			Success:     err == nil,
			Reason:      reason,
		})
	}()

	if req.CurrentPassword == "" || req.NewPassword == "" {
		return shared.ErrCurrentOrNewPasswordMissing
	}

	// Check current password
	_, isCurrentPassword, err := s.AuthenticateUser(actorEmail, req.CurrentPassword)
	if err != nil {
		return err
	}

	if !isCurrentPassword {
		return shared.ErrAuthenticationFailed
	}

	if err := validatePassword(actorEmail, req.NewPassword); err != nil {
		return fmt.Errorf("%w: %v", shared.ErrInvalidPassword, err)
	}

	newHash, err := util.HashPassword(req.NewPassword)
	if err != nil {
		return err
	}

	if err = s.Repo.changePassword(actorUserID, newHash); err != nil {
		return err
	}

	return nil
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

func (s *Service) GetSessionStatus(token string) (UserSessionResponse, error) {
	// Hash token from cookie
	hash := util.HashToken(token)

	userSession, err := s.Repo.getSessionStatus(hash)
	if err != nil {
		if err == sql.ErrNoRows {
			return UserSessionResponse{}, shared.ErrSessionNotFound
		}
		return UserSessionResponse{}, err
	}

	if userSession.SessionExpiry < time.Now().Unix() {
		return UserSessionResponse{}, shared.ErrSessionExpired
	}

	return userSession, nil
}

func (s *Service) ProcessUserSessions() error {
	rows, err := s.Repo.deleteExpiredUserSessions()
	if err != nil {
		s.Logger.Error("Failed to delete expired user sessions", "domain", "user", "error", err)
		return err
	}

	if rows == 0 {
		s.Logger.Debug("No expired user sessions to cleanup", "domain", "user")
	}

	if rows > 0 {
		s.Logger.Debug("Deleted expired user sessions", "domain", "user", "count", rows)
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
