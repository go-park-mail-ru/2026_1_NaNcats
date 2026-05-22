package grpc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/order/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/order/internal/usecase/mocks"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/grpcutil"
	pb "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/order"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestOrderHandler_CreateOrder(t *testing.T) {
	type mockInit func(m *mocks.MockOrderUseCase)

	tests := []struct {
		name         string
		req          *pb.CreateOrderRequest
		mockInit     mockInit
		expectedCode codes.Code
	}{
		{
			name: "Успешное создание заказа",
			req: &pb.CreateOrderRequest{
				UserId:             1,
				AddressPublicId:    "addr-123",
				RestaurantBranchId: 10,
				RestaurantBrandId:  20,
				DeliveryCost:       150,
				ServiceFee:         29,
				PaymentMethodId:    "pm-123",
				PayForAll:          true,
				PayerMapping:       map[int64]int64{2: 500, 3: 1000},
				IdempotencyKey:     "idem-key",
			},
			mockInit: func(m *mocks.MockOrderUseCase) {
				expectedInput := domain.CreateOrderInput{
					UserID:             1,
					AddressPublicID:    "addr-123",
					RestaurantBranchID: 10,
					RestaurantBrandID:  20,
					DeliveryCost:       150,
					ServiceFee:         29,
					PaymentMethodID:    "pm-123",
					PayForAll:          true,
					PayerMapping:       map[int64]int64{2: 500, 3: 1000},
				}
				m.EXPECT().CreateOrder(gomock.Any(), expectedInput, "idem-key").
					Return("pub-order-123", nil)
			},
			expectedCode: codes.OK,
		},
		{
			name: "Системная ошибка в UseCase",
			req: &pb.CreateOrderRequest{
				UserId:         1,
				IdempotencyKey: "idem-err",
			},
			mockInit: func(m *mocks.MockOrderUseCase) {
				m.EXPECT().CreateOrder(gomock.Any(), gomock.Any(), "idem-err").
					Return("", errors.New("internal db error"))
			},
			expectedCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUC := mocks.NewMockOrderUseCase(ctrl)
			tt.mockInit(mockUC)

			handler := NewOrderHandler(mockUC)
			resp, err := handler.CreateOrder(context.Background(), tt.req)

			grpcErr := grpcutil.ToGRPCError(err)

			if tt.expectedCode == codes.OK {
				assert.NoError(t, grpcErr)
				assert.NotNil(t, resp)
				assert.Equal(t, "pub-order-123", resp.OrderPublicId)
			} else {
				st, ok := status.FromError(grpcErr)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedCode, st.Code())
			}
		})
	}
}

func TestOrderHandler_GetOrders(t *testing.T) {
	type mockInit func(m *mocks.MockOrderUseCase)

	ownerID := int64(10)
	now := time.Now()

	tests := []struct {
		name         string
		req          *pb.GetOrdersRequest
		mockInit     mockInit
		expectedCode codes.Code
	}{
		{
			name: "Успешное получение списка заказов",
			req: &pb.GetOrdersRequest{
				UserId: 1,
				Limit:  10,
				Offset: 0,
			},
			mockInit: func(m *mocks.MockOrderUseCase) {
				m.EXPECT().GetOrders(gomock.Any(), int64(1), int32(10), int32(0)).Return([]domain.Order{
					{
						PublicID:       "pub-1",
						RestaurantName: "KFC",
						TotalCost:      1500,
						Status:         "created",
						CreatedAt:      now,
						Items: []domain.OrderDish{
							{DishID: 100, Quantity: 2, Price: 500, OwnerUserID: &ownerID},
							{DishID: 101, Quantity: 1, Price: 500, OwnerUserID: nil},
						},
						Splits: []domain.OrderSplit{
							{ID: "split-1", UserID: 1, Amount: 1500, Status: "pending"},
						},
					},
				}, nil)
			},
			expectedCode: codes.OK,
		},
		{
			name: "Ошибка в UseCase",
			req: &pb.GetOrdersRequest{
				UserId: 1,
				Limit:  10,
				Offset: 20,
			},
			mockInit: func(m *mocks.MockOrderUseCase) {
				m.EXPECT().GetOrders(gomock.Any(), int64(1), int32(10), int32(20)).Return(nil, errors.New("db error"))
			},
			expectedCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUC := mocks.NewMockOrderUseCase(ctrl)
			tt.mockInit(mockUC)

			handler := NewOrderHandler(mockUC)
			resp, err := handler.GetOrders(context.Background(), tt.req)

			grpcErr := grpcutil.ToGRPCError(err)

			if tt.expectedCode == codes.OK {
				require.NoError(t, grpcErr)
				require.NotNil(t, resp)
				require.Len(t, resp.Orders, 1)

				mappedOrder := resp.Orders[0]
				assert.Equal(t, "pub-1", mappedOrder.PublicId)
				assert.Equal(t, "KFC", mappedOrder.RestaurantName)
				assert.Len(t, mappedOrder.Items, 2)
				assert.Len(t, mappedOrder.Splits, 1)

				assert.NotNil(t, mappedOrder.Items[0].OwnerUserId)
				assert.Equal(t, int64(10), *mappedOrder.Items[0].OwnerUserId)
				assert.Nil(t, mappedOrder.Items[1].OwnerUserId)
			} else {
				st, _ := status.FromError(grpcErr)
				assert.Equal(t, tt.expectedCode, st.Code())
			}
		})
	}
}

func TestOrderHandler_UpdateOrderStatusByPaymentID(t *testing.T) {
	type mockInit func(m *mocks.MockOrderUseCase)

	tests := []struct {
		name         string
		req          *pb.UpdateStatusRequest
		mockInit     mockInit
		expectedCode codes.Code
	}{
		{
			name: "Успешное обновление статуса",
			req: &pb.UpdateStatusRequest{
				YookassaPaymentId: "pay-123",
				Status:            "paid",
				IdempotencyKey:    "idem-1",
			},
			mockInit: func(m *mocks.MockOrderUseCase) {
				m.EXPECT().UpdateOrderStatusByPaymentID(gomock.Any(), "pay-123", "paid", "idem-1").Return(nil)
			},
			expectedCode: codes.OK,
		},
		{
			name: "Ошибка обновления статуса",
			req:  &pb.UpdateStatusRequest{YookassaPaymentId: "pay-err"},
			mockInit: func(m *mocks.MockOrderUseCase) {
				m.EXPECT().UpdateOrderStatusByPaymentID(gomock.Any(), "pay-err", "", "").
					Return(errors.New("not found"))
			},
			expectedCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUC := mocks.NewMockOrderUseCase(ctrl)
			tt.mockInit(mockUC)

			handler := NewOrderHandler(mockUC)
			resp, err := handler.UpdateOrderStatusByPaymentID(context.Background(), tt.req)

			grpcErr := grpcutil.ToGRPCError(err)

			if tt.expectedCode == codes.OK {
				assert.NoError(t, grpcErr)
				assert.NotNil(t, resp)
			} else {
				st, _ := status.FromError(grpcErr)
				assert.Equal(t, tt.expectedCode, st.Code())
			}
		})
	}
}

func TestOrderHandler_PayForFriend(t *testing.T) {
	type mockInit func(m *mocks.MockOrderUseCase)

	tests := []struct {
		name         string
		req          *pb.PayForFriendRequest
		mockInit     mockInit
		expectedCode codes.Code
	}{
		{
			name: "Успешная оплата за друга",
			req: &pb.PayForFriendRequest{
				SplitId:         "split-123",
				PayerUserId:     1,
				PaymentMethodId: "pm-1",
				IdempotencyKey:  "idem-1",
			},
			mockInit: func(m *mocks.MockOrderUseCase) {
				m.EXPECT().PayForFriend(gomock.Any(), "split-123", int64(1), "pm-1", "idem-1").Return(nil)
			},
			expectedCode: codes.OK,
		},
		{
			name: "Ошибка оплаты за друга",
			req:  &pb.PayForFriendRequest{SplitId: "split-err"},
			mockInit: func(m *mocks.MockOrderUseCase) {
				m.EXPECT().PayForFriend(gomock.Any(), "split-err", int64(0), "", "").
					Return(errors.New("payment failed"))
			},
			expectedCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUC := mocks.NewMockOrderUseCase(ctrl)
			tt.mockInit(mockUC)

			handler := NewOrderHandler(mockUC)
			resp, err := handler.PayForFriend(context.Background(), tt.req)

			grpcErr := grpcutil.ToGRPCError(err)

			if tt.expectedCode == codes.OK {
				assert.NoError(t, grpcErr)
				assert.NotNil(t, resp)
			} else {
				st, _ := status.FromError(grpcErr)
				assert.Equal(t, tt.expectedCode, st.Code())
			}
		})
	}
}

func TestOrderHandler_CancelOrder(t *testing.T) {
	type mockInit func(m *mocks.MockOrderUseCase)

	tests := []struct {
		name         string
		req          *pb.CancelOrderRequest
		mockInit     mockInit
		expectedCode codes.Code
	}{
		{
			name: "Успешная отмена заказа",
			req:  &pb.CancelOrderRequest{OrderPublicId: "pub-1", UserId: 1},
			mockInit: func(m *mocks.MockOrderUseCase) {
				m.EXPECT().CancelOrder(gomock.Any(), "pub-1", int64(1)).Return(nil)
			},
			expectedCode: codes.OK,
		},
		{
			name: "Ошибка отмены заказа",
			req:  &pb.CancelOrderRequest{OrderPublicId: "pub-err", UserId: 1},
			mockInit: func(m *mocks.MockOrderUseCase) {
				m.EXPECT().CancelOrder(gomock.Any(), "pub-err", int64(1)).
					Return(errors.New("cannot cancel paid order"))
			},
			expectedCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUC := mocks.NewMockOrderUseCase(ctrl)
			tt.mockInit(mockUC)

			handler := NewOrderHandler(mockUC)
			resp, err := handler.CancelOrder(context.Background(), tt.req)

			grpcErr := grpcutil.ToGRPCError(err)

			if tt.expectedCode == codes.OK {
				assert.NoError(t, grpcErr)
				assert.NotNil(t, resp)
			} else {
				st, _ := status.FromError(grpcErr)
				assert.Equal(t, tt.expectedCode, st.Code())
			}
		})
	}
}
