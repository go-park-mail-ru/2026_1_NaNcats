package usecase

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/payment/internal/domain"
	repoMocks "github.com/go-park-mail-ru/2026_1_NaNcats/services/payment/internal/repository/mocks"
	ucMocks "github.com/go-park-mail-ru/2026_1_NaNcats/services/payment/internal/usecase/mocks"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/api_clients/yookassa"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// rewriteTransport - вспомогательный RoundTripper для перенаправления HTTP-запросов
// клиента YooKassa на наш локальный мок-сервер
type rewriteTransport struct {
	Transport http.RoundTripper
	URL       *url.URL
}

func (t *rewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.URL.Scheme = t.URL.Scheme
	req.URL.Host = t.URL.Host
	return t.Transport.RoundTrip(req)
}

func TestPaymentUseCase_CreatePayment(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "idem-123", r.Header.Get("Idempotence-Key"))

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"id": "pay-123",
			"status": "pending",
			"confirmation": {
				"type": "redirect",
				"confirmation_url": "https://yookassa.ru/confirm/123"
			}
		}`))
	}))
	defer server.Close()

	originalTransport := http.DefaultTransport
	if originalTransport == nil {
		originalTransport = http.DefaultTransport
	}
	u, _ := url.Parse(server.URL)
	http.DefaultTransport = &rewriteTransport{
		Transport: http.DefaultTransport,
		URL:       u,
	}
	defer func() { http.DefaultTransport = originalTransport }()

	yc := yookassa.NewClient("shop-id", "secret")
	uc := NewPaymentUseCase(nil, nil, nil, yc, "https://my-app.com/return", logger.NewNopLogger())

	payID, confirmURL, err := uc.CreatePayment(context.Background(), 1500500, "pm-1", "idem-123")

	assert.NoError(t, err)
	assert.Equal(t, "pay-123", payID)
	assert.Equal(t, "https://yookassa.ru/confirm/123", confirmURL)
}

func TestPaymentUseCase_ProcessPaymentMethodWebhook(t *testing.T) {
	type mockInit func(cacheMock *repoMocks.MockPaymentCacheRepository, repoMock *repoMocks.MockPaymentRepository)

	tests := []struct {
		name      string
		pm        *yookassa.WebhookPaymentMethodObject
		mockInit  mockInit
		expectErr bool
	}{
		{
			name: "Игнорирование вебхука (не сохраненная карта)",
			pm: &yookassa.WebhookPaymentMethodObject{
				ID: "pm-1", Status: "active", Saved: false,
			},
			mockInit:  func(c *repoMocks.MockPaymentCacheRepository, r *repoMocks.MockPaymentRepository) {},
			expectErr: false,
		},
		{
			name: "Ошибка: привязка не найдена в кэше",
			pm: &yookassa.WebhookPaymentMethodObject{
				ID: "pm-2", Status: "active", Saved: true,
				Card: &yookassa.PaymentMethodResponseCard{},
			},
			mockInit: func(c *repoMocks.MockPaymentCacheRepository, r *repoMocks.MockPaymentRepository) {
				c.EXPECT().GetUserIDByPaymentID(gomock.Any(), "pm-2").Return(int64(0), errors.New("not found"))
			},
			expectErr: true,
		},
		{
			name: "Успешное сохранение карты",
			pm: &yookassa.WebhookPaymentMethodObject{
				ID: "pm-3", Status: "active", Saved: true,
				Card: &yookassa.PaymentMethodResponseCard{CardType: "Visa", First6: "123456"},
			},
			mockInit: func(c *repoMocks.MockPaymentCacheRepository, r *repoMocks.MockPaymentRepository) {
				c.EXPECT().GetUserIDByPaymentID(gomock.Any(), "pm-3").Return(int64(42), nil)
				r.EXPECT().Create(gomock.Any(), gomock.Any(), "pm-3").Return(int64(10), nil)
				c.EXPECT().DeletePendingBinding(gomock.Any(), "pm-3").Return(nil)
			},
			expectErr: false,
		},
		{
			name: "Успех (карта уже существует - идемпотентность)",
			pm: &yookassa.WebhookPaymentMethodObject{
				ID: "pm-4", Status: "active", Saved: true,
				Card: &yookassa.PaymentMethodResponseCard{},
			},
			mockInit: func(c *repoMocks.MockPaymentCacheRepository, r *repoMocks.MockPaymentRepository) {
				c.EXPECT().GetUserIDByPaymentID(gomock.Any(), "pm-4").Return(int64(42), nil)
				r.EXPECT().Create(gomock.Any(), gomock.Any(), "pm-4").Return(int64(0), domain.ErrPaymentMethodAlreadyExists)
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			cacheMock := repoMocks.NewMockPaymentCacheRepository(ctrl)
			repoMock := repoMocks.NewMockPaymentRepository(ctrl)
			tt.mockInit(cacheMock, repoMock)

			uc := NewPaymentUseCase(repoMock, cacheMock, nil, nil, "", logger.NewNopLogger())
			err := uc.ProcessPaymentMethodWebhook(context.Background(), tt.pm)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPaymentUseCase_ProcessPaymentWebhook(t *testing.T) {
	type mockInit func(orderMock *ucMocks.MockOrderClient)

	tests := []struct {
		name      string
		payment   *yookassa.WebhookPaymentObject
		mockInit  mockInit
		expectErr bool
	}{
		{
			name: "Игнорирование промежуточного статуса",
			payment: &yookassa.WebhookPaymentObject{
				ID: "pay-1", Status: "pending",
			},
			mockInit:  func(o *ucMocks.MockOrderClient) {},
			expectErr: false,
		},
		{
			name: "Успешное уведомление Order Service",
			payment: &yookassa.WebhookPaymentObject{
				ID: "pay-2", Status: "succeeded",
			},
			mockInit: func(o *ucMocks.MockOrderClient) {
				o.EXPECT().UpdateOrderStatus(gomock.Any(), "pay-2", "succeeded").Return(nil)
			},
			expectErr: false,
		},
		{
			name: "Заказ не найден (логируем, но не возвращаем ошибку)",
			payment: &yookassa.WebhookPaymentObject{
				ID: "pay-3", Status: "canceled",
			},
			mockInit: func(o *ucMocks.MockOrderClient) {
				o.EXPECT().UpdateOrderStatus(gomock.Any(), "pay-3", "canceled").
					Return(status.Error(codes.NotFound, "order not found"))
			},
			expectErr: false,
		},
		{
			name: "Критическая ошибка Order Service",
			payment: &yookassa.WebhookPaymentObject{
				ID: "pay-4", Status: "succeeded",
			},
			mockInit: func(o *ucMocks.MockOrderClient) {
				o.EXPECT().UpdateOrderStatus(gomock.Any(), "pay-4", "succeeded").
					Return(status.Error(codes.Internal, "db crash"))
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			orderMock := ucMocks.NewMockOrderClient(ctrl)
			tt.mockInit(orderMock)

			uc := NewPaymentUseCase(nil, nil, orderMock, nil, "", logger.NewNopLogger())
			err := uc.ProcessPaymentWebhook(context.Background(), tt.payment)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPaymentUseCase_CRUD(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repoMock := repoMocks.NewMockPaymentRepository(ctrl)
	uc := NewPaymentUseCase(repoMock, nil, nil, nil, "", logger.NewNopLogger())
	ctx := context.Background()

	t.Run("GetUserCards - Успех", func(t *testing.T) {
		repoMock.EXPECT().GetByUserID(gomock.Any(), int64(1)).Return([]domain.PaymentMethod{{ID: 1}}, nil)

		cards, err := uc.GetUserCards(ctx, 1)
		assert.NoError(t, err)
		assert.Len(t, cards, 1)
	})

	t.Run("SetDefaultCard - Успех", func(t *testing.T) {
		repoMock.EXPECT().SetDefault(gomock.Any(), "ext-1", int64(1)).Return(nil)

		err := uc.SetDefaultCard(ctx, "ext-1", 1, "idem-1")
		assert.NoError(t, err)
	})

	t.Run("SetDefaultCard - Карта не найдена", func(t *testing.T) {
		repoMock.EXPECT().SetDefault(gomock.Any(), "ext-404", int64(1)).Return(domain.ErrPaymentMethodNotFound)

		err := uc.SetDefaultCard(ctx, "ext-404", 1, "idem-1")
		assert.ErrorIs(t, err, domain.ErrPaymentMethodNotFound)
	})

	t.Run("DeleteCard - Успех", func(t *testing.T) {
		repoMock.EXPECT().Delete(gomock.Any(), "ext-1", int64(1)).Return(nil)

		err := uc.DeleteCard(ctx, "ext-1", 1, "idem-1")
		assert.NoError(t, err)
	})
}
