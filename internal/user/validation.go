package user

import "errors"

func (s *Service) validateChangePassword(req ChangePasswordRequest) error {
	if req.Email == "" {
		return errors.New("Missing email")
	}

	if req.OldPassword == "" {
		return errors.New("Missing old password")
	}

	if req.NewPassword == "" {
		return errors.New("Missing new password")
	}

	return nil
}
