package user

import (
	"errors"
	"regexp"
	"strings"
	"unicode/utf8"
)

// Check for something@something
func (s *Service) validateEmail(email string) error {
	re := regexp.MustCompile(`^[^\s@]+@[^\s@]+$`)
	if !re.MatchString(email) {
		return errors.New("Email does not match required pattern")
	}

	return nil
}

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

// Password rules
func validatePassword(email string, password string) error {
	length := utf8.RuneCountInString(password)

	if length < 14 {
		return errors.New("Password must be 14 characters or more")
	}

	if strings.Contains(password, email) {
		return errors.New("Password cannot contain the email")
	}

	if strings.Contains(password, email) {
		return errors.New("email cannot contain the password")
	}

	return nil
}
