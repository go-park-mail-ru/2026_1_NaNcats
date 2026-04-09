package domain

import "errors"

var (
	ErrUserNotFound               = errors.New("user not found")
	ErrEmailAlreadyExists         = errors.New("user with this email already exists")
	ErrPaymentMethodAlreadyExists = errors.New("this payment method already exists")
	ErrInvalidPassword            = errors.New("password too short")
	ErrInvalidEmail               = errors.New("wrong email syntax")
	ErrInvalidCredentials         = errors.New("invalid email or password")
	ErrInvalidInput               = errors.New("invalid input of data")
	ErrSessionNotFound            = errors.New("session not found")
	ErrSessionExpired             = errors.New("session expired")
	ErrRedisResultIsNotOK         = errors.New("result is not OK")
	ErrImageTooLarge              = errors.New("image size exceeds the limit")
	ErrInvalidImageExt            = errors.New("invalid image extension")
	ErrImageNotFound              = errors.New("image not found")
	ErrPaymentMethodNotFound      = errors.New("payment method not found")
	ErrEmptyDBQuery               = errors.New("empty query to database")
	ErrSQLSyntax                  = errors.New("incorrect SQL syntax")
	ErrSQLDeadlock                = errors.New("detected deadlock")
	ErrSQLLockTimeout             = errors.New("lock not avaliable")
	ErrYookassaConfirmationURL    = errors.New("yookassa did not send confirmation url")
	ErrInvalidQuantity            = errors.New("invalid dish quantity")
	ErrDishNotFound               = errors.New("dish not found")
	ErrMultipleRestaurants        = errors.New("restaurant is different in cart and dish")
)
