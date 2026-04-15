package middleware

import "errors"

var (
	ErrNoUserIDInContext = errors.New("user id not found in context")
)
