package domain

import "errors"

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user with this email already exists")
	ErrWrongEmailSyntax  = errors.New("wrong email syntax") // временная ошибка, если будем использовать `net/mail`, то её нужно будет убрать
)
