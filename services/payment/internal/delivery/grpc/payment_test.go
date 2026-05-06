package grpc

import (
	"context"
	"errors"
	"testing"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/payment/internal/domain"
	ucMocks "github.com/go-park-mail-ru/2026_1_NaNcats/services/payment/internal/usecase/mocks"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/api_clients/yookassa"
	pb "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/payment"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestPaymentHandler_CreatePayment(t *testing.T) {
	type mockInit func(m *ucMocks.MockPaymentUseCase)

	tests := []struct {
		name         string
		req          *pb.CreatePaymentRequest
		mockInit     mockInit
		expectedCode codes.Code
	}{
		{
			name: "Успешное создание платежа",
			req: &pb.CreatePaymentRequest{
				Amount:          1500,
				PaymentMethodId: "pm-123",
				IdempotencyKey:  "idem-1",
			},
			mockInit: func(m *ucMocks.MockPaymentUseCase) {
				m.EXPECT().CreatePayment(gomock.Any(), int64(1500), "pm-123", "idem-1").
					Return("pay-123", "https://url.com", nil)
			},
			expectedCode: codes.OK,
		},
		{
			name: "Ошибка в UseCase",
			req: &pb.CreatePaymentRequest{
				Amount: 100,
			},
			mockInit: func(m *ucMocks.MockPaymentUseCase) {
				m.EXPECT().CreatePayment(gomock.Any(), int64(100), "", "").
					Return("", "", errors.New("internal error"))
			},
			expectedCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUC := ucMocks.NewMockPaymentUseCase(ctrl)
			tt.mockInit(mockUC)

			handler := NewPaymentHandler(mockUC)
			resp, err := handler.CreatePayment(context.Background(), tt.req)

			if tt.expectedCode == codes.OK {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, "pay-123", resp.PaymentId)
				assert.Equal(t, "https://url.com", resp.ConfirmationUrl)
			} else {
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedCode, st.Code())
			}
		})
	}
}

func TestPaymentHandler_InitiateCardBinding(t *testing.T) {
	type mockInit func(m *ucMocks.MockPaymentUseCase)

	tests := []struct {
		name         string
		req          *pb.InitiateCardBindingRequest
		mockInit     mockInit
		expectedCode codes.Code
	}{
		{
			name: "Успешная инициализация привязки",
			req: &pb.InitiateCardBindingRequest{
				UserId:         1,
				IdempotencyKey: "idem-bind",
			},
			mockInit: func(m *ucMocks.MockPaymentUseCase) {
				m.EXPECT().InitiateCardBinding(gomock.Any(), int64(1), "idem-bind").
					Return("https://confirm.com", nil)
			},
			expectedCode: codes.OK,
		},
		{
			name: "Ошибка в UseCase",
			req: &pb.InitiateCardBindingRequest{
				UserId: 2,
			},
			mockInit: func(m *ucMocks.MockPaymentUseCase) {
				m.EXPECT().InitiateCardBinding(gomock.Any(), int64(2), "").
					Return("", errors.New("yookassa fail"))
			},
			expectedCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUC := ucMocks.NewMockPaymentUseCase(ctrl)
			tt.mockInit(mockUC)

			handler := NewPaymentHandler(mockUC)
			resp, err := handler.InitiateCardBinding(context.Background(), tt.req)

			if tt.expectedCode == codes.OK {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, "https://confirm.com", resp.ConfirmationUrl)
			} else {
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedCode, st.Code())
			}
		})
	}
}

func TestPaymentHandler_GetUserCards(t *testing.T) {
	type mockInit func(m *ucMocks.MockPaymentUseCase)

	tests := []struct {
		name         string
		req          *pb.GetUserCardsRequest
		mockInit     mockInit
		expectedCode codes.Code
	}{
		{
			name: "Успешное получение списка карт",
			req:  &pb.GetUserCardsRequest{UserId: 1},
			mockInit: func(m *ucMocks.MockPaymentUseCase) {
				m.EXPECT().GetUserCards(gomock.Any(), int64(1)).
					Return([]domain.PaymentMethod{
						{ID: 1, ExternalID: "ext-1", CardType: "Visa", Last4: "1234"},
						{ID: 2, ExternalID: "ext-2", CardType: "MasterCard", Last4: "5678"},
					}, nil)
			},
			expectedCode: codes.OK,
		},
		{
			name: "Ошибка получения карт",
			req:  &pb.GetUserCardsRequest{UserId: 2},
			mockInit: func(m *ucMocks.MockPaymentUseCase) {
				m.EXPECT().GetUserCards(gomock.Any(), int64(2)).
					Return(nil, errors.New("db fail"))
			},
			expectedCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUC := ucMocks.NewMockPaymentUseCase(ctrl)
			tt.mockInit(mockUC)

			handler := NewPaymentHandler(mockUC)
			resp, err := handler.GetUserCards(context.Background(), tt.req)

			if tt.expectedCode == codes.OK {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Len(t, resp.Cards, 2)
				assert.Equal(t, "1234", resp.Cards[0].Last4)
			} else {
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedCode, st.Code())
			}
		})
	}
}

func TestPaymentHandler_SetDefaultCard(t *testing.T) {
	type mockInit func(m *ucMocks.MockPaymentUseCase)

	tests := []struct {
		name         string
		req          *pb.ChangeCardRequest
		mockInit     mockInit
		expectedCode codes.Code
	}{
		{
			name: "Успешная установка дефолтной карты",
			req: &pb.ChangeCardRequest{
				UserId:         1,
				CardId:         "ext-1",
				IdempotencyKey: "idem-1",
			},
			mockInit: func(m *ucMocks.MockPaymentUseCase) {
				m.EXPECT().SetDefaultCard(gomock.Any(), "ext-1", int64(1), "idem-1").Return(nil)
			},
			expectedCode: codes.OK,
		},
		{
			name: "Ошибка в UseCase",
			req:  &pb.ChangeCardRequest{UserId: 1, CardId: "ext-err"},
			mockInit: func(m *ucMocks.MockPaymentUseCase) {
				m.EXPECT().SetDefaultCard(gomock.Any(), "ext-err", int64(1), "").Return(errors.New("db err"))
			},
			expectedCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUC := ucMocks.NewMockPaymentUseCase(ctrl)
			tt.mockInit(mockUC)

			handler := NewPaymentHandler(mockUC)
			resp, err := handler.SetDefaultCard(context.Background(), tt.req)

			if tt.expectedCode == codes.OK {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
			} else {
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedCode, st.Code())
			}
		})
	}
}

func TestPaymentHandler_DeleteCard(t *testing.T) {
	type mockInit func(m *ucMocks.MockPaymentUseCase)

	tests := []struct {
		name         string
		req          *pb.ChangeCardRequest
		mockInit     mockInit
		expectedCode codes.Code
	}{
		{
			name: "Успешное удаление карты",
			req: &pb.ChangeCardRequest{
				UserId:         1,
				CardId:         "ext-1",
				IdempotencyKey: "idem-1",
			},
			mockInit: func(m *ucMocks.MockPaymentUseCase) {
				m.EXPECT().DeleteCard(gomock.Any(), "ext-1", int64(1), "idem-1").Return(nil)
			},
			expectedCode: codes.OK,
		},
		{
			name: "Ошибка при удалении",
			req:  &pb.ChangeCardRequest{UserId: 1, CardId: "ext-err"},
			mockInit: func(m *ucMocks.MockPaymentUseCase) {
				m.EXPECT().DeleteCard(gomock.Any(), "ext-err", int64(1), "").Return(errors.New("db err"))
			},
			expectedCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUC := ucMocks.NewMockPaymentUseCase(ctrl)
			tt.mockInit(mockUC)

			handler := NewPaymentHandler(mockUC)
			resp, err := handler.DeleteCard(context.Background(), tt.req)

			if tt.expectedCode == codes.OK {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
			} else {
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedCode, st.Code())
			}
		})
	}
}

func TestPaymentHandler_ProcessPaymentMethodWebhook(t *testing.T) {
	type mockInit func(m *ucMocks.MockPaymentUseCase)

	tests := []struct {
		name         string
		req          *pb.ProcessPaymentMethodWebhookRequest
		mockInit     mockInit
		expectedCode codes.Code
	}{
		{
			name: "Успешная обработка вебхука с картой",
			req: &pb.ProcessPaymentMethodWebhookRequest{
				Id:     "pm-123",
				Status: "active",
				Saved:  true,
				Type:   "bank_card",
				Card: &pb.CardInfo{
					First6:      "123456",
					Last4:       "7890",
					ExpiryMonth: "12",
					ExpiryYear:  "2025",
					CardType:    "Visa",
					IssuerName:  "Bank",
				},
			},
			mockInit: func(m *ucMocks.MockPaymentUseCase) {
				expectedObj := &yookassa.WebhookPaymentMethodObject{
					ID:     "pm-123",
					Status: "active",
					Saved:  true,
					Type:   "bank_card",
					Card: &yookassa.PaymentMethodResponseCard{
						First6:      "123456",
						Last4:       "7890",
						ExpiryMonth: "12",
						ExpiryYear:  "2025",
						CardType:    "Visa",
						IssuerName:  "Bank",
					},
				}
				m.EXPECT().ProcessPaymentMethodWebhook(gomock.Any(), expectedObj).Return(nil)
			},
			expectedCode: codes.OK,
		},
		{
			name: "Успешная обработка вебхука без карты (Card = nil)",
			req: &pb.ProcessPaymentMethodWebhookRequest{
				Id:     "pm-456",
				Status: "pending",
				Saved:  false,
				Type:   "sberbank",
				Card:   nil,
			},
			mockInit: func(m *ucMocks.MockPaymentUseCase) {
				expectedObj := &yookassa.WebhookPaymentMethodObject{
					ID:     "pm-456",
					Status: "pending",
					Saved:  false,
					Type:   "sberbank",
					Card:   nil,
				}
				m.EXPECT().ProcessPaymentMethodWebhook(gomock.Any(), expectedObj).Return(nil)
			},
			expectedCode: codes.OK,
		},
		{
			name: "Ошибка из UseCase",
			req: &pb.ProcessPaymentMethodWebhookRequest{
				Id: "pm-err",
			},
			mockInit: func(m *ucMocks.MockPaymentUseCase) {
				m.EXPECT().ProcessPaymentMethodWebhook(gomock.Any(), gomock.Any()).Return(errors.New("fail"))
			},
			expectedCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUC := ucMocks.NewMockPaymentUseCase(ctrl)
			tt.mockInit(mockUC)

			handler := NewPaymentHandler(mockUC)
			resp, err := handler.ProcessPaymentMethodWebhook(context.Background(), tt.req)

			if tt.expectedCode == codes.OK {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
			} else {
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedCode, st.Code())
			}
		})
	}
}

func TestPaymentHandler_ProcessPaymentWebhook(t *testing.T) {
	type mockInit func(m *ucMocks.MockPaymentUseCase)

	tests := []struct {
		name         string
		req          *pb.ProcessPaymentWebhookRequest
		mockInit     mockInit
		expectedCode codes.Code
	}{
		{
			name: "Успешная обработка вебхука платежа",
			req: &pb.ProcessPaymentWebhookRequest{
				Id:     "pay-123",
				Status: "succeeded",
			},
			mockInit: func(m *ucMocks.MockPaymentUseCase) {
				expectedObj := &yookassa.WebhookPaymentObject{
					ID:     "pay-123",
					Status: "succeeded",
				}
				m.EXPECT().ProcessPaymentWebhook(gomock.Any(), expectedObj).Return(nil)
			},
			expectedCode: codes.OK,
		},
		{
			name: "Ошибка из UseCase",
			req:  &pb.ProcessPaymentWebhookRequest{Id: "pay-err"},
			mockInit: func(m *ucMocks.MockPaymentUseCase) {
				m.EXPECT().ProcessPaymentWebhook(gomock.Any(), gomock.Any()).Return(errors.New("webhook fail"))
			},
			expectedCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUC := ucMocks.NewMockPaymentUseCase(ctrl)
			tt.mockInit(mockUC)

			handler := NewPaymentHandler(mockUC)
			resp, err := handler.ProcessPaymentWebhook(context.Background(), tt.req)

			if tt.expectedCode == codes.OK {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
			} else {
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedCode, st.Code())
			}
		})
	}
}

func TestPaymentHandler_RefreshPaymentStatus(t *testing.T) {
	type mockInit func(m *ucMocks.MockPaymentUseCase)

	tests := []struct {
		name         string
		req          *pb.RefreshPaymentStatusRequest
		mockInit     mockInit
		expectedCode codes.Code
	}{
		{
			name: "Успешное обновление статуса",
			req:  &pb.RefreshPaymentStatusRequest{YookassaPaymentId: "pay-123"},
			mockInit: func(m *ucMocks.MockPaymentUseCase) {
				m.EXPECT().RefreshPaymentStatus(gomock.Any(), "pay-123").Return("succeeded", nil)
			},
			expectedCode: codes.OK,
		},
		{
			name: "Ошибка из UseCase",
			req:  &pb.RefreshPaymentStatusRequest{YookassaPaymentId: "pay-err"},
			mockInit: func(m *ucMocks.MockPaymentUseCase) {
				m.EXPECT().RefreshPaymentStatus(gomock.Any(), "pay-err").Return("", errors.New("api error"))
			},
			expectedCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUC := ucMocks.NewMockPaymentUseCase(ctrl)
			tt.mockInit(mockUC)

			handler := NewPaymentHandler(mockUC)
			resp, err := handler.RefreshPaymentStatus(context.Background(), tt.req)

			if tt.expectedCode == codes.OK {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, "succeeded", resp.Status)
			} else {
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedCode, st.Code())
			}
		})
	}
}
