package interceptors

import (
	"context"
	"errors"
	"testing"

	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/common"
	loggerMocks "github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestUnaryClientUserID(t *testing.T) {
	tests := []struct {
		name       string
		ctxSetup   func() context.Context
		expectMeta bool
	}{
		{
			name: "Успешное добавление UserID в метаданные",
			ctxSetup: func() context.Context {
				return context.WithValue(context.Background(), common.UserIDKey, int64(123))
			},
			expectMeta: true,
		},
		{
			name: "Отсутствие UserID в контексте, метаданные не меняются",
			ctxSetup: func() context.Context {
				return context.Background()
			},
			expectMeta: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interceptor := UnaryClientUserID()
			ctx := tt.ctxSetup()

			invoker := func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
				md, ok := metadata.FromOutgoingContext(ctx)
				if tt.expectMeta {
					assert.True(t, ok)
					assert.Equal(t, "123", md.Get(mdUserIDKey)[0])
				} else {
					assert.Empty(t, md.Get(mdUserIDKey))
				}
				return nil
			}

			err := interceptor(ctx, "/test.Method", nil, nil, nil, invoker)
			assert.NoError(t, err)
		})
	}
}

func TestUnaryClientRequestID(t *testing.T) {
	tests := []struct {
		name       string
		ctxSetup   func() context.Context
		expectMeta bool
	}{
		{
			name: "Успешное добавление RequestID в метаданные",
			ctxSetup: func() context.Context {
				return context.WithValue(context.Background(), common.RequestIDKey, "req-123")
			},
			expectMeta: true,
		},
		{
			name: "Отсутствие RequestID в контексте",
			ctxSetup: func() context.Context {
				return context.Background()
			},
			expectMeta: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interceptor := UnaryClientRequestID()
			ctx := tt.ctxSetup()

			invoker := func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
				md, _ := metadata.FromOutgoingContext(ctx)
				if tt.expectMeta {
					assert.Equal(t, "req-123", md.Get(mdRequestIDKey)[0])
				} else {
					assert.Empty(t, md.Get(mdRequestIDKey))
				}
				return nil
			}

			err := interceptor(ctx, "/test.Method", nil, nil, nil, invoker)
			assert.NoError(t, err)
		})
	}
}

func TestUnaryClientLogging(t *testing.T) {
	type mockInit func(m *loggerMocks.MockLogger)

	tests := []struct {
		name        string
		invokerErr  error
		mockInit    mockInit
		expectedErr error
	}{
		{
			name:       "Успешный вызов, пишется Debug лог",
			invokerErr: nil,
			mockInit: func(m *loggerMocks.MockLogger) {
				m.EXPECT().Debug("grpc outgoing call succeeded", gomock.Any(), gomock.Any(), gomock.Any())
			},
			expectedErr: nil,
		},
		{
			name:       "Вызов с ошибкой, пишется Error лог",
			invokerErr: errors.New("network error"),
			mockInit: func(m *loggerMocks.MockLogger) {
				m.EXPECT().Error("grpc outgoing call failed", errors.New("network error"), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())
			},
			expectedErr: errors.New("network error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockLogger := loggerMocks.NewMockLogger(ctrl)
			tt.mockInit(mockLogger)

			interceptor := UnaryClientLogging(mockLogger)
			cc := &grpc.ClientConn{}

			invoker := func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
				return tt.invokerErr
			}

			err := interceptor(context.Background(), "/test.Method", nil, nil, cc, invoker)

			assert.Equal(t, tt.expectedErr, err)
		})
	}
}
