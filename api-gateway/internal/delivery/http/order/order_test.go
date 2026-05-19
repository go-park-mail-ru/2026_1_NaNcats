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
	userMocks "github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/grpc_client/userclient/mocks"
	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/middleware"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	pbUser "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/user"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// Вспомогательная функция для инжекта UserID в контекст
func withUserIDContext(req *http.Request, userID int64) *http.Request {
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
	return req.WithContext(ctx)
}

func setupTestHandler(ctrl *gomock.Controller) (*OrderHandler, *orderMocks.MockOrderClient, *paymentMocks.MockPaymentClient, *restaurantMocks.MockRestaurantClient, *userMocks.MockUserClient) {
	mockOrder := orderMocks.NewMockOrderClient(ctrl)
	mockPayment := paymentMocks.NewMockPaymentClient(ctrl)
	mockRestaurant := restaurantMocks.NewMockRestaurantClient(ctrl)
	mockUser := userMocks.NewMockUserClient(ctrl)
	log := logger.NewNopLogger()

	handler := NewOrderHandler(mockOrder, mockPayment, mockRestaurant, mockUser, nil, log)
	return handler, mockOrder, mockPayment, mockRestaurant, mockUser
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

			handler, mockOrder, _, _, _ := setupTestHandler(ctrl)
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
	type mockBehavior func(order *orderMocks.MockOrderClient, user *userMocks.MockUserClient)

	validReq := CreateOrderRequest{
		AddressID:          "addr-1",
		RestaurantBranchID: 10,
		RestaurantBrandID:  20,
		PayForAll:          true,
		PayerMapping:       map[string]string{"target-uuid": "payer-uuid"},
	}

	reqEmptyCart := CreateOrderRequest{
		AddressID:          "addr-1",
		RestaurantBranchID: 10,
		RestaurantBrandID:  20,
		PayForAll:          true,
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
			mockBehavior: func(order *orderMocks.MockOrderClient, user *userMocks.MockUserClient) {
				user.EXPECT().ResolvePublicID(gomock.Any(), "target-uuid").Return(int64(2), nil)
				user.EXPECT().ResolvePublicID(gomock.Any(), "payer-uuid").Return(int64(1), nil)

				order.EXPECT().CreateOrder(gomock.Any(), int64(1), gomock.Any(), "idem-123").
					Return("order-uuid-123", nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:     "Ошибка: невалидный target user id",
			reqBody:  validReq,
			headers:  map[string]string{"Idempotency-Key": "idem-123"},
			withAuth: true,
			mockBehavior: func(order *orderMocks.MockOrderClient, user *userMocks.MockUserClient) {
				user.EXPECT().ResolvePublicID(gomock.Any(), "target-uuid").Return(int64(0), errors.New("not found"))
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:     "Ошибка: невалидный payer user id",
			reqBody:  validReq,
			headers:  map[string]string{"Idempotency-Key": "idem-123"},
			withAuth: true,
			mockBehavior: func(order *orderMocks.MockOrderClient, user *userMocks.MockUserClient) {
				user.EXPECT().ResolvePublicID(gomock.Any(), "target-uuid").Return(int64(2), nil)
				user.EXPECT().ResolvePublicID(gomock.Any(), "payer-uuid").Return(int64(0), errors.New("not found"))
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:     "Ошибка: не указан адрес или бранч",
			reqBody:  CreateOrderRequest{RestaurantBrandID: 20},
			headers:  map[string]string{"Idempotency-Key": "idem-123"},
			withAuth: true,
			mockBehavior: func(order *orderMocks.MockOrderClient, user *userMocks.MockUserClient) {
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:     "Ошибка: пустая корзина",
			reqBody:  reqEmptyCart,
			headers:  map[string]string{"Idempotency-Key": "idem-123"},
			withAuth: true,
			mockBehavior: func(order *orderMocks.MockOrderClient, user *userMocks.MockUserClient) {
				order.EXPECT().CreateOrder(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return("", orderclient.ErrCartIsEmpty)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:     "Ошибка: нераспределенные блюда",
			reqBody:  reqEmptyCart,
			headers:  map[string]string{"Idempotency-Key": "idem-123"},
			withAuth: true,
			mockBehavior: func(order *orderMocks.MockOrderClient, user *userMocks.MockUserClient) {
				order.EXPECT().CreateOrder(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return("", orderclient.ErrUnassignedItems)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:     "Ошибка: адрес не найден",
			reqBody:  reqEmptyCart,
			headers:  map[string]string{"Idempotency-Key": "idem-123"},
			withAuth: true,
			mockBehavior: func(order *orderMocks.MockOrderClient, user *userMocks.MockUserClient) {
				order.EXPECT().CreateOrder(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return("", orderclient.ErrAddressNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:     "Ошибка: неизвестная ошибка grpc",
			reqBody:  reqEmptyCart,
			headers:  map[string]string{"Idempotency-Key": "idem-123"},
			withAuth: true,
			mockBehavior: func(order *orderMocks.MockOrderClient, user *userMocks.MockUserClient) {
				order.EXPECT().CreateOrder(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return("", errors.New("internal error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "Ошибка: нет Idempotency-Key",
			reqBody:        validReq,
			headers:        map[string]string{},
			withAuth:       true,
			mockBehavior:   func(order *orderMocks.MockOrderClient, user *userMocks.MockUserClient) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Ошибка: неверный формат json",
			reqBody:        "invalid-json",
			headers:        map[string]string{"Idempotency-Key": "idem-123"},
			withAuth:       true,
			mockBehavior:   func(order *orderMocks.MockOrderClient, user *userMocks.MockUserClient) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Ошибка: неавторизован",
			reqBody:        validReq,
			headers:        map[string]string{"Idempotency-Key": "idem-123"},
			withAuth:       false,
			mockBehavior:   func(order *orderMocks.MockOrderClient, user *userMocks.MockUserClient) {},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			handler, mockOrder, _, _, mockUser := setupTestHandler(ctrl)
			tt.mockBehavior(mockOrder, mockUser)

			body, _ := json.Marshal(tt.reqBody)
			if strBody, ok := tt.reqBody.(string); ok {
				body = []byte(strBody)
			}

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
	type mockBehavior func(order *orderMocks.MockOrderClient, rest *restaurantMocks.MockRestaurantClient, user *userMocks.MockUserClient)

	now := time.Now()
	var ownerID int64 = 1

	tests := []struct {
		name           string
		userID         int64
		query          string
		withAuth       bool
		mockBehavior   mockBehavior
		expectedStatus int
	}{
		{
			name:     "Успешное получение истории заказов (с блюдами, логотипами и юзерами, кастомная пагинация)",
			userID:   1,
			query:    "?limit=5&offset=2",
			withAuth: true,
			mockBehavior: func(order *orderMocks.MockOrderClient, rest *restaurantMocks.MockRestaurantClient, user *userMocks.MockUserClient) {
				order.EXPECT().GetOrders(gomock.Any(), int64(1), int32(5), int32(2)).Return([]orderclient.Order{
					{
						PublicID:          "ord-1",
						Status:            "paid",
						RestaurantBrandID: 20,
						CreatedAt:         now,
						Items: []orderclient.OrderDish{
							{DishID: 100, Quantity: 2, OwnerUserID: &ownerID},
						},
						Splits: []orderclient.OrderSplit{
							{SplitID: "sp-1", UserID: 1},
						},
					},
				}, nil)

				rest.EXPECT().GetDishesByIDs(gomock.Any(), gomock.Any()).Return([]restaurantclient.Dish{
					{ID: 100, Name: "Burger", ImageURL: "img.png"},
				}, nil)

				rest.EXPECT().GetRestaurantLogos(gomock.Any(), gomock.Any()).Return(map[int64]string{
					20: "logo.png",
				}, nil)

				user.EXPECT().GetUsersByIDs(gomock.Any(), gomock.Any()).Return(map[int64]*pbUser.User{
					1: {PublicId: "user-uuid", Name: "Ivan", AvatarUrl: "avatar.png"},
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:     "Успех, даже если restaurant и user сервисы упали (fallback без имен и логотипов, дефолтная пагинация)",
			userID:   1,
			query:    "",
			withAuth: true,
			mockBehavior: func(order *orderMocks.MockOrderClient, rest *restaurantMocks.MockRestaurantClient, user *userMocks.MockUserClient) {
				order.EXPECT().GetOrders(gomock.Any(), int64(1), int32(10), int32(0)).Return([]orderclient.Order{
					{
						PublicID:          "ord-1",
						RestaurantBrandID: 20,
						Items: []orderclient.OrderDish{
							{DishID: 100, OwnerUserID: &ownerID},
						},
						Splits: []orderclient.OrderSplit{
							{SplitID: "sp-1", UserID: 2},
						},
					},
				}, nil)

				user.EXPECT().GetUsersByIDs(gomock.Any(), gomock.Any()).Return(nil, errors.New("user service down"))

				rest.EXPECT().GetDishesByIDs(gomock.Any(), gomock.Any()).Return(nil, errors.New("restaurant service down"))

				rest.EXPECT().GetRestaurantLogos(gomock.Any(), gomock.Any()).Return(nil, errors.New("restaurant service down"))
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:     "Успех, нет заказов",
			userID:   1,
			query:    "",
			withAuth: true,
			mockBehavior: func(order *orderMocks.MockOrderClient, rest *restaurantMocks.MockRestaurantClient, user *userMocks.MockUserClient) {
				order.EXPECT().GetOrders(gomock.Any(), int64(1), int32(10), int32(0)).Return([]orderclient.Order{}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:     "Успех, невалидный limit и offset, fallback на default",
			userID:   1,
			query:    "?limit=abc&offset=xyz",
			withAuth: true,
			mockBehavior: func(order *orderMocks.MockOrderClient, rest *restaurantMocks.MockRestaurantClient, user *userMocks.MockUserClient) {
				order.EXPECT().GetOrders(gomock.Any(), int64(1), int32(10), int32(0)).Return([]orderclient.Order{}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:     "Ошибка: заказ недоступен",
			userID:   1,
			query:    "",
			withAuth: true,
			mockBehavior: func(order *orderMocks.MockOrderClient, rest *restaurantMocks.MockRestaurantClient, user *userMocks.MockUserClient) {
				order.EXPECT().GetOrders(gomock.Any(), int64(1), int32(10), int32(0)).
					Return(nil, errors.New("grpc error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:     "Ошибка: неавторизован",
			userID:   1,
			query:    "",
			withAuth: false,
			mockBehavior: func(order *orderMocks.MockOrderClient, rest *restaurantMocks.MockRestaurantClient, user *userMocks.MockUserClient) {
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			handler, mockOrder, _, mockRest, mockUser := setupTestHandler(ctrl)
			tt.mockBehavior(mockOrder, mockRest, mockUser)

			req := httptest.NewRequest(http.MethodGet, "/orders"+tt.query, nil)
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
		{
			name:           "Ошибка: отсутствует Idempotency-Key",
			splitID:        "split-1",
			reqBody:        reqBody,
			headers:        map[string]string{},
			withAuth:       true,
			mockBehavior:   func(order *orderMocks.MockOrderClient) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Ошибка: неверный JSON",
			splitID:        "split-1",
			reqBody:        "invalid-json",
			headers:        map[string]string{"Idempotency-Key": "idem"},
			withAuth:       true,
			mockBehavior:   func(order *orderMocks.MockOrderClient) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Ошибка: неавторизован",
			splitID:        "split-1",
			reqBody:        reqBody,
			headers:        map[string]string{"Idempotency-Key": "idem"},
			withAuth:       false,
			mockBehavior:   func(order *orderMocks.MockOrderClient) {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:     "Ошибка: сбой grpc",
			splitID:  "split-1",
			reqBody:  reqBody,
			headers:  map[string]string{"Idempotency-Key": "idem"},
			withAuth: true,
			mockBehavior: func(order *orderMocks.MockOrderClient) {
				order.EXPECT().PayForFriend(gomock.Any(), "split-1", int64(1), "pm-1", "idem").
					Return(errors.New("grpc err"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			handler, mockOrder, _, _, _ := setupTestHandler(ctrl)
			tt.mockBehavior(mockOrder)

			body, _ := json.Marshal(tt.reqBody)
			if strBody, ok := tt.reqBody.(string); ok {
				body = []byte(strBody)
			}

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

func TestOrderHandler_TrackOrderWS(t *testing.T) {
	tests := []struct {
		name           string
		orderID        string
		expectedStatus int
	}{
		{
			name:           "Ошибка: пустой order id",
			orderID:        "",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			handler, _, _, _, _ := setupTestHandler(ctrl)

			req := httptest.NewRequest(http.MethodGet, "/ws/orders/"+tt.orderID, nil)
			req.SetPathValue("id", tt.orderID)

			w := httptest.NewRecorder()
			handler.TrackOrderWS(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
