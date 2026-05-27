package cart

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/delivery/websocket"
	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/grpc_client/cartclient"
	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/grpc_client/cartclient/mocks"
	userMocks "github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/grpc_client/userclient/mocks"
	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/middleware"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func withUserIDContext(req *http.Request, userID int64) *http.Request {
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
	return req.WithContext(ctx)
}

func setupTestHandler(ctrl *gomock.Controller) (*CartHandler, *mocks.MockCartClient) {
	mockCartClient := mocks.NewMockCartClient(ctrl)
	mockUserClient := userMocks.NewMockUserClient(ctrl)

	mockUserClient.EXPECT().GetUsersByIDs(gomock.Any(), gomock.Any()).Return(nil, errors.New("mock fallback")).AnyTimes()
	mockUserClient.EXPECT().ResolvePublicID(gomock.Any(), gomock.Any()).Return(int64(0), errors.New("mock fallback")).AnyTimes()

	log := logger.NewNopLogger()
	handler := NewCartHandler(mockCartClient, mockUserClient, (*websocket.WsManager)(nil), log)
	return handler, mockCartClient
}

func ptr[T any](v T) *T {
	return &v
}

func TestCartHandler_GetCart(t *testing.T) {
	type mockBehavior func(m *mocks.MockCartClient)

	tests := []struct {
		name           string
		userID         int64
		setupAuth      bool
		mockBehavior   mockBehavior
		expectedStatus int
	}{
		{
			name:      "Успешное получение корзины",
			userID:    1,
			setupAuth: true,
			mockBehavior: func(m *mocks.MockCartClient) {
				m.EXPECT().GetCart(gomock.Any(), int64(1)).Return(&cartclient.Cart{
					ID:                "cart-123",
					AdminID:           1,
					RestaurantBrandID: 10,
					Mode:              "shared",
					Status:            "active",
					TotalCost:         1000,
					Items: []cartclient.Item{
						{DishID: 5, Quantity: 2, Price: 500, OwnerUserID: ptr(int64(1))},
					},
					Members: []cartclient.Member{
						{UserID: 1, JoinedAt: "2023-10-10T10:00:00Z"},
					},
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Ошибка: неавторизован",
			userID:         0,
			setupAuth:      false,
			mockBehavior:   func(m *mocks.MockCartClient) {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:      "Ошибка внутреннего сервиса",
			userID:    1,
			setupAuth: true,
			mockBehavior: func(m *mocks.MockCartClient) {
				m.EXPECT().GetCart(gomock.Any(), gomock.Any()).Return(nil, cartclient.ErrInternal)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			handler, mockClient := setupTestHandler(ctrl)
			tt.mockBehavior(mockClient)

			req := httptest.NewRequest(http.MethodGet, "/api/cart", nil)
			if tt.setupAuth {
				req = withUserIDContext(req, tt.userID)
			}

			w := httptest.NewRecorder()
			handler.GetCart(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedStatus == http.StatusOK {
				var resp CartResponse
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.Equal(t, "cart-123", resp.CartID)
			}
		})
	}
}

func TestCartHandler_AddItem(t *testing.T) {
	type mockBehavior func(m *mocks.MockCartClient)

	validReq := AddItemRequest{
		CartID:   "cart-123",
		DishID:   10,
		Quantity: 1,
	}

	tests := []struct {
		name           string
		reqBody        interface{}
		headers        map[string]string
		setupAuth      bool
		mockBehavior   mockBehavior
		expectedStatus int
	}{
		{
			name:      "Успешное добавление",
			reqBody:   validReq,
			headers:   map[string]string{"Idempotency-Key": "idem-1"},
			setupAuth: true,
			mockBehavior: func(m *mocks.MockCartClient) {
				m.EXPECT().AddItem(gomock.Any(), "cart-123", int64(1), int64(10), int32(1), "idem-1").
					Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Ошибка: нет Idempotency-Key",
			reqBody:        validReq,
			headers:        map[string]string{}, // Пусто
			setupAuth:      true,
			mockBehavior:   func(m *mocks.MockCartClient) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:      "Ошибка: 409 Multiple Restaurants",
			reqBody:   validReq,
			headers:   map[string]string{"Idempotency-Key": "idem-1"},
			setupAuth: true,
			mockBehavior: func(m *mocks.MockCartClient) {
				m.EXPECT().AddItem(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(cartclient.ErrMultipleRestaurants)
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name:      "Ошибка: 409 Cart Locked",
			reqBody:   validReq,
			headers:   map[string]string{"Idempotency-Key": "idem-1"},
			setupAuth: true,
			mockBehavior: func(m *mocks.MockCartClient) {
				m.EXPECT().AddItem(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(cartclient.ErrCartLocked)
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name:      "Ошибка: 403 Forbidden",
			reqBody:   validReq,
			headers:   map[string]string{"Idempotency-Key": "idem-1"},
			setupAuth: true,
			mockBehavior: func(m *mocks.MockCartClient) {
				m.EXPECT().AddItem(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(cartclient.ErrForbidden)
			},
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			handler, mockClient := setupTestHandler(ctrl)
			tt.mockBehavior(mockClient)

			body, _ := json.Marshal(tt.reqBody)
			req := httptest.NewRequest(http.MethodPost, "/api/cart/items", bytes.NewBuffer(body))
			if tt.setupAuth {
				req = withUserIDContext(req, 1)
			}
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			w := httptest.NewRecorder()
			handler.AddItem(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestCartHandler_JoinCart(t *testing.T) {
	type mockBehavior func(m *mocks.MockCartClient)

	validReq := JoinCartRequest{Token: "invite-token"}

	tests := []struct {
		name           string
		reqBody        interface{}
		setupAuth      bool
		mockBehavior   mockBehavior
		expectedStatus int
	}{
		{
			name:      "Успешное присоединение",
			reqBody:   validReq,
			setupAuth: true,
			mockBehavior: func(m *mocks.MockCartClient) {
				m.EXPECT().JoinCart(gomock.Any(), "invite-token", int64(1)).
					Return("cart-123", nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:      "Ошибка: Инвайт недействителен (404)",
			reqBody:   validReq,
			setupAuth: true,
			mockBehavior: func(m *mocks.MockCartClient) {
				m.EXPECT().JoinCart(gomock.Any(), gomock.Any(), gomock.Any()).
					Return("", cartclient.ErrNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			handler, mockClient := setupTestHandler(ctrl)
			tt.mockBehavior(mockClient)

			body, _ := json.Marshal(tt.reqBody)
			req := httptest.NewRequest(http.MethodPost, "/api/cart/join", bytes.NewBuffer(body))
			if tt.setupAuth {
				req = withUserIDContext(req, 1)
			}

			w := httptest.NewRecorder()
			handler.JoinCart(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestCartHandler_ConnectCartWS_Failures(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	handler, _ := setupTestHandler(ctrl)

	t.Run("Ошибка: неавторизован", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/ws/cart?cart_id=123", nil)
		w := httptest.NewRecorder()

		handler.ConnectCartWS(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Ошибка: нет cart_id", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/ws/cart", nil)
		req = withUserIDContext(req, 1)
		w := httptest.NewRecorder()

		handler.ConnectCartWS(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestCartHandler_ClearCart(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	handler, mockClient := setupTestHandler(ctrl)

	t.Run("Успешная очистка корзины", func(t *testing.T) {
		mockClient.EXPECT().ClearCart(gomock.Any(), "cart-123", int64(1), "idem-1").Return(nil)

		body, _ := json.Marshal(BasicCartOperationRequest{CartID: "cart-123"})
		req := httptest.NewRequest(http.MethodDelete, "/api/cart", bytes.NewBuffer(body))
		req.Header.Set("Idempotency-Key", "idem-1")
		req = withUserIDContext(req, 1)

		w := httptest.NewRecorder()
		handler.ClearCart(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}
