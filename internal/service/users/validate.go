package users

import (
	"errors"
	"strings"
	"unicode/utf8"
)

func validatePassword(username string, password string) error {
	length := utf8.RuneCountInString(password)

	if length < 14 {
		return errors.New("Password must be 14 characters or more")
	}

	if strings.Contains(password, username) {
		return errors.New("Password cannot contain the username")
	}

	if strings.Contains(password, username) {
		return errors.New("Username cannot contain the password")
	}

	return nil
}
