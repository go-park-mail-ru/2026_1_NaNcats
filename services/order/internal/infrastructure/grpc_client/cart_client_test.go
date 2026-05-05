package grpc_client

import (
	"context"
	"errors"
	"testing"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/order/internal/domain"
	pb "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/cart"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/cart/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc"
)

//go:generate mockgen -destination=../../../../../shared/proto/cart/mocks/cart_client_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/cart CartServiceClient

func TestCartClient_GetCart(t *testing.T) {
	type mockBehavior func(mock *mocks.MockCartServiceClient, userID int64)

	ownerID := int64(99)

	tests := []struct {
		name         string
		userID       int64
		mockBehavior mockBehavior
		expectedCart domain.Cart
		expectedCost int64
		expectedErr  error
	}{
		{
			name:   "Успешное получение корзины с маппингом",
			userID: 1,
			mockBehavior: func(m *mocks.MockCartServiceClient, userID int64) {
				resp := &pb.GetCartResponse{
					Cart: &pb.Cart{
						RestaurantBrandId: 42,
						Items: []*pb.CartItem{
							{DishId: 10, Quantity: 2, Price: 500, OwnerUserId: &ownerID},
							{DishId: 11, Quantity: 1, Price: 1500},
						},
					},
					TotalCost: 2500,
				}

				// Используем gomock.Any() для protobuf и проверяем поля вручную
				m.EXPECT().GetCart(gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, req *pb.GetCartRequest, opts ...grpc.CallOption) (*pb.GetCartResponse, error) {
						assert.Equal(t, userID, req.UserId)
						return resp, nil
					},
				)
			},
			expectedCart: domain.Cart{
				RestaurantBrandID: 42,
				Items: []domain.CartItem{
					{DishID: 10, Quantity: 2, Price: 500, OwnerUserID: &ownerID},
					{DishID: 11, Quantity: 1, Price: 1500, OwnerUserID: nil},
				},
			},
			expectedCost: 2500,
			expectedErr:  nil,
		},
		{
			name:   "Ошибка получения корзины из gRPC",
			userID: 1,
			mockBehavior: func(m *mocks.MockCartServiceClient, userID int64) {
				m.EXPECT().GetCart(gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, req *pb.GetCartRequest, opts ...grpc.CallOption) (*pb.GetCartResponse, error) {
						assert.Equal(t, userID, req.UserId)
						return nil, errors.New("grpc connection failed")
					},
				)
			},
			expectedCart: domain.Cart{},
			expectedCost: 0,
			expectedErr:  errors.New("grpc connection failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClient := mocks.NewMockCartServiceClient(ctrl)
			tt.mockBehavior(mockClient, tt.userID)

			client := NewCartClient(mockClient)
			cart, cost, err := client.GetCart(context.Background(), tt.userID)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedCart, cart)
				assert.Equal(t, tt.expectedCost, cost)
			}
		})
	}
}
