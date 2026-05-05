package grpc_client

import (
	"context"
	"errors"
	"testing"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/payment/internal/infrastructure/grpc_client/mocks"
	pb "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/order"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

//go:generate mockgen -destination=mocks/order_client_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/order OrderServiceClient

func TestOrderClient_UpdateOrderStatus(t *testing.T) {
	type mockInit func(m *mocks.MockOrderServiceClient)

	tests := []struct {
		name        string
		paymentID   string
		status      string
		mockInit    mockInit
		expectedErr error
	}{
		{
			name:      "Успешное обновление статуса заказа",
			paymentID: "pay-123",
			status:    "succeeded",
			mockInit: func(m *mocks.MockOrderServiceClient) {
				m.EXPECT().UpdateOrderStatusByPaymentID(
					gomock.Any(),
					&pb.UpdateStatusRequest{
						YookassaPaymentId: "pay-123",
						Status:            "succeeded",
					},
				).Return(nil, nil) // Возвращаем nil в качестве ответа и nil в качестве ошибки
			},
			expectedErr: nil,
		},
		{
			name:      "Ошибка сети или gRPC при обновлении",
			paymentID: "pay-err",
			status:    "canceled",
			mockInit: func(m *mocks.MockOrderServiceClient) {
				m.EXPECT().UpdateOrderStatusByPaymentID(
					gomock.Any(),
					&pb.UpdateStatusRequest{
						YookassaPaymentId: "pay-err",
						Status:            "canceled",
					},
				).Return(nil, errors.New("rpc error: context deadline exceeded"))
			},
			expectedErr: errors.New("rpc error: context deadline exceeded"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockPBClient := mocks.NewMockOrderServiceClient(ctrl)
			tt.mockInit(mockPBClient)

			client := NewOrderClient(mockPBClient)

			// Act
			err := client.UpdateOrderStatus(context.Background(), tt.paymentID, tt.status)

			// Assert
			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
