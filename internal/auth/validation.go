package auth

import "ez2boot/internal/shared"

func validateLogin(req UserLoginRequest) error {
	if req.Email == "" || req.Password == "" {
		return shared.ErrEmailOrPasswordMissing
	}

	if len(req.Email) > 254 {
		return shared.ErrInputTooLong
	}

	if len(req.Password) > 128 {
		return shared.ErrInputTooLong
	}

	return nil
}
