package shared

import "errors"

var (
	ErrAuthenticationFailed = errors.New("Authentication failed")
	ErrInvalidPassword      = errors.New("Password does not match complexity requirements")
	ErrSessionNotFound      = errors.New("Supplied session token not found")
	ErrSessionExpired       = errors.New("Session for supplied token has expired")
	ErrUserNotFound         = errors.New("User not found")
	ErrAuthTypeDenied       = errors.New("User not allowed to use this auth type")
	ErrUserNotActive        = errors.New("User is not active")
	ErrNoRowsDeleted        = errors.New("No rows were deleted")
	ErrNoRowsUpdated        = errors.New("No rows were updated")
)
