package domain

import "errors"

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user with this email already exists")
	ErrInvalidEmail      = errors.New("wrong email syntax")
	ErrSessionNotFound   = errors.New("session not found")
	ErrSessionExpired    = errors.New("session expired")
	ErrImageTooLarge     = errors.New("image size exceeds the limit")
	ErrInvalidImageExt   = errors.New("invalid image extension")
	ErrImageNotFound     = errors.New("image not found")
)
