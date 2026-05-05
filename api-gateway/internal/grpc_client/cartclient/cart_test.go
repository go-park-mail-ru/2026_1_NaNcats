package cartclient

import (
	"context"
	"errors"
	"testing"

	pbCart "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/cart"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/cart/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func ptr[T any](v T) *T {
	return &v
}

func TestCartClient_GetCart(t *testing.T) {
	type mockBehavior func(m *mocks.MockCartServiceClient)

	tests := []struct {
		name         string
		userID       int64
		mockBehavior mockBehavior
		expectedRes  *Cart
		expectedErr  error
	}{
		{
			name:   "Успешное получение корзины (Shared + Locked)",
			userID: 1,
			mockBehavior: func(m *mocks.MockCartServiceClient) {
				m.EXPECT().GetCart(gomock.Any(), &pbCart.GetCartRequest{UserId: 1}).
					Return(&pbCart.GetCartResponse{
						TotalCost: 1500,
						Cart: &pbCart.Cart{
							CartId:            "cart-uuid",
							AdminId:           1,
							RestaurantBrandId: 10,
							Mode:              pbCart.CartMode_CART_MODE_SHARED,
							Status:            pbCart.CartStatus_CART_STATUS_LOCKED,
							Items: []*pbCart.CartItem{
								{DishId: 100, Quantity: 2, Name: "Burger", Price: 500, ImageUrl: "url", OwnerUserId: ptr(int64(2))},
							},
							Members: []*pbCart.CartMember{
								{UserId: 1, JoinedAt: "2023-10-10T10:00:00Z"},
								{UserId: 2, JoinedAt: "2023-10-10T10:05:00Z"},
							},
						},
					}, nil)
			},
			expectedRes: &Cart{
				ID:                "cart-uuid",
				AdminID:           1,
				RestaurantBrandID: 10,
				Mode:              "shared",
				Status:            "locked",
				TotalCost:         1500,
				Items: []Item{
					{DishID: 100, Quantity: 2, Name: "Burger", Price: 500, ImageURL: "url", OwnerUserID: ptr(int64(2))},
				},
				Members: []Member{
					{UserID: 1, JoinedAt: "2023-10-10T10:00:00Z"},
					{UserID: 2, JoinedAt: "2023-10-10T10:05:00Z"},
				},
			},
			expectedErr: nil,
		},
		{
			name:   "Успешное получение корзины (Solo + Active)",
			userID: 2,
			mockBehavior: func(m *mocks.MockCartServiceClient) {
				m.EXPECT().GetCart(gomock.Any(), &pbCart.GetCartRequest{UserId: 2}).
					Return(&pbCart.GetCartResponse{
						TotalCost: 0,
						Cart: &pbCart.Cart{
							CartId: "cart-solo",
							Mode:   pbCart.CartMode_CART_MODE_SOLO,
							Status: pbCart.CartStatus_CART_STATUS_ACTIVE,
						},
					}, nil)
			},
			expectedRes: &Cart{
				ID:      "cart-solo",
				Mode:    "solo",
				Status:  "active",
				Items:   []Item{},
				Members: []Member{},
			},
			expectedErr: nil,
		},
		{
			name:   "Ошибка получения корзины (NotFound)",
			userID: 3,
			mockBehavior: func(m *mocks.MockCartServiceClient) {
				m.EXPECT().GetCart(gomock.Any(), gomock.Any()).
					Return(nil, status.Error(codes.NotFound, "cart not found"))
			},
			expectedRes: nil,
			expectedErr: ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockGRPCClient := mocks.NewMockCartServiceClient(ctrl)
			tt.mockBehavior(mockGRPCClient)

			client := NewCartClient(mockGRPCClient)
			res, err := client.GetCart(context.Background(), tt.userID)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
				assert.Nil(t, res)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedRes, res)
			}
		})
	}
}

func TestCartClient_ErrorMapping(t *testing.T) {
	type mockBehavior func(m *mocks.MockCartServiceClient)

	tests := []struct {
		name         string
		mockBehavior mockBehavior
		expectedErr  error
	}{
		{
			name: "Ошибка: MULTIPLE_RESTAURANTS",
			mockBehavior: func(m *mocks.MockCartServiceClient) {
				m.EXPECT().AddItem(gomock.Any(), gomock.Any()).
					Return(nil, status.Error(codes.InvalidArgument, "MULTIPLE_RESTAURANTS"))
			},
			expectedErr: ErrMultipleRestaurants,
		},
		{
			name: "Ошибка: INVALID_QUANTITY",
			mockBehavior: func(m *mocks.MockCartServiceClient) {
				m.EXPECT().AddItem(gomock.Any(), gomock.Any()).
					Return(nil, status.Error(codes.InvalidArgument, "INVALID_QUANTITY"))
			},
			expectedErr: ErrInvalidQuantity,
		},
		{
			name: "Ошибка: InvalidArgument (дефолтная)",
			mockBehavior: func(m *mocks.MockCartServiceClient) {
				m.EXPECT().AddItem(gomock.Any(), gomock.Any()).
					Return(nil, status.Error(codes.InvalidArgument, "SOME_OTHER_ERROR"))
			},
			expectedErr: ErrInvalidCart,
		},
		{
			name: "Ошибка: CART_LOCKED",
			mockBehavior: func(m *mocks.MockCartServiceClient) {
				m.EXPECT().AddItem(gomock.Any(), gomock.Any()).
					Return(nil, status.Error(codes.FailedPrecondition, "CART_LOCKED"))
			},
			expectedErr: ErrCartLocked,
		},
		{
			name: "Ошибка: Forbidden",
			mockBehavior: func(m *mocks.MockCartServiceClient) {
				m.EXPECT().AddItem(gomock.Any(), gomock.Any()).
					Return(nil, status.Error(codes.PermissionDenied, "forbidden"))
			},
			expectedErr: ErrForbidden,
		},
		{
			name: "Ошибка: Обычная Go ошибка без gRPC статуса",
			mockBehavior: func(m *mocks.MockCartServiceClient) {
				m.EXPECT().AddItem(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("connection refused"))
			},
			expectedErr: ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockGRPCClient := mocks.NewMockCartServiceClient(ctrl)
			tt.mockBehavior(mockGRPCClient)

			client := NewCartClient(mockGRPCClient)
			err := client.AddItem(context.Background(), "cart-1", 1, 10, 1, "idem")

			assert.ErrorIs(t, err, tt.expectedErr)
		})
	}
}

func TestCartClient_LockCart(t *testing.T) {
	type mockBehavior func(m *mocks.MockCartServiceClient)

	mapping := map[int64]int64{2: 1}

	tests := []struct {
		name         string
		mockBehavior mockBehavior
		expectedErr  error
	}{
		{
			name: "Успешная блокировка",
			mockBehavior: func(m *mocks.MockCartServiceClient) {
				m.EXPECT().LockCart(gomock.Any(), &pbCart.LockCartRequest{
					CartId:         "cart-1",
					UserId:         1,
					PayForAll:      true,
					PayerMapping:   mapping,
					IdempotencyKey: "idem-1",
				}).Return(&emptypb.Empty{}, nil)
			},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockGRPCClient := mocks.NewMockCartServiceClient(ctrl)
			tt.mockBehavior(mockGRPCClient)

			client := NewCartClient(mockGRPCClient)
			err := client.LockCart(context.Background(), "cart-1", 1, true, mapping, "idem-1")

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCartClient_GenerateInvite(t *testing.T) {
	type mockBehavior func(m *mocks.MockCartServiceClient)

	tests := []struct {
		name         string
		mockBehavior mockBehavior
		expectedRes  *InviteResponse
		expectedErr  error
	}{
		{
			name: "Успешная генерация инвайта",
			mockBehavior: func(m *mocks.MockCartServiceClient) {
				m.EXPECT().GenerateInvite(gomock.Any(), &pbCart.GenerateInviteRequest{
					CartId: "cart-1",
					UserId: 1,
				}).Return(&pbCart.GenerateInviteResponse{
					Token:     "token-hash",
					ExpiresAt: "2023-12-31T23:59:59Z",
				}, nil)
			},
			expectedRes: &InviteResponse{
				Token:     "token-hash",
				ExpiresAt: "2023-12-31T23:59:59Z",
			},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockGRPCClient := mocks.NewMockCartServiceClient(ctrl)
			tt.mockBehavior(mockGRPCClient)

			client := NewCartClient(mockGRPCClient)
			res, err := client.GenerateInvite(context.Background(), "cart-1", 1)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
				assert.Nil(t, res)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedRes, res)
			}
		})
	}
}

func TestCartClient_ReassignItemOwner(t *testing.T) {
	type mockBehavior func(m *mocks.MockCartServiceClient)

	newOwner := ptr(int64(3))

	tests := []struct {
		name         string
		mockBehavior mockBehavior
		expectedErr  error
	}{
		{
			name: "Успешное переназначение позиции",
			mockBehavior: func(m *mocks.MockCartServiceClient) {
				m.EXPECT().ReassignItemOwner(gomock.Any(), &pbCart.ReassignOwnerRequest{
					CartId:         "cart-1",
					AdminUserId:    1,
					DishId:         100,
					NewOwnerUserId: newOwner,
					IdempotencyKey: "idem",
				}).Return(&emptypb.Empty{}, nil)
			},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockGRPCClient := mocks.NewMockCartServiceClient(ctrl)
			tt.mockBehavior(mockGRPCClient)

			client := NewCartClient(mockGRPCClient)
			err := client.ReassignItemOwner(context.Background(), "cart-1", 1, 100, newOwner, "idem")

			assert.ErrorIs(t, err, tt.expectedErr)
		})
	}
}
