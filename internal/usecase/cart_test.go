package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	repoMocks "github.com/go-park-mail-ru/2026_1_NaNcats/internal/repository/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestCartUseCase_GetCart(t *testing.T) {
	type mockInit func(cr *repoMocks.MockCartRepository)

	tests := []struct {
		name      string
		userID    int
		mockInit  mockInit
		wantCart  domain.Cart
		wantTotal int64
		wantErr   error
	}{
		{
			name:   "Успешное получение и расчет стоимости",
			userID: 1,
			mockInit: func(cr *repoMocks.MockCartRepository) {
				cr.EXPECT().GetCartByUserID(gomock.Any(), 1).Return(domain.Cart{
					Items: []domain.CartItem{
						{DishID: 1, Price: 100, Quantity: 2},
						{DishID: 2, Price: 300, Quantity: 1},
					},
				}, nil)
			},
			wantCart: domain.Cart{
				Items: []domain.CartItem{
					{DishID: 1, Price: 100, Quantity: 2},
					{DishID: 2, Price: 300, Quantity: 1},
				},
			},
			wantTotal: 500,
			wantErr:   nil,
		},
		{
			name:   "Ошибка репозитория",
			userID: 1,
			mockInit: func(cr *repoMocks.MockCartRepository) {
				cr.EXPECT().GetCartByUserID(gomock.Any(), 1).Return(domain.Cart{}, errors.New("db error"))
			},
			wantCart:  domain.Cart{},
			wantTotal: 0,
			wantErr:   errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			cr := repoMocks.NewMockCartRepository(ctrl)
			dr := repoMocks.NewMockDishRepository(ctrl)
			uc := NewCartUseCase(cr, dr, "")

			tt.mockInit(cr)

			cart, total, err := uc.GetCart(context.Background(), tt.userID)

			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.wantCart, cart)
			assert.Equal(t, tt.wantTotal, total)
		})
	}
}

func TestCartUseCase_UpdateCart(t *testing.T) {
	type mockInit func(cr *repoMocks.MockCartRepository, dr *repoMocks.MockDishRepository)

	cartData := domain.Cart{
		RestaurantBrandID: 10,
		Items: []domain.CartItem{
			{DishID: 1, Quantity: 2},
		},
	}

	tests := []struct {
		name     string
		userID   int
		cartData domain.Cart
		mockInit mockInit
		wantErr  error
	}{
		{
			name:     "Успешное обновление",
			userID:   1,
			cartData: cartData,
			mockInit: func(cr *repoMocks.MockCartRepository, dr *repoMocks.MockDishRepository) {
				dr.EXPECT().GetDishByID(gomock.Any(), 1).Return(domain.Dish{RestaurantBrandID: 10}, nil)
				cr.EXPECT().UpdateCart(gomock.Any(), 1, 10, cartData.Items).Return(nil)
			},
			wantErr: nil,
		},
		{
			name:     "Очистка корзины при пустом списке",
			userID:   1,
			cartData: domain.Cart{Items: []domain.CartItem{}},
			mockInit: func(cr *repoMocks.MockCartRepository, dr *repoMocks.MockDishRepository) {
				cr.EXPECT().ClearCart(gomock.Any(), 1).Return(nil)
			},
			wantErr: nil,
		},
		{
			name:   "Ошибка: количество равно 0",
			userID: 1,
			cartData: domain.Cart{
				Items: []domain.CartItem{{DishID: 1, Quantity: 0}},
			},
			mockInit: func(cr *repoMocks.MockCartRepository, dr *repoMocks.MockDishRepository) {},
			wantErr:  domain.ErrInvalidQuantity,
		},
		{
			name:     "Ошибка: товары из разных ресторанов",
			userID:   1,
			cartData: cartData,
			mockInit: func(cr *repoMocks.MockCartRepository, dr *repoMocks.MockDishRepository) {
				dr.EXPECT().GetDishByID(gomock.Any(), 1).Return(domain.Dish{RestaurantBrandID: 99}, nil)
			},
			wantErr: domain.ErrMultipleRestaurants,
		},
		{
			name:     "Ошибка: блюдо не найдено",
			userID:   1,
			cartData: cartData,
			mockInit: func(cr *repoMocks.MockCartRepository, dr *repoMocks.MockDishRepository) {
				dr.EXPECT().GetDishByID(gomock.Any(), 1).Return(domain.Dish{}, errors.New("not found"))
			},
			wantErr: errors.New("not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			cr := repoMocks.NewMockCartRepository(ctrl)
			dr := repoMocks.NewMockDishRepository(ctrl)
			uc := NewCartUseCase(cr, dr, "")

			tt.mockInit(cr, dr)

			err := uc.UpdateCart(context.Background(), tt.userID, tt.cartData)

			assert.Equal(t, tt.wantErr, err)
		})
	}
}
