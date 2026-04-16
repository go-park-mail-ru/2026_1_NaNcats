package grpcutil

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Контракт, чтобы ошибку можно было перевести в gRPC
type StatusCoder interface {
	GRPCStatus() codes.Code
}

// Превращает ошибку Go в ошибку gRPC
func ToGRPCError(err error) error {
	if err == nil {
		return nil
	}

	// Если ошибка реализует наш контракт
	if s, ok := err.(StatusCoder); ok {
		return status.Error(s.GRPCStatus(), err.Error())
	}

	// Если это посторонняя ошибка
	return status.Error(codes.Internal, "internal server error")
}
