package interceptors

import (
	"context"
	"strconv"

	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/common"
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
