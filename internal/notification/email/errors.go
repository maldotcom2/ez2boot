package email

import "errors"

var (
	ErrMissingAuthValues = errors.New("Cannot perform authenticated send without username and password")
)
