package interceptors

import (
	"context"
	"fmt"
	"runtime/debug"

	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Перехватывает паники внутри запросов
func UnaryServerRecovery(l logger.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		defer func() {
			if r := recover(); r != nil {
				// Получаем стек вызовов (где именно все упало)
				stackTrace := string(debug.Stack())

				l.Error("gRPC server panic recovered",
					fmt.Errorf("%v", r),
					logger.String("method", info.FullMethod),
					logger.String("stack_trace", stackTrace),
				)

				// Перезаписываем ответ для клиента
				err = status.Errorf(codes.Internal, "internal server error")
			}
		}()

		return handler(ctx, req)
	}
}
