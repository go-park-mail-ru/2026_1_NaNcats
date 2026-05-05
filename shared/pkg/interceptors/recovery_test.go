package interceptors

import (
	"context"
	"testing"

	loggerMocks "github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestUnaryServerRecovery(t *testing.T) {
	type mockInit func(m *loggerMocks.MockLogger)

	tests := []struct {
		name         string
		req          any
		handlerPanic bool
		mockInit     mockInit
		expectedCode codes.Code
	}{
		{
			name:         "Успешное выполнение, нет паники",
			req:          "normal request",
			handlerPanic: false,
			mockInit:     func(m *loggerMocks.MockLogger) {},
			expectedCode: codes.OK,
		},
		{
			name:         "Перехват паники, возврат Internal",
			req:          "panic request",
			handlerPanic: true,
			mockInit: func(m *loggerMocks.MockLogger) {
				m.EXPECT().Error("gRPC server panic recovered", gomock.Any(), gomock.Any(), gomock.Any())
			},
			expectedCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockLogger := loggerMocks.NewMockLogger(ctrl)
			tt.mockInit(mockLogger)

			interceptor := UnaryServerRecovery(mockLogger)
			info := &grpc.UnaryServerInfo{FullMethod: "/test.Method"}

			handler := func(ctx context.Context, req any) (any, error) {
				if tt.handlerPanic {
					panic("unexpected boom")
				}
				return "response", nil
			}

			resp, err := interceptor(context.Background(), tt.req, info, handler)

			if tt.expectedCode == codes.OK {
				assert.NoError(t, err)
				assert.Equal(t, "response", resp)
			} else {
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedCode, st.Code())
			}
		})
	}
}
