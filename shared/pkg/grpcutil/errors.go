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

	// Если ошибка уже пришла от gRPC
	if _, ok := status.FromError(err); ok {
		return err // Ошибка уже в правильном формате, просто пробрасываем
	}

	// Если ошибка реализует наш контракт (пришла из другого слоя)
	if s, ok := err.(StatusCoder); ok {
		return status.Error(s.GRPCStatus(), err.Error())
	}

	// Если это посторонняя ошибка
	return status.Error(codes.Internal, "internal server error")
}
