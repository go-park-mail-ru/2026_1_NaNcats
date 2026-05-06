package grpc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/cart/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/cart/internal/usecase/mocks"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/grpcutil"
	pb "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/cart"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestCartHandler_GetCart(t *testing.T) {
	type mockInit func(m *mocks.MockCartUseCase)

	userID := int64(42)
	now := time.Now()

	tests := []struct {
		name         string
		req          *pb.GetCartRequest
		mockInit     mockInit
		expectedCode codes.Code
	}{
		{
			name: "Успешное получение корзины",
			req:  &pb.GetCartRequest{UserId: userID},
			mockInit: func(m *mocks.MockCartUseCase) {
				m.EXPECT().GetCart(gomock.Any(), userID).
					Return(domain.Cart{
						ID:                "cart-uuid",
						AdminID:           userID,
						RestaurantBrandID: 1,
						Mode:              "solo",
						Status:            "active",
						Items: []domain.CartItem{
							{DishID: 10, Quantity: 2, Name: "Burger", Price: 500},
						},
						Members: []domain.CartMember{
							{UserID: userID, JoinedAt: now},
						},
					}, int64(1000), nil)
			},
			expectedCode: codes.OK,
		},
		{
			name: "Ошибка: корзина не найдена",
			req:  &pb.GetCartRequest{UserId: userID},
			mockInit: func(m *mocks.MockCartUseCase) {
				m.EXPECT().GetCart(gomock.Any(), userID).
					Return(domain.Cart{}, int64(0), errors.New("not found"))
			},
			expectedCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUC := mocks.NewMockCartUseCase(ctrl)
			tt.mockInit(mockUC)

			handler := NewCartHandler(mockUC)
			resp, err := handler.GetCart(context.Background(), tt.req)

			grpcErr := grpcutil.ToGRPCError(err)

			if tt.expectedCode == codes.OK {
				assert.NoError(t, grpcErr)
				require.NotNil(t, resp)
				assert.Equal(t, "cart-uuid", resp.Cart.CartId)
				assert.Equal(t, int64(1000), resp.TotalCost)
				assert.Equal(t, pb.CartMode_CART_MODE_SOLO, resp.Cart.Mode)
			} else {
				st, ok := status.FromError(grpcErr)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedCode, st.Code())
			}
		})
	}
}

func TestCartHandler_AddItem(t *testing.T) {
	type mockInit func(m *mocks.MockCartUseCase)

	req := &pb.AddItemRequest{
		CartId:         "uuid",
		UserId:         1,
		DishId:         10,
		Quantity:       2,
		IdempotencyKey: "key",
	}

	tests := []struct {
		name         string
		req          *pb.AddItemRequest
		mockInit     mockInit
		expectedCode codes.Code
	}{
		{
			name: "Успешное добавление",
			req:  req,
			mockInit: func(m *mocks.MockCartUseCase) {
				m.EXPECT().AddItem(gomock.Any(), req.CartId, req.UserId, req.DishId, req.Quantity, req.IdempotencyKey).
					Return(nil)
			},
			expectedCode: codes.OK,
		},
		{
			name: "Ошибка: корзина заблокирована",
			req:  req,
			mockInit: func(m *mocks.MockCartUseCase) {
				m.EXPECT().AddItem(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New("locked"))
			},
			expectedCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUC := mocks.NewMockCartUseCase(ctrl)
			tt.mockInit(mockUC)

			handler := NewCartHandler(mockUC)
			resp, err := handler.AddItem(context.Background(), tt.req)

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

func TestCartHandler_GenerateInvite(t *testing.T) {
	type mockInit func(m *mocks.MockCartUseCase)
	expires := time.Now().Add(time.Hour)

	tests := []struct {
		name         string
		req          *pb.GenerateInviteRequest
		mockInit     mockInit
		expectedCode codes.Code
	}{
		{
			name: "Успешная генерация инвайта",
			req:  &pb.GenerateInviteRequest{CartId: "uuid", UserId: 1},
			mockInit: func(m *mocks.MockCartUseCase) {
				m.EXPECT().GenerateInvite(gomock.Any(), "uuid", int64(1)).
					Return(domain.CartInvite{Token: "token123", ExpiresAt: expires}, nil)
			},
			expectedCode: codes.OK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUC := mocks.NewMockCartUseCase(ctrl)
			tt.mockInit(mockUC)

			handler := NewCartHandler(mockUC)
			resp, err := handler.GenerateInvite(context.Background(), tt.req)

			if tt.expectedCode == codes.OK {
				assert.NoError(t, err)
				assert.Equal(t, "token123", resp.Token)
				assert.Equal(t, expires.Format("2006-01-02T15:04:05Z07:00"), resp.ExpiresAt)
			}
		})
	}
}

func TestCartHandler_KickMember(t *testing.T) {
	type mockInit func(m *mocks.MockCartUseCase)

	req := &pb.CartMemberOperationRequest{
		CartId:         "uuid",
		AdminUserId:    1,
		TargetUserId:   2,
		IdempotencyKey: "idem",
	}

	tests := []struct {
		name         string
		req          *pb.CartMemberOperationRequest
		mockInit     mockInit
		expectedCode codes.Code
	}{
		{
			name: "Успешный кик",
			req:  req,
			mockInit: func(m *mocks.MockCartUseCase) {
				m.EXPECT().KickMember(gomock.Any(), req.CartId, req.AdminUserId, req.TargetUserId, req.IdempotencyKey).
					Return(nil)
			},
			expectedCode: codes.OK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUC := mocks.NewMockCartUseCase(ctrl)
			tt.mockInit(mockUC)

			handler := NewCartHandler(mockUC)
			_, err := handler.KickMember(context.Background(), tt.req)

			if tt.expectedCode == codes.OK {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCartHandler_UpdateItemQuantity(t *testing.T) {
	type mockInit func(m *mocks.MockCartUseCase)

	req := &pb.UpdateQuantityRequest{
		CartId:         "uuid",
		UserId:         1,
		DishId:         10,
		NewQuantity:    5,
		IdempotencyKey: "key",
	}

	tests := []struct {
		name         string
		req          *pb.UpdateQuantityRequest
		mockInit     mockInit
		expectedCode codes.Code
	}{
		{
			name: "Успешное обновление",
			req:  req,
			mockInit: func(m *mocks.MockCartUseCase) {
				m.EXPECT().UpdateItemQuantity(gomock.Any(), req.CartId, req.UserId, req.DishId, req.NewQuantity, req.IdempotencyKey).
					Return(nil)
			},
			expectedCode: codes.OK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUC := mocks.NewMockCartUseCase(ctrl)
			tt.mockInit(mockUC)

			handler := NewCartHandler(mockUC)
			_, err := handler.UpdateItemQuantity(context.Background(), tt.req)

			if tt.expectedCode == codes.OK {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCartHandler_ReassignItemOwner(t *testing.T) {
	type mockInit func(m *mocks.MockCartUseCase)

	newOwner := int64(2)
	req := &pb.ReassignOwnerRequest{
		CartId:         "uuid",
		AdminUserId:    1,
		DishId:         10,
		NewOwnerUserId: &newOwner,
		IdempotencyKey: "key",
	}

	tests := []struct {
		name         string
		req          *pb.ReassignOwnerRequest
		mockInit     mockInit
		expectedCode codes.Code
	}{
		{
			name: "Успешное переназначение",
			req:  req,
			mockInit: func(m *mocks.MockCartUseCase) {
				m.EXPECT().ReassignItemOwner(gomock.Any(), req.CartId, req.AdminUserId, req.DishId, req.NewOwnerUserId, req.IdempotencyKey).
					Return(nil)
			},
			expectedCode: codes.OK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUC := mocks.NewMockCartUseCase(ctrl)
			tt.mockInit(mockUC)

			handler := NewCartHandler(mockUC)
			_, err := handler.ReassignItemOwner(context.Background(), tt.req)

			if tt.expectedCode == codes.OK {
				assert.NoError(t, err)
			}
		})
	}
}
