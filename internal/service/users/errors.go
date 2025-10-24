package users

import "errors"

var (
	ErrAuthenticationFailed = errors.New("Authentication failed")
	ErrInvalidPassword      = errors.New("Password does not match complexity requirements")
	ErrSessionNotFound      = errors.New("Supplied session token not found")
	ErrSessionExpired       = errors.New("Session for supplied token has expired")
	ErrUserNotFound         = errors.New("User not found")
)
