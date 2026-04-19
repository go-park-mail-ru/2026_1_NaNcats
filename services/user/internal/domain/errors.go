package domain

import "errors"

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailAlreadyExists = errors.New("user with this email already exists")
	ErrInvalidInput       = errors.New("invalid input of data")
	ErrInvalidImageExt    = errors.New("invalid image extension")
	ErrImageNotFound      = errors.New("image not found")
	ErrEmptyDBQuery       = errors.New("empty query to database")
)
