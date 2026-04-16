package errutil

import "google.golang.org/grpc/codes"

type domainError struct {
	Message string
	Code    codes.Code
}

func (e domainError) Error() string {
	return e.Message
}

func (e domainError) GRPCStatus() codes.Code {
	return e.Code
}

func New(msg string, code codes.Code) domainError {
	return domainError{
		Message: msg,
		Code:    code,
	}
}
