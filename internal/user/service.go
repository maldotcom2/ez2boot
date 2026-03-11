package user

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"ez2boot/internal/audit"
	"ez2boot/internal/ctxutil"
	"ez2boot/internal/shared"
	"ez2boot/internal/util"
	"fmt"
	"image/png"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/pquerna/otp/totp"
)

func (s *Service) CreateUserSession(hash string, expiry int64, userID int64) error {
	if err := s.Repo.createUserSession(hash, expiry, userID); err != nil {
		return err
	}

	return nil
}

func (s *Service) DeleteUserSession(hash string) error {
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
	auth, err := s.AuthenticateUser(actorEmail, req.CurrentPassword)
	if err != nil {
		return err
	}

	if !auth.Authenticated {
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

// Authenticate user with even time, returns userID, IDP, match
func (s *Service) AuthenticateUser(email string, password string) (shared.AuthResult, error) {
	user, err := s.Repo.getUserInfoByEmail(email)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return shared.AuthResult{}, err // generic error other than no user
	}

	// User not found
	if errors.Is(err, sql.ErrNoRows) {
		user.PasswordHash = "$argon2id$v=19$m=131072,t=4,p=1$fCSLCAorTbr9UeFcmUW3Jg$q8wabA06xx+zN8j80pwmxTMk0b/T88R+M3ycbFWZPlc" // dummy
		user.IdentityProvider = "local"
		user.UserID = 0
	}

	// User found, has external IDP
	if user.IdentityProvider != "local" {
		return shared.AuthResult{UserID: user.UserID, IdentityProvider: user.IdentityProvider, Authenticated: false}, nil
	}

	// User exists and is local, but no password - defensive
	if user.PasswordHash == "" {
		return shared.AuthResult{UserID: user.UserID, IdentityProvider: user.IdentityProvider, Authenticated: false}, shared.ErrNoLocalPassword
	}

	match, err := argon2id.ComparePasswordAndHash(password, user.PasswordHash)
	if err != nil {
		return shared.AuthResult{UserID: user.UserID, IdentityProvider: user.IdentityProvider, Authenticated: false}, err
	}

	if user.UserID == 0 { // User doesn't exist
		return shared.AuthResult{UserID: user.UserID, IdentityProvider: user.IdentityProvider, Authenticated: false}, shared.ErrUserNotFound
	}

	if match {
		if err = s.Repo.updateLastLogin(user.UserID); err != nil {
			return shared.AuthResult{UserID: user.UserID, IdentityProvider: user.IdentityProvider, Authenticated: false}, err
		}
	}

	return shared.AuthResult{UserID: user.UserID, IdentityProvider: user.IdentityProvider, Authenticated: match}, nil
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

func (s *Service) enrolMFA(userID int64, email string) (_ []byte, err error) {
	defer func() {
		var reason string
		if err != nil {
			reason = err.Error()
		}

		s.Audit.Log(audit.Event{
			ActorUserID: userID,
			ActorEmail:  email,
			Action:      "enrol mfa",
			Resource:    "user",
			Success:     err == nil,
			Reason:      reason,
		})
	}()

	user, err := s.GetUserAuthorisation(userID)
	if err != nil {
		return nil, err
	}

	// SSO users would get MFA via IDP
	if user.IdentityProvider == "oidc" {
		return nil, shared.ErrMFANotSupported
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "ez2boot",
		AccountName: email,
	})
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	img, err := key.Image(200, 200)
	if err != nil {
		return nil, err
	}

	if err = png.Encode(&buf, img); err != nil {
		return nil, err
	}

	// store secret
	secret := key.Secret()
	rows, err := s.Repo.setMFASecret(&secret, userID)
	if err != nil {
		return nil, err
	}

	if rows == 0 {
		return nil, shared.ErrNoRowsUpdated
	}

	return buf.Bytes(), nil
}

// Initial enrolment only
func (s *Service) confirmMFA(req MFARequest, ctx context.Context) (err error) {
	actorUserID, actorEmail := ctxutil.GetActor(ctx)

	defer func() {
		var reason string
		if err != nil {
			reason = err.Error()
		}

		s.Audit.Log(audit.Event{
			ActorUserID: actorUserID,
			ActorEmail:  actorEmail,
			Action:      "confirm mfa",
			Resource:    "user",
			Success:     err == nil,
			Reason:      reason,
		})
	}()

	ok, err := s.checkMFA(req)
	if err != nil {
		return err
	}

	if !ok {
		return shared.ErrIncorrectMFACode
	}

	rows, err := s.Repo.confirmMFA(req.UserID)
	if err != nil {
		return err
	}

	if rows == 0 { // Nothing happened, user trying to validate when already validated
		return shared.ErrNoRowsUpdated
	}

	return nil
}

func (s *Service) checkMFA(req MFARequest) (bool, error) {
	// Get secret from DB
	secret, err := s.Repo.getMFASecret(req.UserID)
	if err != nil {
		return false, err
	}

	if secret == nil {
		return false, shared.ErrMFANotEnrolled
	}

	// Check if code already used
	if s.MFACache.Has(req.UserID, req.Code) {
		return false, nil
	}

	if !totp.Validate(req.Code, *secret) {
		return false, nil
	}

	// Add used code to cache
	s.MFACache.Set(req.UserID, req.Code)

	return true, nil
}

func (s *Service) deleteMFA(req MFARequest, ctx context.Context) (err error) {
	actorUserID, actorEmail := ctxutil.GetActor(ctx)

	defer func() {
		var reason string
		if err != nil {
			reason = err.Error()
		}

		s.Audit.Log(audit.Event{
			ActorUserID: actorUserID,
			ActorEmail:  actorEmail,
			Action:      "delete mfa",
			Resource:    "user",
			Success:     err == nil,
			Reason:      reason,
		})
	}()

	// Check code
	ok, err := s.checkMFA(req)
	if err != nil {
		return err
	}

	if !ok {
		return shared.ErrIncorrectMFACode
	}

	// null the secret
	rows, err := s.Repo.setMFASecret(nil, req.UserID)
	if err != nil {
		return err
	}

	// This is defensive - should never run
	if rows == 0 {
		return shared.ErrNoRowsUpdated
	}

	return nil
}

func (s *Service) verifyMFA(req MFARequest, pendingToken string) (_ string, _ string, err error) {
	// Public handler, no context injection
	var actorUserID int64
	var actorEmail string

	defer func() {
		var reason string
		if err != nil {
			reason = err.Error()
		}

		s.Audit.Log(audit.Event{
			ActorUserID: actorUserID,
			ActorEmail:  actorEmail,
			Action:      "login",
			Resource:    "user",
			Success:     err == nil,
			Reason:      reason,
		})
	}()

	hash := util.HashToken(pendingToken)

	// Look up pending session
	m, err := s.Repo.getMFAPendingSessionStatus(hash)
	if err != nil {
		return "", "", shared.ErrSessionNotFound
	}

	actorUserID = m.UserID
	actorEmail = m.Email

	if time.Now().Unix() > m.SessionExpiry {
		return "", m.Email, shared.ErrSessionExpired
	}

	req.UserID = m.UserID

	// Validate TOTP code
	ok, err := s.checkMFA(req)
	if err != nil {
		return "", m.Email, err
	}

	if !ok {
		return "", m.Email, shared.ErrIncorrectMFACode
	}

	// Delete pending session
	if err = s.Repo.deleteMFAPendingSession(hash); err != nil {
		if errors.Is(err, shared.ErrNoRowsDeleted) {
			s.Logger.Warn("Failed to delete mfa_pending_session", "user", m.Email, "domain", "user", "error", err)
		} else {
			return "", m.Email, err
		}
	}

	// Create user session
	token, err := s.CreateSession(m.UserID)

	return token, m.Email, nil
}

func (s *Service) CreateMFAPendingSession(tokenHash string, expiry int64, userID int64) error {
	if err := s.Repo.createMFAPendingSession(tokenHash, expiry, userID); err != nil {
		return err
	}

	return nil
}

// Create user session
func (s *Service) CreateSession(userID int64) (string, error) {
	token, err := util.GenerateRandomString(32)
	if err != nil {
		return "", err
	}
	hash := util.HashToken(token)
	expiry := time.Now().Add(s.Config.UserSessionDuration).Unix()
	if err = s.Repo.createUserSession(hash, expiry, userID); err != nil {
		return "", err
	}
	return token, nil
}
