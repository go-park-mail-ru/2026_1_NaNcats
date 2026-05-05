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

func TestUnaryServerLogging(t *testing.T) {
	type mockInit func(m *loggerMocks.MockLogger)

	tests := []struct {
		name       string
		ctxSetup   func() context.Context
		handlerErr error
		mockInit   mockInit
	}{
		{
			name: "Успешный вызов с RequestID из метаданных",
			ctxSetup: func() context.Context {
				md := metadata.Pairs(mdRequestIDKey, "existing-req-id")
				return metadata.NewIncomingContext(context.Background(), md)
			},
			handlerErr: nil,
			mockInit: func(m *loggerMocks.MockLogger) {
				m.EXPECT().Debug("grpc call succeeded", gomock.Any(), gomock.Any(), gomock.Any())
			},
		},
		{
			name: "Ошибка вызова, генерация orphan RequestID",
			ctxSetup: func() context.Context {
				return context.Background()
			},
			handlerErr: errors.New("some handler error"),
			mockInit: func(m *loggerMocks.MockLogger) {
				m.EXPECT().Error("grpc call failed", errors.New("some handler error"), gomock.Any(), gomock.Any(), gomock.Any())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockLogger := loggerMocks.NewMockLogger(ctrl)
			tt.mockInit(mockLogger)

			interceptor := UnaryServerLogging(mockLogger)
			info := &grpc.UnaryServerInfo{FullMethod: "/test.Method"}

			handler := func(ctx context.Context, req any) (any, error) {
				reqID, ok := ctx.Value(common.RequestIDKey).(string)
				assert.True(t, ok)
				assert.NotEmpty(t, reqID)

				return nil, tt.handlerErr
			}

			_, err := interceptor(tt.ctxSetup(), nil, info, handler)

			assert.Equal(t, tt.handlerErr, err)
		})
	}
}
