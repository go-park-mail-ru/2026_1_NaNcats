package payment

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/grpc_client/paymentclient"
	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/grpc_client/paymentclient/mocks"
	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/middleware"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func withUserIDContext(req *http.Request, userID int64) *http.Request {
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
	return req.WithContext(ctx)
}

func setupTestHandler(ctrl *gomock.Controller) (*PaymentHandler, *mocks.MockPaymentClient) {
	mockClient := mocks.NewMockPaymentClient(ctrl)
	log := logger.NewNopLogger()
	handler := NewPaymentHandler(mockClient, log)
	return handler, mockClient
}

func TestPaymentHandler_InitiateCardBinding(t *testing.T) {
	type mockBehavior func(m *mocks.MockPaymentClient)

	tests := []struct {
		name           string
		userID         int64
		headers        map[string]string
		setupCtx       bool
		mockBehavior   mockBehavior
		expectedStatus int
	}{
		{
			name:     "Успешная инициализация",
			userID:   1,
			headers:  map[string]string{"Idempotency-Key": "idem-123"},
			setupCtx: true,
			mockBehavior: func(m *mocks.MockPaymentClient) {
				m.EXPECT().InitiateCardBinding(gomock.Any(), int64(1), "idem-123").
					Return("https://yookassa.ru/confirm", nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Ошибка: неавторизован (нет в контексте)",
			headers:        map[string]string{"Idempotency-Key": "idem-123"},
			setupCtx:       false,
			mockBehavior:   func(m *mocks.MockPaymentClient) {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Ошибка: нет Idempotency-Key",
			userID:         1,
			headers:        map[string]string{},
			setupCtx:       true,
			mockBehavior:   func(m *mocks.MockPaymentClient) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:     "Ошибка на стороне gRPC клиента",
			userID:   1,
			headers:  map[string]string{"Idempotency-Key": "idem-err"},
			setupCtx: true,
			mockBehavior: func(m *mocks.MockPaymentClient) {
				m.EXPECT().InitiateCardBinding(gomock.Any(), int64(1), "idem-err").
					Return("", paymentclient.ErrInternal)
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

			req := httptest.NewRequest(http.MethodPost, "/profile/cards/bind", nil)
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}
			if tt.setupCtx {
				req = withUserIDContext(req, tt.userID)
			}

			w := httptest.NewRecorder()
			handler.InitiateCardBinding(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedStatus == http.StatusOK {
				var resp BindingResponse
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.Equal(t, "https://yookassa.ru/confirm", resp.ConfirmationURL)
			}
		})
	}
}

func TestPaymentHandler_GetUserCards(t *testing.T) {
	type mockBehavior func(m *mocks.MockPaymentClient)

	tests := []struct {
		name           string
		userID         int64
		setupCtx       bool
		mockBehavior   mockBehavior
		expectedStatus int
	}{
		{
			name:     "Успешное получение списка",
			userID:   1,
			setupCtx: true,
			mockBehavior: func(m *mocks.MockPaymentClient) {
				m.EXPECT().GetUserCards(gomock.Any(), int64(1)).
					Return([]paymentclient.PaymentMethod{
						{ID: "card-1", CardType: "Mir", Last4: "1234"},
					}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Ошибка: неавторизован",
			setupCtx:       false,
			mockBehavior:   func(m *mocks.MockPaymentClient) {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:     "Ошибка gRPC",
			userID:   1,
			setupCtx: true,
			mockBehavior: func(m *mocks.MockPaymentClient) {
				m.EXPECT().GetUserCards(gomock.Any(), int64(1)).
					Return(nil, errors.New("db error"))
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

			req := httptest.NewRequest(http.MethodGet, "/profile/cards", nil)
			if tt.setupCtx {
				req = withUserIDContext(req, tt.userID)
			}

			w := httptest.NewRecorder()
			handler.GetUserCards(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestPaymentHandler_DeleteCard(t *testing.T) {
	type mockBehavior func(m *mocks.MockPaymentClient)

	tests := []struct {
		name           string
		userID         int64
		cardID         string
		headers        map[string]string
		mockBehavior   mockBehavior
		expectedStatus int
	}{
		{
			name:    "Успешное удаление",
			userID:  1,
			cardID:  "card-123",
			headers: map[string]string{"Idempotency-Key": "idem-del"},
			mockBehavior: func(m *mocks.MockPaymentClient) {
				m.EXPECT().DeleteCard(gomock.Any(), int64(1), "card-123", "idem-del").Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:    "Ошибка: Карта не найдена",
			userID:  1,
			cardID:  "card-404",
			headers: map[string]string{"Idempotency-Key": "idem-del"},
			mockBehavior: func(m *mocks.MockPaymentClient) {
				m.EXPECT().DeleteCard(gomock.Any(), int64(1), "card-404", "idem-del").
					Return(paymentclient.ErrPaymentMethodNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Ошибка: пустой Card ID",
			userID:         1,
			cardID:         "", // Симулируем отсутствие параметра
			headers:        map[string]string{"Idempotency-Key": "idem-del"},
			mockBehavior:   func(m *mocks.MockPaymentClient) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			handler, mockClient := setupTestHandler(ctrl)
			tt.mockBehavior(mockClient)

			req := httptest.NewRequest(http.MethodDelete, "/profile/cards/{id}", nil)
			req = withUserIDContext(req, tt.userID)
			req.SetPathValue("id", tt.cardID)

			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			w := httptest.NewRecorder()
			handler.DeleteCard(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestPaymentHandler_SetDefaultCard(t *testing.T) {
	type mockBehavior func(m *mocks.MockPaymentClient)

	tests := []struct {
		name           string
		userID         int64
		cardID         string
		headers        map[string]string
		mockBehavior   mockBehavior
		expectedStatus int
	}{
		{
			name:    "Успешное обновление дефолтной карты",
			userID:  1,
			cardID:  "card-123",
			headers: map[string]string{"Idempotency-Key": "idem-set"},
			mockBehavior: func(m *mocks.MockPaymentClient) {
				m.EXPECT().SetDefaultCard(gomock.Any(), int64(1), "card-123", "idem-set").Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:    "Ошибка: Карта не найдена",
			userID:  1,
			cardID:  "card-404",
			headers: map[string]string{"Idempotency-Key": "idem-set"},
			mockBehavior: func(m *mocks.MockPaymentClient) {
				m.EXPECT().SetDefaultCard(gomock.Any(), int64(1), "card-404", "idem-set").
					Return(paymentclient.ErrPaymentMethodNotFound)
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

			req := httptest.NewRequest(http.MethodPost, "/profile/cards/{id}/default", nil)
			req = withUserIDContext(req, tt.userID)
			req.SetPathValue("id", tt.cardID)

			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			w := httptest.NewRecorder()
			handler.SetDefaultCard(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestPaymentHandler_YookassaWebhook(t *testing.T) {
	type mockBehavior func(m *mocks.MockPaymentClient)

	tests := []struct {
		name           string
		payload        string // JSON payload
		mockBehavior   mockBehavior
		expectedStatus int
	}{
		{
			name: "Успешная обработка payment_method.active",
			payload: `{
				"type": "notification",
				"event": "payment_method.active",
				"object": {
					"id": "pm-123",
					"status": "active",
					"saved": true,
					"type": "bank_card",
					"card": {
						"first6": "123456",
						"last4": "7890",
						"expiry_year": "2025",
						"expiry_month": "12",
						"card_type": "Visa",
						"issuer_name": "Sber"
					}
				}
			}`,
			mockBehavior: func(m *mocks.MockPaymentClient) {
				m.EXPECT().ProcessPaymentMethodWebhook(gomock.Any(), "pm-123", "active", "bank_card", true, paymentclient.CardInfo{
					First6:      "123456",
					Last4:       "7890",
					ExpiryMonth: "12",
					ExpiryYear:  "2025",
					CardType:    "Visa",
					IssuerName:  "Sber",
				}).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Успешная обработка payment.succeeded",
			payload: `{
				"type": "notification",
				"event": "payment.succeeded",
				"object": {
					"id": "pay-123",
					"status": "succeeded"
				}
			}`,
			mockBehavior: func(m *mocks.MockPaymentClient) {
				m.EXPECT().ProcessPaymentWebhook(gomock.Any(), "pay-123", "succeeded").Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Ошибка: невалидный JSON",
			payload:        `{ invalid json `,
			mockBehavior:   func(m *mocks.MockPaymentClient) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Игнорирование неизвестного эвента",
			payload: `{
				"type": "notification",
				"event": "refund.succeeded",
				"object": {}
			}`,
			mockBehavior:   func(m *mocks.MockPaymentClient) {},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			handler, mockClient := setupTestHandler(ctrl)
			tt.mockBehavior(mockClient)

			req := httptest.NewRequest(http.MethodPost, "/webhooks/yookassa", bytes.NewBufferString(tt.payload))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			handler.YookassaWebhook(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
