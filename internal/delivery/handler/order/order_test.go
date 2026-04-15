package order

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
	ucMocks "github.com/go-park-mail-ru/2026_1_NaNcats/internal/usecase/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestOrderHandler_CreateOrder(t *testing.T) {
	type mockInit func(uc *ucMocks.MockOrderUseCase)

	tests := []struct {
		name           string
		userID         any
		body           any
		mockInit       mockInit
		expectedStatus int
	}{
		{
			name:   "Успешное создание заказа",
			userID: 1,
			body: CreateOrderRequest{
				AddressID:          "addr-uuid",
				RestaurantBranchID: 10,
				PaymentMethodID:    "pm-uuid",
			},
			mockInit: func(uc *ucMocks.MockOrderUseCase) {
				uc.EXPECT().
					CreateOrder(gomock.Any(), 1, gomock.Any()).
					Return("order-uuid", "http://yookassa.url", nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "Ошибка авторизации",
			userID: nil,
			body:   CreateOrderRequest{AddressID: "id", RestaurantBranchID: 1},
			mockInit: func(uc *ucMocks.MockOrderUseCase) {
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "Некорректный JSON",
			userID:         1,
			body:           "{invalid json",
			mockInit:       func(uc *ucMocks.MockOrderUseCase) {},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:   "Пустые обязательные поля",
			userID: 1,
			body:   CreateOrderRequest{AddressID: "", RestaurantBranchID: 0},
			mockInit: func(uc *ucMocks.MockOrderUseCase) {
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "Пустая корзина",
			userID: 1,
			body:   CreateOrderRequest{AddressID: "id", RestaurantBranchID: 1},
			mockInit: func(uc *ucMocks.MockOrderUseCase) {
				uc.EXPECT().
					CreateOrder(gomock.Any(), 1, gomock.Any()).
					Return("", "", domain.ErrCartIsEmpty)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "Адрес не найден",
			userID: 1,
			body:   CreateOrderRequest{AddressID: "id", RestaurantBranchID: 1},
			mockInit: func(uc *ucMocks.MockOrderUseCase) {
				uc.EXPECT().
					CreateOrder(gomock.Any(), 1, gomock.Any()).
					Return("", "", domain.ErrAddressNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:   "Внутренняя ошибка сервера",
			userID: 1,
			body:   CreateOrderRequest{AddressID: "id", RestaurantBranchID: 1},
			mockInit: func(uc *ucMocks.MockOrderUseCase) {
				uc.EXPECT().
					CreateOrder(gomock.Any(), 1, gomock.Any()).
					Return("", "", errors.New("db fail"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			uc := ucMocks.NewMockOrderUseCase(ctrl)
			l := domainMocks.NewNopLogger()
			h := NewOrderHandler(uc, l)

			var jsonBody []byte
			if s, ok := tt.body.(string); ok {
				jsonBody = []byte(s)
			} else {
				jsonBody, _ = json.Marshal(tt.body)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/orders", bytes.NewBuffer(jsonBody))
			if tt.userID != nil {
				ctx := context.WithValue(req.Context(), middleware.UserIDKey, tt.userID)
				req = req.WithContext(ctx)
			}

			w := httptest.NewRecorder()
			tt.mockInit(uc)

			h.CreateOrder(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestOrderHandler_GetMyOrders(t *testing.T) {
	type mockInit func(uc *ucMocks.MockOrderUseCase)

	tests := []struct {
		name           string
		userID         any
		mockInit       mockInit
		expectedStatus int
	}{
		{
			name:   "Успешное получение истории заказов",
			userID: 1,
			mockInit: func(uc *ucMocks.MockOrderUseCase) {
				orders := []domain.Order{
					{
						PublicID:        "pub-1",
						PaymentMethodID: "RestoName",
						TotalCost:       1000,
						Status:          "paid",
						CreatedAt:       time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					},
				}
				uc.EXPECT().GetOrders(gomock.Any(), 1).Return(orders, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Пользователь не авторизован",
			userID:         nil,
			mockInit:       func(uc *ucMocks.MockOrderUseCase) {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:   "Ошибка получения заказов",
			userID: 1,
			mockInit: func(uc *ucMocks.MockOrderUseCase) {
				uc.EXPECT().GetOrders(gomock.Any(), 1).Return(nil, errors.New("fail"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			uc := ucMocks.NewMockOrderUseCase(ctrl)
			l := domainMocks.NewNopLogger()
			h := NewOrderHandler(uc, l)

			req := httptest.NewRequest(http.MethodGet, "/api/orders", nil)
			if tt.userID != nil {
				ctx := context.WithValue(req.Context(), middleware.UserIDKey, tt.userID)
				req = req.WithContext(ctx)
			}

			w := httptest.NewRecorder()
			tt.mockInit(uc)

			h.GetMyOrders(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
