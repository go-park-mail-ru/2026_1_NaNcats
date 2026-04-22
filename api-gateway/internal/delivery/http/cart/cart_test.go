package cart

/*
import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/delivery/middleware"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	domainMocks "github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain/mocks"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/usecase/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestCartHandler_GetCart(t *testing.T) {
	type mockInit func(uc *mocks.MockCartUseCase)

	tests := []struct {
		name           string
		userID         any
		mockInit       mockInit
		expectedStatus int
	}{
		{
			name:   "Успешное получение корзины",
			userID: 1,
			mockInit: func(uc *mocks.MockCartUseCase) {
				cart := domain.Cart{
					RestaurantBrandID: 10,
					UpdatedAt:         time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
					Items: []domain.CartItem{
						{DishID: 1, Name: "Блюдо", Price: 100, Quantity: 2, ImageURL: "img.png"},
					},
				}
				uc.EXPECT().GetCart(gomock.Any(), 1).Return(cart, int64(200), nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Ошибка получения ID пользователя из контекста",
			userID:         nil,
			mockInit:       func(uc *mocks.MockCartUseCase) {},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			uc := mocks.NewMockCartUseCase(ctrl)
			l := domainMocks.NewNopLogger()
			h := NewCartHandler(uc, l)

			req := httptest.NewRequest(http.MethodGet, "/api/cart", nil)
			if tt.userID != nil {
				ctx := context.WithValue(req.Context(), middleware.UserIDKey, tt.userID)
				req = req.WithContext(ctx)
			}

			w := httptest.NewRecorder()
			tt.mockInit(uc)

			h.GetCart(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestCartHandler_UpdateCart(t *testing.T) {
	type mockInit func(uc *mocks.MockCartUseCase)

	tests := []struct {
		name           string
		userID         any
		body           any
		mockInit       mockInit
		expectedStatus int
	}{
		{
			name:   "Успешное обновление корзины",
			userID: 1,
			body: CartRequest{
				RestaurantID: 10,
				Items:        []CartItemDTO{{DishID: 1, Quantity: 2}},
			},
			mockInit: func(uc *mocks.MockCartUseCase) {
				uc.EXPECT().UpdateCart(gomock.Any(), 1, gomock.Any()).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "Ошибка: товары из разных ресторанов",
			userID: 1,
			body: CartRequest{
				RestaurantID: 10,
				Items:        []CartItemDTO{{DishID: 1, Quantity: 2}},
			},
			mockInit: func(uc *mocks.MockCartUseCase) {
				uc.EXPECT().UpdateCart(gomock.Any(), 1, gomock.Any()).Return(domain.ErrMultipleRestaurants)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "Ошибка: некорректное количество товара",
			userID: 1,
			body: CartRequest{
				RestaurantID: 10,
				Items:        []CartItemDTO{{DishID: 1, Quantity: -1}},
			},
			mockInit: func(uc *mocks.MockCartUseCase) {
				uc.EXPECT().UpdateCart(gomock.Any(), 1, gomock.Any()).Return(domain.ErrInvalidQuantity)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Ошибка декодирования тела запроса",
			userID:         1,
			body:           "invalid-json",
			mockInit:       func(uc *mocks.MockCartUseCase) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "Внутренняя ошибка сервера при обновлении",
			userID: 1,
			body: CartRequest{
				RestaurantID: 10,
				Items:        []CartItemDTO{{DishID: 1, Quantity: 2}},
			},
			mockInit: func(uc *mocks.MockCartUseCase) {
				uc.EXPECT().UpdateCart(gomock.Any(), 1, gomock.Any()).Return(errors.New("internal error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			uc := mocks.NewMockCartUseCase(ctrl)
			l := domainMocks.NewNopLogger()
			h := NewCartHandler(uc, l)

			var jsonBody []byte
			if s, ok := tt.body.(string); ok {
				jsonBody = []byte(s)
			} else {
				jsonBody, _ = json.Marshal(tt.body)
			}

			req := httptest.NewRequest(http.MethodPut, "/api/cart", bytes.NewBuffer(jsonBody))
			if tt.userID != nil {
				ctx := context.WithValue(req.Context(), middleware.UserIDKey, tt.userID)
				req = req.WithContext(ctx)
			}

			w := httptest.NewRecorder()
			tt.mockInit(uc)

			h.UpdateCart(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
*/
