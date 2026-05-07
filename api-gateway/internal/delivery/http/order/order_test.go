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

	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/grpc_client/orderclient"
	orderMocks "github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/grpc_client/orderclient/mocks"
	paymentMocks "github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/grpc_client/paymentclient/mocks"
	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/grpc_client/restaurantclient"
	restaurantMocks "github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/grpc_client/restaurantclient/mocks"
	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/middleware"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// Вспомогательная функция для инжекта UserID в контекст
func withUserIDContext(req *http.Request, userID int64) *http.Request {
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
	return req.WithContext(ctx)
}

func setupTestHandler(ctrl *gomock.Controller) (*OrderHandler, *orderMocks.MockOrderClient, *paymentMocks.MockPaymentClient, *restaurantMocks.MockRestaurantClient) {
	mockOrder := orderMocks.NewMockOrderClient(ctrl)
	mockPayment := paymentMocks.NewMockPaymentClient(ctrl)
	mockRestaurant := restaurantMocks.NewMockRestaurantClient(ctrl)
	log := logger.NewNopLogger()

	// wsManager передаем nil, так как полноценно протестировать WS
	// через httptest без поднятия реального сервера сложно и выходит за рамки unit-тестов
	handler := NewOrderHandler(mockOrder, mockPayment, mockRestaurant, nil, log)
	return handler, mockOrder, mockPayment, mockRestaurant
}

func TestOrderHandler_CancelOrder(t *testing.T) {
	type mockBehavior func(order *orderMocks.MockOrderClient)

	tests := []struct {
		name           string
		userID         int64
		orderID        string
		withAuth       bool
		mockBehavior   mockBehavior
		expectedStatus int
	}{
		{
			name:     "Успешная отмена заказа",
			userID:   1,
			orderID:  "ord-123",
			withAuth: true,
			mockBehavior: func(order *orderMocks.MockOrderClient) {
				order.EXPECT().CancelOrder(gomock.Any(), "ord-123", int64(1)).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Ошибка: неавторизован",
			orderID:        "ord-123",
			withAuth:       false,
			mockBehavior:   func(order *orderMocks.MockOrderClient) {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Ошибка: пустой order id",
			userID:         1,
			orderID:        "",
			withAuth:       true,
			mockBehavior:   func(order *orderMocks.MockOrderClient) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:     "Ошибка: отказ со стороны grpc клиента",
			userID:   1,
			orderID:  "ord-123",
			withAuth: true,
			mockBehavior: func(order *orderMocks.MockOrderClient) {
				order.EXPECT().CancelOrder(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New("cannot cancel"))
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			handler, mockOrder, _, _ := setupTestHandler(ctrl)
			tt.mockBehavior(mockOrder)

			req := httptest.NewRequest(http.MethodPost, "/orders/"+tt.orderID+"/cancel", nil)
			if tt.withAuth {
				req = withUserIDContext(req, tt.userID)
			}
			req.SetPathValue("id", tt.orderID)

			w := httptest.NewRecorder()
			handler.CancelOrder(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestOrderHandler_CreateOrder(t *testing.T) {
	type mockBehavior func(order *orderMocks.MockOrderClient)

	validReq := CreateOrderRequest{
		AddressID:          "addr-1",
		RestaurantBranchID: 10,
		RestaurantBrandID:  20,
	}

	tests := []struct {
		name           string
		reqBody        interface{}
		headers        map[string]string
		withAuth       bool
		mockBehavior   mockBehavior
		expectedStatus int
	}{
		{
			name:     "Успешное создание заказа",
			reqBody:  validReq,
			headers:  map[string]string{"Idempotency-Key": "idem-123"},
			withAuth: true,
			mockBehavior: func(order *orderMocks.MockOrderClient) {
				order.EXPECT().CreateOrder(gomock.Any(), int64(1), gomock.Any(), "idem-123").
					Return("order-uuid-123", nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:     "Ошибка: пустая корзина",
			reqBody:  validReq,
			headers:  map[string]string{"Idempotency-Key": "idem-123"},
			withAuth: true,
			mockBehavior: func(order *orderMocks.MockOrderClient) {
				order.EXPECT().CreateOrder(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return("", orderclient.ErrCartIsEmpty)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:     "Ошибка: адрес не найден",
			reqBody:  validReq,
			headers:  map[string]string{"Idempotency-Key": "idem-123"},
			withAuth: true,
			mockBehavior: func(order *orderMocks.MockOrderClient) {
				order.EXPECT().CreateOrder(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return("", orderclient.ErrAddressNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Ошибка: нет Idempotency-Key",
			reqBody:        validReq,
			headers:        map[string]string{},
			withAuth:       true,
			mockBehavior:   func(order *orderMocks.MockOrderClient) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			handler, mockOrder, _, _ := setupTestHandler(ctrl)
			tt.mockBehavior(mockOrder)

			body, _ := json.Marshal(tt.reqBody)
			req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewBuffer(body))
			if tt.withAuth {
				req = withUserIDContext(req, 1)
			}
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			w := httptest.NewRecorder()
			handler.CreateOrder(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestOrderHandler_GetMyOrders(t *testing.T) {
	type mockBehavior func(order *orderMocks.MockOrderClient, rest *restaurantMocks.MockRestaurantClient)

	now := time.Now()

	tests := []struct {
		name           string
		userID         int64
		withAuth       bool
		mockBehavior   mockBehavior
		expectedStatus int
	}{
		{
			name:     "Успешное получение истории заказов (с блюдами)",
			userID:   1,
			withAuth: true,
			mockBehavior: func(order *orderMocks.MockOrderClient, rest *restaurantMocks.MockRestaurantClient) {
				order.EXPECT().GetOrders(gomock.Any(), int64(1)).Return([]orderclient.Order{
					{
						PublicID:  "ord-1",
						Status:    "paid",
						CreatedAt: now,
						Items: []orderclient.OrderDish{
							{DishID: 100, Quantity: 2},
						},
					},
				}, nil)

				// Проверка обогащения через restaurantClient (теперь используем доменную модель Dish)
				rest.EXPECT().GetDishesByIDs(gomock.Any(), []int64{100}).Return([]restaurantclient.Dish{
					{ID: 100, Name: "Burger", ImageURL: "img.png"},
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:     "Успех, даже если restaurant_service упал (fallback без имен)",
			userID:   1,
			withAuth: true,
			mockBehavior: func(order *orderMocks.MockOrderClient, rest *restaurantMocks.MockRestaurantClient) {
				order.EXPECT().GetOrders(gomock.Any(), int64(1)).Return([]orderclient.Order{
					{PublicID: "ord-1", Items: []orderclient.OrderDish{{DishID: 100}}},
				}, nil)

				rest.EXPECT().GetDishesByIDs(gomock.Any(), []int64{100}).
					Return(nil, errors.New("restaurant service down"))
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:     "Ошибка: заказ недоступен",
			userID:   1,
			withAuth: true,
			mockBehavior: func(order *orderMocks.MockOrderClient, rest *restaurantMocks.MockRestaurantClient) {
				order.EXPECT().GetOrders(gomock.Any(), int64(1)).
					Return(nil, errors.New("grpc error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			handler, mockOrder, _, mockRest := setupTestHandler(ctrl)
			tt.mockBehavior(mockOrder, mockRest)

			req := httptest.NewRequest(http.MethodGet, "/orders", nil)
			if tt.withAuth {
				req = withUserIDContext(req, tt.userID)
			}

			w := httptest.NewRecorder()
			handler.GetMyOrders(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestOrderHandler_PayForFriend(t *testing.T) {
	type mockBehavior func(order *orderMocks.MockOrderClient)

	reqBody := PayForFriendRequest{PaymentMethodID: "pm-1"}

	tests := []struct {
		name           string
		splitID        string
		reqBody        interface{}
		headers        map[string]string
		withAuth       bool
		mockBehavior   mockBehavior
		expectedStatus int
	}{
		{
			name:     "Успешный вызов оплаты за друга",
			splitID:  "split-1",
			reqBody:  reqBody,
			headers:  map[string]string{"Idempotency-Key": "idem"},
			withAuth: true,
			mockBehavior: func(order *orderMocks.MockOrderClient) {
				order.EXPECT().PayForFriend(gomock.Any(), "split-1", int64(1), "pm-1", "idem").
					Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Ошибка: пустой split id",
			splitID:        "",
			reqBody:        reqBody,
			headers:        map[string]string{"Idempotency-Key": "idem"},
			withAuth:       true,
			mockBehavior:   func(order *orderMocks.MockOrderClient) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			handler, mockOrder, _, _ := setupTestHandler(ctrl)
			tt.mockBehavior(mockOrder)

			body, _ := json.Marshal(tt.reqBody)
			req := httptest.NewRequest(http.MethodPost, "/splits/"+tt.splitID+"/pay", bytes.NewBuffer(body))
			if tt.withAuth {
				req = withUserIDContext(req, 1)
			}
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}
			req.SetPathValue("id", tt.splitID)

			w := httptest.NewRecorder()
			handler.PayForFriend(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
