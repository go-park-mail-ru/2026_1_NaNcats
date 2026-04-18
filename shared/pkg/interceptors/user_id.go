package interceptors

import (
	"context"
	"strconv"

	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/common"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const mdUserIDKey = "x-user-id"

// Сторона API Gateway

// Берет User ID из контекста и отправляет в gRPC
func UnaryClientUserIDKey() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		if userID, ok := ctx.Value(common.UserIDKey).(int64); ok {
			md := metadata.Pairs(mdUserIDKey, strconv.FormatInt(userID, 10))
			ctx = metadata.NewOutgoingContext(ctx, md)
		}

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// Сторона микросервисов

// Достает userID из метаданных и кладет его в контекст Go
func UnaryServerInterceptorIDKey() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			if ids := md.Get(mdUserIDKey); len(ids) > 0 {
				userID, err := strconv.ParseInt(ids[0], 10, 64)
				if err != nil {
					return nil, status.Error(codes.InvalidArgument, "malformed user-id in metadata")
				}
				ctx = context.WithValue(ctx, common.UserIDKey, userID)
			}
		}
		return handler(ctx, req)
	}
}
