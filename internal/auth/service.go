package auth

import (
	"context"
	"errors"
	"ez2boot/internal/audit"
	"ez2boot/internal/ctxutil"
	"ez2boot/internal/shared"
	"ez2boot/internal/util"
	"fmt"
	"time"
)

func (s *Service) login(u UserLoginRequest) (token string, mfaRequired bool, err error) {
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
	if err := validateLogin(u); err != nil {
		return "", false, err
	}

	// Authenticate user
	auth, authErr := s.UserService.AuthenticateUser(u.Email, u.Password)

	userID = auth.UserID // For deferred audit logging

	switch auth.IdentityProvider {
	case "local":
		if errors.Is(authErr, shared.ErrUserNotFound) {
			return "", false, shared.ErrUserNotFound
		}

		if authErr != nil {
			return "", false, authErr
		}

		if !auth.Authenticated {
			return "", false, shared.ErrAuthenticationFailed
		}

	case "ldap":
		ldapErr := s.LdapService.Authenticate(u.Email, u.Password)
		if ldapErr != nil {
			if errors.Is(ldapErr, shared.ErrLDAPConfigNotFound) {
				return "", false, ldapErr
			}

			if errors.Is(ldapErr, shared.ErrLDAPConnection) {
				return "", false, ldapErr
			}

			return "", false, fmt.Errorf("%w: %v", shared.ErrAuthenticationFailed, ldapErr)
		}

	case "oidc":
		return "", false, shared.ErrAuthenticationFailed
	default:
		return "", false, shared.ErrUserNotFound
	}

	// User exists and is authenticated
	user, err := s.UserService.GetUserAuthorisation(auth.UserID)
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
