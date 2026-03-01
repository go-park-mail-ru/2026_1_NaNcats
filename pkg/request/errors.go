package request

import "errors"

var (
	ErrBodyTooLarge   = errors.New("request body is too large")
	ErrInvalidJSON    = errors.New("request body contains badly-formed JSON")
	ErrEmptyBody      = errors.New("request body is empty")
	ErrNotOnlyJSONVal = errors.New("body must only contain a single JSON value")
)
