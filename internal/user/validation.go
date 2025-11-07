package user

import (
	"ez2boot/internal/shared"
	"regexp"
	"strings"
	"unicode/utf8"
)

// Check for something@something
func (s *Service) validateEmail(email string) error {
	re := regexp.MustCompile(`^[^\s@]+@[^\s@]+$`)
	if !re.MatchString(email) {
		return shared.ErrEmailPattern
	}

	return nil
}

// Password rules
func validatePassword(email string, password string) error {
	length := utf8.RuneCountInString(password)

	if length < 14 {
		return shared.ErrPasswordLength
	}

	if strings.Contains(password, email) {
		return shared.ErrPasswordContainsEmail
	}

	if strings.Contains(email, password) {
		return shared.ErrEmailContainsPassword
	}

	return nil
}
