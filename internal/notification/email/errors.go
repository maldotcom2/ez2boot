package email

import "errors"

var (
	ErrFieldsMissing     = errors.New("Host, port, to and from fields are required")
	ErrMissingAuthValues = errors.New("Cannot perform authenticated send without username and password")
)
