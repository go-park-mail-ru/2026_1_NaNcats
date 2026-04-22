package interceptors

import (
	"context"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/common"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const mdRequestIDKey = "x-request-id"

// Автоматическая запись каждого gRPC вызова в логи
func UnaryServerLogging(l logger.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{}, // поступивший запрос
		info *grpc.UnaryServerInfo, // метаданные
		handler grpc.UnaryHandler, // хендлер на выполнение
	) (interface{}, error) {
		start := time.Now()

		var reqID string
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			if ids := md.Get(mdRequestIDKey); len(ids) > 0 {
				reqID = ids[0]
			}
		}

		if reqID == "" {
			// Генерируем временный ID, чтобы логи не были пустыми
			reqID = "orphan-" + uuid.NewString()
		}

		ctx = context.WithValue(ctx, common.RequestIDKey, reqID)

		resp, err := handler(ctx, req)
		if err != nil {
			l.Error("grpc call failed", err,
				logger.String("method", info.FullMethod),
				logger.String("duration", time.Since(start).String()),
				logger.String("request_id", reqID),
			)
		} else {
			l.Debug("grpc call succeeded",
				logger.String("method", info.FullMethod),
				logger.String("duration", time.Since(start).String()),
				logger.String("request_id", reqID),
			)
		}

		return resp, err
	}
}
