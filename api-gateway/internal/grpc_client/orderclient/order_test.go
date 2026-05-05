package orderclient

import (
	"context"
	"errors"
	"testing"
	"time"

	pbOrder "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/order"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/order/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func ptr[T any](v T) *T {
	return &v
}

func TestOrderClient_CreateOrder(t *testing.T) {
	type mockBehavior func(m *mocks.MockOrderServiceClient)

	input := CreateOrderInput{
		AddressPublicID:    "addr-1",
		RestaurantBranchID: 10,
		RestaurantBrandID:  20,
		DeliveryCost:       150,
		ServiceFee:         20,
		PaymentMethodID:    "pm-1",
		PayForAll:          true,
		PayerMapping:       map[int64]int64{2: 1},
	}

	tests := []struct {
		name         string
		mockBehavior mockBehavior
		expectedRes  string
		expectedErr  error
		errContains  string
	}{
		{
			name: "Успешное создание заказа",
			mockBehavior: func(m *mocks.MockOrderServiceClient) {
				m.EXPECT().CreateOrder(gomock.Any(), &pbOrder.CreateOrderRequest{
					UserId:             1,
					AddressPublicId:    input.AddressPublicID,
					RestaurantBranchId: input.RestaurantBranchID,
					RestaurantBrandId:  input.RestaurantBrandID,
					PaymentMethodId:    input.PaymentMethodID,
					PayForAll:          input.PayForAll,
					PayerMapping:       input.PayerMapping,
					DeliveryCost:       input.DeliveryCost,
					ServiceFee:         input.ServiceFee,
					IdempotencyKey:     "idem-1",
				}).Return(&pbOrder.CreateOrderResponse{
					OrderPublicId: "order-pub-123",
				}, nil)
			},
			expectedRes: "order-pub-123",
			expectedErr: nil,
		},
		{
			name: "Ошибка: Адрес не найден",
			mockBehavior: func(m *mocks.MockOrderServiceClient) {
				m.EXPECT().CreateOrder(gomock.Any(), gomock.Any()).
					Return(nil, status.Error(codes.NotFound, "address missing"))
			},
			expectedRes: "",
			expectedErr: ErrAddressNotFound,
		},
		{
			name: "Ошибка: Нераспределенные позиции (FailedPrecondition)",
			mockBehavior: func(m *mocks.MockOrderServiceClient) {
				m.EXPECT().CreateOrder(gomock.Any(), gomock.Any()).
					Return(nil, status.Error(codes.FailedPrecondition, "unassigned items"))
			},
			expectedRes: "",
			expectedErr: ErrUnassignedItems,
		},
		{
			name: "Ошибка: Пустая корзина (InvalidArgument)",
			mockBehavior: func(m *mocks.MockOrderServiceClient) {
				m.EXPECT().CreateOrder(gomock.Any(), gomock.Any()).
					Return(nil, status.Error(codes.InvalidArgument, "cart empty"))
			},
			expectedRes: "",
			expectedErr: ErrCartIsEmpty,
		},
		{
			name: "Другая gRPC ошибка с сохранением сообщения",
			mockBehavior: func(m *mocks.MockOrderServiceClient) {
				m.EXPECT().CreateOrder(gomock.Any(), gomock.Any()).
					Return(nil, status.Error(codes.Aborted, "payment rejected"))
			},
			expectedRes: "",
			expectedErr: nil,
			errContains: "order service error: payment rejected",
		},
		{
			name: "Обычная ошибка (не gRPC)",
			mockBehavior: func(m *mocks.MockOrderServiceClient) {
				m.EXPECT().CreateOrder(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("network error"))
			},
			expectedRes: "",
			expectedErr: ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockGRPC := mocks.NewMockOrderServiceClient(ctrl)
			tt.mockBehavior(mockGRPC)

			client := NewOrderClient(mockGRPC)
			res, err := client.CreateOrder(context.Background(), 1, input, "idem-1")

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else if tt.errContains != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedRes, res)
			}
		})
	}
}

func TestOrderClient_GetOrders(t *testing.T) {
	type mockBehavior func(m *mocks.MockOrderServiceClient)

	now := time.Now()
	pbNow := timestamppb.New(now)

	tests := []struct {
		name         string
		userID       int64
		mockBehavior mockBehavior
		expectedRes  []Order
		expectedErr  error
	}{
		{
			name:   "Успешное получение заказов",
			userID: 1,
			mockBehavior: func(m *mocks.MockOrderServiceClient) {
				m.EXPECT().GetOrders(gomock.Any(), &pbOrder.GetOrdersRequest{
					UserId: 1,
				}).Return(&pbOrder.GetOrdersResponse{
					Orders: []*pbOrder.Order{
						{
							PublicId:          "pub-1",
							RestaurantName:    "KFC",
							RestaurantLogoUrl: "url",
							TotalCost:         1000,
							Status:            "paid",
							CreatedAt:         pbNow,
							Items: []*pbOrder.OrderDish{
								{DishId: 10, Quantity: 2, Price: 500, OwnerUserId: ptr(int64(1))},
							},
							Splits: []*pbOrder.OrderSplit{
								{SplitId: "split-1", UserId: 1, Amount: 1000, Status: "paid"},
							},
						},
					},
				}, nil)
			},
			expectedRes: []Order{
				{
					PublicID:          "pub-1",
					RestaurantName:    "KFC",
					RestaurantLogoURL: "url",
					TotalCost:         1000,
					Status:            "paid",
					CreatedAt:         pbNow.AsTime(),
					Items: []OrderDish{
						{DishID: 10, Quantity: 2, Price: 500, OwnerUserID: ptr(int64(1))},
					},
					Splits: []OrderSplit{
						{SplitID: "split-1", UserID: 1, Amount: 1000, Status: "paid"},
					},
				},
			},
			expectedErr: nil,
		},
		{
			name:   "Ошибка получения",
			userID: 1,
			mockBehavior: func(m *mocks.MockOrderServiceClient) {
				m.EXPECT().GetOrders(gomock.Any(), gomock.Any()).
					Return(nil, status.Error(codes.Internal, "db error"))
			},
			expectedRes: nil,
			expectedErr: ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockGRPC := mocks.NewMockOrderServiceClient(ctrl)
			tt.mockBehavior(mockGRPC)

			client := NewOrderClient(mockGRPC)
			res, err := client.GetOrders(context.Background(), tt.userID)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedRes, res)
			}
		})
	}
}

func TestOrderClient_PayForFriend(t *testing.T) {
	type mockBehavior func(m *mocks.MockOrderServiceClient)

	tests := []struct {
		name         string
		mockBehavior mockBehavior
		expectedErr  error
	}{
		{
			name: "Успешная оплата",
			mockBehavior: func(m *mocks.MockOrderServiceClient) {
				m.EXPECT().PayForFriend(gomock.Any(), &pbOrder.PayForFriendRequest{
					SplitId:         "split-1",
					PayerUserId:     1,
					PaymentMethodId: "pm-1",
					IdempotencyKey:  "idem-1",
				}).Return(&pbOrder.PayForFriendResponse{ConfirmationUrl: "url"}, nil)
			},
			expectedErr: nil,
		},
		{
			name: "Ошибка при оплате",
			mockBehavior: func(m *mocks.MockOrderServiceClient) {
				m.EXPECT().PayForFriend(gomock.Any(), gomock.Any()).
					Return(nil, status.Error(codes.Internal, "error"))
			},
			expectedErr: ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockGRPC := mocks.NewMockOrderServiceClient(ctrl)
			tt.mockBehavior(mockGRPC)

			client := NewOrderClient(mockGRPC)
			err := client.PayForFriend(context.Background(), "split-1", 1, "pm-1", "idem-1")

			assert.ErrorIs(t, err, tt.expectedErr)
		})
	}
}

func TestOrderClient_GetOrderPaymentID(t *testing.T) {
	type mockBehavior func(m *mocks.MockOrderServiceClient)

	tests := []struct {
		name         string
		mockBehavior mockBehavior
		expectedRes  string
		expectedErr  error
		errContains  string
	}{
		{
			name: "Успешное получение Payment ID",
			mockBehavior: func(m *mocks.MockOrderServiceClient) {
				m.EXPECT().GetOrderPaymentID(gomock.Any(), &pbOrder.GetOrderPaymentIDRequest{
					OrderPublicId: "pub-1",
					UserId:        1,
				}).Return(&pbOrder.GetOrderPaymentIDResponse{YookassaPaymentId: "pay-123"}, nil)
			},
			expectedRes: "pay-123",
			expectedErr: nil,
		},
		{
			name: "Заказ не найден",
			mockBehavior: func(m *mocks.MockOrderServiceClient) {
				m.EXPECT().GetOrderPaymentID(gomock.Any(), gomock.Any()).
					Return(nil, status.Error(codes.NotFound, "not found"))
			},
			expectedErr: ErrAddressNotFound,
		},
		{
			name: "Отказ в доступе (PermissionDenied)",
			mockBehavior: func(m *mocks.MockOrderServiceClient) {
				m.EXPECT().GetOrderPaymentID(gomock.Any(), gomock.Any()).
					Return(nil, status.Error(codes.PermissionDenied, "forbidden"))
			},
			expectedErr: ErrInternal,
		},
		{
			name: "Платеж еще не готов (FailedPrecondition)",
			mockBehavior: func(m *mocks.MockOrderServiceClient) {
				m.EXPECT().GetOrderPaymentID(gomock.Any(), gomock.Any()).
					Return(nil, status.Error(codes.FailedPrecondition, "pending"))
			},
			errContains: "payment not ready: pending",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockGRPC := mocks.NewMockOrderServiceClient(ctrl)
			tt.mockBehavior(mockGRPC)

			client := NewOrderClient(mockGRPC)
			res, err := client.GetOrderPaymentID(context.Background(), "pub-1", 1)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else if tt.errContains != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedRes, res)
			}
		})
	}
}

func TestOrderClient_CancelOrder(t *testing.T) {
	type mockBehavior func(m *mocks.MockOrderServiceClient)

	tests := []struct {
		name         string
		mockBehavior mockBehavior
		expectedErr  error
		errContains  string
	}{
		{
			name: "Успешная отмена заказа",
			mockBehavior: func(m *mocks.MockOrderServiceClient) {
				m.EXPECT().CancelOrder(gomock.Any(), &pbOrder.CancelOrderRequest{
					OrderPublicId: "pub-1",
					UserId:        1,
				}).Return(&emptypb.Empty{}, nil)
			},
			expectedErr: nil,
		},
		{
			name: "Заказ не найден (NotFound)",
			mockBehavior: func(m *mocks.MockOrderServiceClient) {
				m.EXPECT().CancelOrder(gomock.Any(), gomock.Any()).
					Return(nil, status.Error(codes.NotFound, "not found"))
			},
			expectedErr: ErrAddressNotFound,
		},
		{
			name: "Нельзя отменить (FailedPrecondition)",
			mockBehavior: func(m *mocks.MockOrderServiceClient) {
				m.EXPECT().CancelOrder(gomock.Any(), gomock.Any()).
					Return(nil, status.Error(codes.FailedPrecondition, "already paid"))
			},
			errContains: "cannot cancel: already paid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockGRPC := mocks.NewMockOrderServiceClient(ctrl)
			tt.mockBehavior(mockGRPC)

			client := NewOrderClient(mockGRPC)
			err := client.CancelOrder(context.Background(), "pub-1", 1)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else if tt.errContains != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
