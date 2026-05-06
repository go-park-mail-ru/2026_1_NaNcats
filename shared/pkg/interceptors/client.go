package interceptors

import (
	"context"
	"strconv"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/common"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func UnaryClientUserID() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if userID, ok := ctx.Value(common.UserIDKey).(int64); ok {
			ctx = metadata.AppendToOutgoingContext(ctx, mdUserIDKey, strconv.FormatInt(userID, 10))
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func UnaryClientRequestID() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if reqID, ok := ctx.Value(common.RequestIDKey).(string); ok {
			ctx = metadata.AppendToOutgoingContext(ctx, mdRequestIDKey, reqID)
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func UnaryClientLogging(l logger.Logger) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		start := time.Now()
		reqID, _ := ctx.Value(common.RequestIDKey).(string)
		err := invoker(ctx, method, req, reply, cc, opts...)
		duration := time.Since(start)

		if err != nil {
			l.Error("grpc outgoing call failed", err,
				logger.String("method", method),
				logger.String("duration", duration.String()),
				logger.String("request_id", reqID),
				logger.String("target", cc.Target()),
			)
		} else {
			l.Debug("grpc outgoing call succeeded",
				logger.String("method", method),
				logger.String("duration", duration.String()),
				logger.String("request_id", reqID),
			)
		}

		return err
	}
}
