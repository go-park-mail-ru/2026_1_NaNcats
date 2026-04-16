package errutil

import "google.golang.org/grpc/codes"

type domainError struct {
	message string
	code    codes.Code
	cause   error
}

func (e domainError) Error() string {
	return e.message
}

func (e domainError) GRPCStatus() codes.Code {
	return e.code
}

// Создает базовую ошибку (для констант в domain)
func New(msg string, code codes.Code) domainError {
	return domainError{
		message: msg,
		code:    code,
	}
}

// Позволяет добавить контекст к существующей ошибке
func Wrap(msg string, cause error, code codes.Code) domainError {
	return domainError{
		message: msg,
		cause:   cause,
		code:    code,
	}
}

// Если нужно быстро вернуть ошибку без создания переменной
func Message(msg string) domainError {
	return domainError{message: msg, code: codes.Internal}
}
