package usecase

// import (
// 	"context"
// 	"errors"
// 	"testing"

// 	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
// 	repoMocks "github.com/go-park-mail-ru/2026_1_NaNcats/internal/repository/mocks"
// 	"github.com/stretchr/testify/assert"
// 	"go.uber.org/mock/gomock"
// )

// func TestCartUseCase_UpdateCart(t *testing.T) {
// 	// Тип для инициализации моков
// 	type mockInit func(cr *repoMocks.MockCartRepository, dr *repoMocks.MockDishRepository)

// 	userID := 1
// 	restaurantID := 10

// 	// Данные для успешного кейса
// 	cartData := domain.Cart{
// 		RestaurantBrandID: restaurantID,
// 		Items: []domain.CartItem{
// 			{DishID: 1, Quantity: 2},
// 		},
// 	}

// 	tests := []struct {
// 		name     string
// 		userID   int
// 		cartData domain.Cart
// 		mockInit mockInit
// 		wantErr  error
// 	}{
// 		{
// 			name:     "Успешное обновление",
// 			userID:   userID,
// 			cartData: cartData,
// 			mockInit: func(cr *repoMocks.MockCartRepository, dr *repoMocks.MockDishRepository) {
// 				dr.EXPECT().
// 					GetDishesByIDs(gomock.Any(), []int{1}).
// 					Return([]domain.Dish{{ID: 1, RestaurantBrandID: restaurantID}}, nil)

// 				cr.EXPECT().
// 					UpdateCart(gomock.Any(), userID, restaurantID, cartData.Items).
// 					Return(nil)
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			name:     "Очистка корзины при пустом списке",
// 			userID:   userID,
// 			cartData: domain.Cart{Items: []domain.CartItem{}},
// 			mockInit: func(cr *repoMocks.MockCartRepository, dr *repoMocks.MockDishRepository) {
// 				cr.EXPECT().
// 					ClearCart(gomock.Any(), userID).
// 					Return(nil)
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			name:   "Ошибка: количество меньше или равно 0",
// 			userID: userID,
// 			cartData: domain.Cart{
// 				Items: []domain.CartItem{{DishID: 1, Quantity: 0}},
// 			},
// 			mockInit: func(cr *repoMocks.MockCartRepository, dr *repoMocks.MockDishRepository) {
// 			},
// 			wantErr: domain.ErrInvalidQuantity,
// 		},
// 		{
// 			name:     "Ошибка: товары из разных ресторанов",
// 			userID:   userID,
// 			cartData: cartData,
// 			mockInit: func(cr *repoMocks.MockCartRepository, dr *repoMocks.MockDishRepository) {
// 				dr.EXPECT().
// 					GetDishesByIDs(gomock.Any(), []int{1}).
// 					Return([]domain.Dish{{ID: 1, RestaurantBrandID: 99}}, nil)
// 			},
// 			wantErr: domain.ErrMultipleRestaurants,
// 		},
// 		{
// 			name:     "Ошибка: блюдо не найдено (база вернула меньше записей)",
// 			userID:   userID,
// 			cartData: cartData,
// 			mockInit: func(cr *repoMocks.MockCartRepository, dr *repoMocks.MockDishRepository) {
// 				dr.EXPECT().
// 					GetDishesByIDs(gomock.Any(), []int{1}).
// 					Return([]domain.Dish{}, nil)
// 			},
// 			wantErr: domain.ErrDishNotFound,
// 		},
// 		{
// 			name:     "Ошибка: сбой в работе репозитория блюд",
// 			userID:   userID,
// 			cartData: cartData,
// 			mockInit: func(cr *repoMocks.MockCartRepository, dr *repoMocks.MockDishRepository) {
// 				dr.EXPECT().
// 					GetDishesByIDs(gomock.Any(), []int{1}).
// 					Return(nil, errors.New("internal DB error"))
// 			},
// 			wantErr: errors.New("internal DB error"),
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			ctrl := gomock.NewController(t)
// 			defer ctrl.Finish()

// 			cr := repoMocks.NewMockCartRepository(ctrl)
// 			dr := repoMocks.NewMockDishRepository(ctrl)
// 			uc := NewCartUseCase(cr, dr, "default_url")

// 			tt.mockInit(cr, dr)

// 			err := uc.UpdateCart(context.Background(), tt.userID, tt.cartData)

// 			if tt.wantErr != nil {
// 				assert.Error(t, err)
// 				assert.Equal(t, tt.wantErr.Error(), err.Error())
// 			} else {
// 				assert.NoError(t, err)
// 			}
// 		})
// 	}
// }
