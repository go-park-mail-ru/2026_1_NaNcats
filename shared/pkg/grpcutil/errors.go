package grpcutil

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Контракт, чтобы ошибку можно было перевести в gRPC
type StatusCoder interface {
	GRPCStatus() codes.Code
}

// Интерфейс для чтения Слага
type Slugger interface {
	Slug() string
}

// Превращает ошибку Go в ошибку gRPC
func ToGRPCError(err error) error {
	if err == nil {
		return nil
	}

	// Если ошибка уже пришла от gRPC
	if _, ok := status.FromError(err); ok {
		return err
	}

	// Если ошибка реализует наш контракт (пришла из другого слоя)
	if s, ok := err.(StatusCoder); ok {
		message := err.Error()

		if slugErr, ok := err.(Slugger); ok {
			message = slugErr.Slug()
		}

		return status.Error(s.GRPCStatus(), message)
	}

	// Если это посторонняя ошибка — пробрасываем оригинальный текст,
	// чтобы при дебаге было видно реальную причину.
	return status.Error(codes.Internal, err.Error())
}
