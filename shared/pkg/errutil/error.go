package errutil

import (
	"fmt"

	"google.golang.org/grpc/codes"
)

type domainError struct {
	slug    string
	message string
	code    codes.Code
	cause   error
}

func (e domainError) Error() string {
	res := fmt.Sprintf("[%s] %s", e.slug, e.message)
	if e.cause != nil {
		res = fmt.Sprintf("%s: %v", res, e.cause)
	}
	return res
}

func (e domainError) GRPCStatus() codes.Code {
	return e.code
}

// Возвращает машинный код ошибки
func (e domainError) Slug() string {
	return e.slug
}

// Создает базовую ошибку (для констант в domain)
func New(slug, msg string, code codes.Code) domainError {
	return domainError{
		slug:    slug,
		message: msg,
		code:    code,
	}
}

// Позволяет добавить контекст к существующей ошибке
func Wrap(slug, msg string, cause error, code codes.Code) domainError {
	return domainError{
		slug:    slug,
		message: msg,
		cause:   cause,
		code:    code,
	}
}

// Если нужно быстро вернуть ошибку без создания переменной
func Message(slug, msg string) domainError {
	return domainError{slug: slug, message: msg, code: codes.Internal}
}

const InternalSlug = "INTERNAL_SERVER_ERROR"

func Internal(msg string, err error) domainError {
	return domainError{slug: InternalSlug, message: msg, code: codes.Internal}
}
