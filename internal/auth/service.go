package auth

import (
	"context"
	"ez2boot/internal/audit"
	"ez2boot/internal/ctxutil"
	"ez2boot/internal/shared"
	"ez2boot/internal/util"
	"time"
)

func (s *Service) login(u UserLogin) (token string, mfaRequired bool, err error) {
	var userID int64

	defer func() {
		// MFA succcess still pending
		if mfaRequired {
			return
		}

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
		return "", false, shared.ErrEmailOrPasswordMissing
	}

	// Authenticate user
	userID, authenticated, authErr := s.UserService.AuthenticateUser(u.Email, u.Password)
	if authErr != nil && authErr != shared.ErrUserNotFound {
		return "", false, authErr
	}

	// User not found
	if authErr == shared.ErrUserNotFound {
		return "", false, shared.ErrUserNotFound
	}

	// Found but not authenticated
	if !authenticated {
		return "", false, shared.ErrAuthenticationFailed
	}

	// User exists and is authenticated
	user, err := s.UserService.GetUserAuthorisation(userID)
	if err != nil {
		return "", false, err
	}

	if !user.IsActive {
		return "", false, shared.ErrUserInactive
	}

	if !user.UIEnabled {
		return "", false, shared.ErrUserNotAuthorised
	}

	// Check if MFA is required
	if user.MFAConfirmed {
		token, err = util.GenerateRandomString(32)
		if err != nil {
			return "", false, err
		}
		hash := util.HashToken(token)
		expiry := time.Now().Add(3 * time.Minute).Unix()
		if err = s.UserService.CreateMFAPendingSession(hash, expiry, userID); err != nil {
			return "", false, err
		}
		return token, true, nil
	}

	// Create user session
	token, err = s.UserService.CreateSession(userID)

	return token, false, nil
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
