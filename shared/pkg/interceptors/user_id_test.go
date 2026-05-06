package interceptors

import (
	"context"
	"testing"

	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/common"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestUnaryClientUserIDKey(t *testing.T) {
	tests := []struct {
		name       string
		ctxSetup   func() context.Context
		expectMeta bool
	}{
		{
			name: "Добавление UserID в исходящий контекст",
			ctxSetup: func() context.Context {
				return context.WithValue(context.Background(), common.UserIDKey, int64(456))
			},
			expectMeta: true,
		},
		{
			name: "Отсутствие UserID",
			ctxSetup: func() context.Context {
				return context.Background()
			},
			expectMeta: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interceptor := UnaryClientUserIDKey()
			ctx := tt.ctxSetup()

			invoker := func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
				md, _ := metadata.FromOutgoingContext(ctx)
				if tt.expectMeta {
					assert.Equal(t, "456", md.Get(mdUserIDKey)[0])
				} else {
					assert.Empty(t, md.Get(mdUserIDKey))
				}
				return nil
			}

			err := interceptor(ctx, "/test", nil, nil, nil, invoker)
			assert.NoError(t, err)
		})
	}
}

func TestUnaryServerUserIDKey(t *testing.T) {
	tests := []struct {
		name         string
		ctxSetup     func() context.Context
		expectedID   int64
		expectedCode codes.Code
	}{
		{
			name: "Успешное извлечение валидного UserID",
			ctxSetup: func() context.Context {
				md := metadata.Pairs(mdUserIDKey, "789")
				return metadata.NewIncomingContext(context.Background(), md)
			},
			expectedID:   789,
			expectedCode: codes.OK,
		},
		{
			name: "Отсутствие UserID в метаданных",
			ctxSetup: func() context.Context {
				return context.Background()
			},
			expectedID:   0,
			expectedCode: codes.OK,
		},
		{
			name: "Ошибка: невалидный формат UserID",
			ctxSetup: func() context.Context {
				md := metadata.Pairs(mdUserIDKey, "not-a-number")
				return metadata.NewIncomingContext(context.Background(), md)
			},
			expectedCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interceptor := UnaryServerUserIDKey()
			info := &grpc.UnaryServerInfo{}

			handler := func(ctx context.Context, req any) (any, error) {
				val := ctx.Value(common.UserIDKey)
				if tt.expectedID != 0 {
					assert.Equal(t, tt.expectedID, val)
				} else {
					assert.Nil(t, val)
				}
				return nil, nil
			}

			_, err := interceptor(tt.ctxSetup(), nil, info, handler)

			if tt.expectedCode == codes.OK {
				assert.NoError(t, err)
			} else {
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedCode, st.Code())
			}
		})
	}
}
