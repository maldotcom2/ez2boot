package auth

import (
	"context"
	"ez2boot/internal/audit"
	"ez2boot/internal/ctxutil"
	"ez2boot/internal/shared"
	"ez2boot/internal/util"
	"time"
)

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
	userID, authenticated, authErr := s.UserService.AuthenticateUser(u.Email, u.Password)
	if authErr != nil && authErr != shared.ErrUserNotFound {
		return "", authErr
	}

	// User not found
	if authErr == shared.ErrUserNotFound {
		return "", shared.ErrUserNotFound
	}

	// Found but not authenticated
	if !authenticated {
		return "", shared.ErrAuthenticationFailed
	}

	// User exists and is authenticated

	user, err := s.UserService.GetUserAuthorisation(userID)
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
	if err = s.UserService.CreateUserSession(hash, sessionExpiry, userID); err != nil {
		return "", err
	}

	return token, nil
}

func (s *Service) logout(token string, ctx context.Context) error {
	// Hash supplied session token
	hash := util.HashToken(token)

	if err := s.UserService.DeleteUserSession(hash); err != nil {
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
