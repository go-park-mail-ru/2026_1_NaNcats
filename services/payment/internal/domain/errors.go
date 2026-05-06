package domain

import "errors"

var (
	ErrPaymentMethodAlreadyExists = errors.New("payment method already exists")
	ErrPaymentMethodNotFound      = errors.New("payment method not found")
	ErrRedisResultIsNotOK         = errors.New("redis result is not OK")
	ErrUserIDNotFoundInCache      = errors.New("user id not found in cache for this payment")
)
