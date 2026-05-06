package yookassa

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_CreatePayment(t *testing.T) {
	tests := []struct {
		name           string
		req            CreatePaymentRequest
		idemKey        string
		mockBehavior   func(w http.ResponseWriter, r *http.Request)
		expectedStatus string
		expectedErr    error
	}{
		{
			name: "Успешное создание платежа",
			req: CreatePaymentRequest{
				Amount: CreatePaymentRequestAmount{Value: "100.00", Currency: "RUB"},
			},
			idemKey: "idem-123",
			mockBehavior: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, "idem-123", r.Header.Get("Idempotence-Key"))
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"id": "pay-123", "status": "pending", "amount": {"value": "100.00", "currency": "RUB"}}`))
			},
			expectedStatus: "pending",
			expectedErr:    nil,
		},
		{
			name: "Ошибка 400 Bad Request",
			mockBehavior: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
			},
			expectedErr: ErrBadRequest,
		},
		{
			name: "Ошибка 401 Unauthorized",
			mockBehavior: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
			},
			expectedErr: ErrUnauthorized,
		},
		{
			name: "Ошибка 404 Not Found",
			mockBehavior: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			},
			expectedErr: ErrNotFound,
		},
		{
			name: "Неизвестная ошибка 500",
			mockBehavior: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tt.mockBehavior))
			defer server.Close()

			client := NewClient("shop-id", "secret-key")
			client.baseURL = server.URL

			resp, err := client.CreatePayment(context.Background(), tt.req, tt.idemKey)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
				assert.Nil(t, resp)
			} else if tt.expectedStatus == "" {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, resp)
				assert.Equal(t, tt.expectedStatus, resp.Status)
			}
		})
	}
}

func TestClient_GetPayment(t *testing.T) {
	tests := []struct {
		name         string
		paymentID    string
		mockBehavior func(w http.ResponseWriter, r *http.Request)
		expectedErr  error
	}{
		{
			name:      "Успешное получение платежа",
			paymentID: "pay-456",
			mockBehavior: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"id": "pay-456", "status": "succeeded"}`))
			},
			expectedErr: nil,
		},
		{
			name:      "Ошибка 404",
			paymentID: "pay-not-found",
			mockBehavior: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			},
			expectedErr: ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tt.mockBehavior))
			defer server.Close()

			client := NewClient("test", "test")
			client.baseURL = server.URL

			resp, err := client.GetPayment(context.Background(), tt.paymentID)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, resp)
				assert.Equal(t, tt.paymentID, resp.ID)
			}
		})
	}
}

func TestClient_CreatePaymentMethod(t *testing.T) {
	tests := []struct {
		name         string
		req          CreatePaymentMethodRequest
		mockBehavior func(w http.ResponseWriter, r *http.Request)
		expectError  bool
	}{
		{
			name: "Успешное создание метода оплаты",
			req:  CreatePaymentMethodRequest{Type: "bank_card"},
			mockBehavior: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"id": "method-1", "type": "bank_card", "saved": true}`))
			},
			expectError: false,
		},
		{
			name: "Ошибка от API (400)",
			req:  CreatePaymentMethodRequest{},
			mockBehavior: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tt.mockBehavior))
			defer server.Close()

			client := NewClient("shop", "secret")
			client.baseURL = server.URL

			resp, err := client.CreatePaymentMethod(context.Background(), tt.req, "idem-1")

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, resp)
				assert.Equal(t, "method-1", resp.ID)
				assert.True(t, resp.Saved)
			}
		})
	}
}
