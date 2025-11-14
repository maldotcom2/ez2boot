package telegram

import "errors"

var (
	ErrMissingValues = errors.New("Token and ChatID are required")
)
