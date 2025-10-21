package users

import "errors"

var (
	ErrAuthenticationFailed = errors.New("Authentication failed")
	ErrInvalidPassword      = errors.New("Password does not match complexity requirements")
)
