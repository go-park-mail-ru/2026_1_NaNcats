package address

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/grpc_client/addressclient"
	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/grpc_client/addressclient/mocks"
	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/middleware"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestAddressHandler_AddAddress(t *testing.T) {
	type mockInit func(m *mocks.MockAddressClient)
	tests := []struct {
		name           string
		userID         any
		idemKey        string
		body           any
		mockInit       mockInit
		expectedStatus int
	}{
		{
			name:    "Успешное добавление адреса",
			userID:  int64(1),
			idemKey: "test-key",
			body: AddressRequest{
				AddressText: "Москва",
				Lat:         55.75,
				Lon:         37.61,
				Label:       "Дом",
			},
			mockInit: func(m *mocks.MockAddressClient) {
				m.EXPECT().
					AddAddress(gomock.Any(), int64(1), gomock.Any(), "test-key").
					Return("uuid-123", nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "Ошибка: пользователь не авторизован",
			userID:         nil,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Ошибка: отсутствует заголовок идемпотентности",
			userID:         int64(1),
			idemKey:        "",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Ошибка: невалидный JSON",
			userID:         int64(1),
			idemKey:        "key",
			body:           "invalid-json",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:    "Ошибка gRPC клиента",
			userID:  int64(1),
			idemKey: "key",
			body:    AddressRequest{AddressText: "Тест"},
			mockInit: func(m *mocks.MockAddressClient) {
				m.EXPECT().
					AddAddress(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return("", errors.New("grpc error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClient := mocks.NewMockAddressClient(ctrl)
			if tt.mockInit != nil {
				tt.mockInit(mockClient)
			}

			handler := NewAddressHandler(mockClient, logger.NewNopLogger())

			var buf bytes.Buffer
			if s, ok := tt.body.(string); ok {
				buf.WriteString(s)
			} else {
				_ = json.NewEncoder(&buf).Encode(tt.body)
			}

			req := httptest.NewRequest(http.MethodPost, "/profile/addresses", &buf)
			if tt.userID != nil {
				req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, tt.userID))
			}
			if tt.idemKey != "" {
				req.Header.Set("Idempotency-Key", tt.idemKey)
			}

			rec := httptest.NewRecorder()
			handler.AddAddress(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

func TestAddressHandler_GetAddresses(t *testing.T) {
	type mockInit func(m *mocks.MockAddressClient)
	tests := []struct {
		name           string
		userID         any
		mockInit       mockInit
		expectedStatus int
	}{
		{
			name:   "Успешное получение списка адресов",
			userID: int64(1),
			mockInit: func(m *mocks.MockAddressClient) {
				m.EXPECT().
					GetMyAddresses(gomock.Any(), int64(1)).
					Return([]addressclient.Address{
						{PublicID: "uuid-1", Label: "Дом", Location: addressclient.Location{AddressText: "Арбат"}},
					}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Ошибка: пользователь не авторизован",
			userID:         nil,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:   "Ошибка gRPC клиента",
			userID: int64(1),
			mockInit: func(m *mocks.MockAddressClient) {
				m.EXPECT().
					GetMyAddresses(gomock.Any(), int64(1)).
					Return(nil, errors.New("fail"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClient := mocks.NewMockAddressClient(ctrl)
			if tt.mockInit != nil {
				tt.mockInit(mockClient)
			}

			handler := NewAddressHandler(mockClient, logger.NewNopLogger())

			req := httptest.NewRequest(http.MethodGet, "/profile/addresses", nil)
			if tt.userID != nil {
				req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, tt.userID))
			}

			rec := httptest.NewRecorder()
			handler.GetAddresses(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

func TestAddressHandler_DeleteAddress(t *testing.T) {
	type mockInit func(m *mocks.MockAddressClient)
	tests := []struct {
		name           string
		userID         any
		idemKey        string
		pathID         string
		mockInit       mockInit
		expectedStatus int
	}{
		{
			name:    "Успешное удаление адреса",
			userID:  int64(1),
			idemKey: "idem-key",
			pathID:  "uuid-123",
			mockInit: func(m *mocks.MockAddressClient) {
				m.EXPECT().
					DeleteAddress(gomock.Any(), int64(1), "uuid-123", "idem-key").
					Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Ошибка: не авторизован",
			userID:         nil,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Ошибка: отсутствует Idempotency-Key",
			userID:         int64(1),
			idemKey:        "",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Ошибка: пустой ID в пути",
			userID:         int64(1),
			idemKey:        "key",
			pathID:         "",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:    "Ошибка: адрес не найден",
			userID:  int64(1),
			idemKey: "key",
			pathID:  "not-found",
			mockInit: func(m *mocks.MockAddressClient) {
				m.EXPECT().
					DeleteAddress(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(addressclient.ErrAddressNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClient := mocks.NewMockAddressClient(ctrl)
			if tt.mockInit != nil {
				tt.mockInit(mockClient)
			}

			handler := NewAddressHandler(mockClient, logger.NewNopLogger())

			req := httptest.NewRequest(http.MethodDelete, "/profile/addresses/"+tt.pathID, nil)
			if tt.userID != nil {
				req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, tt.userID))
			}
			if tt.idemKey != "" {
				req.Header.Set("Idempotency-Key", tt.idemKey)
			}
			req.SetPathValue("id", tt.pathID)

			rec := httptest.NewRecorder()
			handler.DeleteAddress(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

func TestAddressHandler_UpdateAddress(t *testing.T) {
	type mockInit func(m *mocks.MockAddressClient)

	tests := []struct {
		name           string
		userID         any
		idemKey        string
		pathID         string
		body           any
		mockInit       mockInit
		expectedStatus int
	}{
		{
			name:    "Успешное обновление адреса",
			userID:  int64(1),
			idemKey: "idem-upd-123",
			pathID:  "addr-uuid-456",
			body: AddressRequest{
				AddressText: "Новая Москва",
				Lat:         55.1,
				Lon:         37.1,
			},
			mockInit: func(m *mocks.MockAddressClient) {
				m.EXPECT().
					UpdateAddress(gomock.Any(), int64(1), gomock.Any(), "idem-upd-123").
					Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Ошибка: пользователь не авторизован",
			userID:         nil,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Ошибка: отсутствует заголовок идемпотентности",
			userID:         int64(1),
			idemKey:        "",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Ошибка: пустой ID адреса в пути",
			userID:         int64(1),
			idemKey:        "key",
			pathID:         "",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Ошибка: невалидный JSON в теле запроса",
			userID:         int64(1),
			idemKey:        "key",
			pathID:         "uuid",
			body:           "{invalid",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:    "Ошибка: адрес не найден",
			userID:  int64(1),
			idemKey: "key",
			pathID:  "missing-uuid",
			body:    AddressRequest{AddressText: "Тест"},
			mockInit: func(m *mocks.MockAddressClient) {
				m.EXPECT().
					UpdateAddress(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(addressclient.ErrAddressNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:    "Ошибка: системная ошибка gRPC",
			userID:  int64(1),
			idemKey: "key",
			pathID:  "uuid",
			body:    AddressRequest{AddressText: "Тест"},
			mockInit: func(m *mocks.MockAddressClient) {
				m.EXPECT().
					UpdateAddress(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New("grpc crash"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClient := mocks.NewMockAddressClient(ctrl)
			if tt.mockInit != nil {
				tt.mockInit(mockClient)
			}

			handler := NewAddressHandler(mockClient, logger.NewNopLogger())

			var buf bytes.Buffer
			if s, ok := tt.body.(string); ok {
				buf.WriteString(s)
			} else {
				_ = json.NewEncoder(&buf).Encode(tt.body)
			}

			req := httptest.NewRequest(http.MethodPatch, "/api/profile/addresses/"+tt.pathID, &buf)
			if tt.userID != nil {
				req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, tt.userID))
			}
			if tt.idemKey != "" {
				req.Header.Set("Idempotency-Key", tt.idemKey)
			}
			req.SetPathValue("id", tt.pathID)

			rec := httptest.NewRecorder()
			handler.UpdateAddress(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}
