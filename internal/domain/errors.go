package domain

import "errors"

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailAlreadyExists = errors.New("user with this email already exists")
	ErrInvalidPassword    = errors.New("password too short")
	ErrInvalidEmail       = errors.New("wrong email syntax")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrInvalidInput       = errors.New("invalid input of data")
	ErrSessionNotFound    = errors.New("session not found")
	ErrSessionExpired     = errors.New("session expired")
	ErrImageTooLarge      = errors.New("image size exceeds the limit")
	ErrInvalidImageExt    = errors.New("invalid image extension")
	ErrImageNotFound      = errors.New("image not found")
	ErrEmptyDBQuery       = errors.New("empty query to database")
	ErrSQLSyntax          = errors.New("incorrect SQL syntax")
	ErrSQLDeadlock        = errors.New("detected deadlock")
	ErrSQLLockTimeout     = errors.New("lock not avaliable")
	ErrAddressNotFound    = errors.New("address not found")
)
