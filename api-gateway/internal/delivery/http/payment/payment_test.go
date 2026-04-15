package payment

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/delivery/middleware"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	domainMocks "github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain/mocks"
	ucMocks "github.com/go-park-mail-ru/2026_1_NaNcats/internal/usecase/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestPaymentHandler_InitiateCardBinding(t *testing.T) {
	type mockInit func(uc *ucMocks.MockPaymentUseCase)

	tests := []struct {
		name           string
		userID         any
		mockInit       mockInit
		expectedStatus int
	}{
		{
			name:   "Успешная инициализация привязки",
			userID: 1,
			mockInit: func(uc *ucMocks.MockPaymentUseCase) {
				uc.EXPECT().
					InitiateCardBinding(gomock.Any(), 1).
					Return("http://confirmation.url", nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Ошибка авторизации",
			userID:         nil,
			mockInit:       func(uc *ucMocks.MockPaymentUseCase) {},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:   "Ошибка UseCase при инициализации",
			userID: 1,
			mockInit: func(uc *ucMocks.MockPaymentUseCase) {
				uc.EXPECT().
					InitiateCardBinding(gomock.Any(), 1).
					Return("", errors.New("binding error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			uc := ucMocks.NewMockPaymentUseCase(ctrl)
			l := domainMocks.NewNopLogger()
			h := NewPaymentHandler(uc, l)

			req := httptest.NewRequest(http.MethodPost, "/profile/cards/bind", nil)
			if tt.userID != nil {
				ctx := context.WithValue(req.Context(), middleware.UserIDKey, tt.userID)
				req = req.WithContext(ctx)
			}

			w := httptest.NewRecorder()
			tt.mockInit(uc)

			h.InitiateCardBinding(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestPaymentHandler_GetUserCards(t *testing.T) {
	type mockInit func(uc *ucMocks.MockPaymentUseCase)

	tests := []struct {
		name           string
		userID         any
		mockInit       mockInit
		expectedStatus int
	}{
		{
			name:   "Успешное получение списка карт",
			userID: 1,
			mockInit: func(uc *ucMocks.MockPaymentUseCase) {
				cards := []domain.PaymentMethod{{ID: 1, Last4: "4444"}}
				uc.EXPECT().GetUserCards(gomock.Any(), 1).Return(cards, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Пользователь не авторизован",
			userID:         nil,
			mockInit:       func(uc *ucMocks.MockPaymentUseCase) {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:   "Ошибка UseCase при получении карт",
			userID: 1,
			mockInit: func(uc *ucMocks.MockPaymentUseCase) {
				uc.EXPECT().GetUserCards(gomock.Any(), 1).Return(nil, errors.New("db error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			uc := ucMocks.NewMockPaymentUseCase(ctrl)
			l := domainMocks.NewNopLogger()
			h := NewPaymentHandler(uc, l)

			req := httptest.NewRequest(http.MethodGet, "/profile/cards", nil)
			if tt.userID != nil {
				ctx := context.WithValue(req.Context(), middleware.UserIDKey, tt.userID)
				req = req.WithContext(ctx)
			}

			w := httptest.NewRecorder()
			tt.mockInit(uc)

			h.GetUserCards(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestPaymentHandler_DeleteCard(t *testing.T) {
	type mockInit func(uc *ucMocks.MockPaymentUseCase)

	tests := []struct {
		name           string
		userID         any
		cardID         string
		mockInit       mockInit
		expectedStatus int
	}{
		{
			name:   "Успешное удаление карты",
			userID: 1,
			cardID: "card-123",
			mockInit: func(uc *ucMocks.MockPaymentUseCase) {
				uc.EXPECT().DeleteCard(gomock.Any(), "card-123", 1).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "Карта не найдена",
			userID: 1,
			cardID: "none",
			mockInit: func(uc *ucMocks.MockPaymentUseCase) {
				uc.EXPECT().DeleteCard(gomock.Any(), "none", 1).Return(domain.ErrPaymentMethodNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Пользователь не авторизован",
			userID:         nil,
			cardID:         "1",
			mockInit:       func(uc *ucMocks.MockPaymentUseCase) {},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			uc := ucMocks.NewMockPaymentUseCase(ctrl)
			l := domainMocks.NewNopLogger()
			h := NewPaymentHandler(uc, l)

			req := httptest.NewRequest(http.MethodDelete, "/profile/cards/"+tt.cardID, nil)
			req.SetPathValue("id", tt.cardID)
			if tt.userID != nil {
				ctx := context.WithValue(req.Context(), middleware.UserIDKey, tt.userID)
				req = req.WithContext(ctx)
			}

			w := httptest.NewRecorder()
			tt.mockInit(uc)

			h.DeleteCard(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestPaymentHandler_SetDefaultCard(t *testing.T) {
	type mockInit func(uc *ucMocks.MockPaymentUseCase)

	tests := []struct {
		name           string
		userID         any
		cardID         string
		mockInit       mockInit
		expectedStatus int
	}{
		{
			name:   "Успешная установка карты по умолчанию",
			userID: 1,
			cardID: "card-123",
			mockInit: func(uc *ucMocks.MockPaymentUseCase) {
				uc.EXPECT().SetDefaultCard(gomock.Any(), "card-123", 1).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "Карта не найдена при установке по умолчанию",
			userID: 1,
			cardID: "unknown",
			mockInit: func(uc *ucMocks.MockPaymentUseCase) {
				uc.EXPECT().SetDefaultCard(gomock.Any(), "unknown", 1).Return(domain.ErrPaymentMethodNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			uc := ucMocks.NewMockPaymentUseCase(ctrl)
			l := domainMocks.NewNopLogger()
			h := NewPaymentHandler(uc, l)

			req := httptest.NewRequest(http.MethodPost, "/profile/cards/"+tt.cardID+"/default", nil)
			req.SetPathValue("id", tt.cardID)
			if tt.userID != nil {
				ctx := context.WithValue(req.Context(), middleware.UserIDKey, tt.userID)
				req = req.WithContext(ctx)
			}

			w := httptest.NewRecorder()
			tt.mockInit(uc)

			h.SetDefaultCard(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestPaymentHandler_YookassaWebhook(t *testing.T) {
	type mockInit func(uc *ucMocks.MockPaymentUseCase)

	tests := []struct {
		name           string
		payload        string
		mockInit       mockInit
		expectedStatus int
	}{
		{
			name:    "Успешная обработка активации платежного метода",
			payload: `{"event": "payment_method.active", "object": {"id": "pm_123"}}`,
			mockInit: func(uc *ucMocks.MockPaymentUseCase) {
				uc.EXPECT().ProcessPaymentMethodWebhook(gomock.Any(), gomock.Any()).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:    "Успешная обработка успешного платежа",
			payload: `{"event": "payment.succeeded", "object": {"id": "pay_123"}}`,
			mockInit: func(uc *ucMocks.MockPaymentUseCase) {
				uc.EXPECT().ProcessPaymentWebhook(gomock.Any(), gomock.Any()).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Некорректный JSON вебхука",
			payload:        `{invalid}`,
			mockInit:       func(uc *ucMocks.MockPaymentUseCase) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			uc := ucMocks.NewMockPaymentUseCase(ctrl)
			l := domainMocks.NewNopLogger()
			h := NewPaymentHandler(uc, l)

			req := httptest.NewRequest(http.MethodPost, "/webhooks/yookassa", bytes.NewBuffer([]byte(tt.payload)))
			w := httptest.NewRecorder()
			tt.mockInit(uc)

			h.YookassaWebhook(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
