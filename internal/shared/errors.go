package shared

import "errors"

var (
	ErrAuthenticationFailed    = errors.New("authentication failed")
	ErrInvalidPassword         = errors.New("password does not match complexity requirements")
	ErrSessionNotFound         = errors.New("supplied session token not found")
	ErrSessionExpired          = errors.New("session for supplied token has expired")
	ErrUserNotFound            = errors.New("user not found")
	ErrUserAlreadyExists       = errors.New("User already exists")
	ErrAuthTypeDenied          = errors.New("user not allowed to use this auth type")
	ErrUserNotActive           = errors.New("user is not active")
	ErrNoRowsDeleted           = errors.New("no rows were deleted")
	ErrNoRowsUpdated           = errors.New("no rows were updated")
	ErrPasswordLength          = errors.New("password must be 14 chars")
	ErrPasswordContainsEmail   = errors.New("password contains email address")
	ErrEmailContainsPassword   = errors.New("email contains password")
	ErrEmailPattern            = errors.New("email does not match required pattern")
	ErrEmailOrPasswordMissing  = errors.New("email and password field required")
	ErrEmailMissing            = errors.New("email field missing")
	ErrOldOrNewPasswordMissing = errors.New("old_password and new_password field required")
	ErrCannotModifyOwnAuth     = errors.New("cannot modify own authorisation")
)
