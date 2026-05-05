package paymentclient

import (
	"context"
	"errors"
	"testing"

	pbPayment "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/payment"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/payment/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestPaymentClient_InitiateCardBinding(t *testing.T) {
	type mockBehavior func(m *mocks.MockPaymentServiceClient)

	tests := []struct {
		name         string
		userID       int64
		idemKey      string
		mockBehavior mockBehavior
		expectedURL  string
		expectedErr  error
	}{
		{
			name:    "Успешная инициализация привязки",
			userID:  1,
			idemKey: "idem-1",
			mockBehavior: func(m *mocks.MockPaymentServiceClient) {
				m.EXPECT().InitiateCardBinding(gomock.Any(), &pbPayment.InitiateCardBindingRequest{
					UserId:         1,
					IdempotencyKey: "idem-1",
				}).Return(&pbPayment.InitiateCardBindingResponse{
					ConfirmationUrl: "https://yookassa.ru/confirm",
				}, nil)
			},
			expectedURL: "https://yookassa.ru/confirm",
			expectedErr: nil,
		},
		{
			name:    "Внутренняя ошибка сервиса",
			userID:  1,
			idemKey: "idem-2",
			mockBehavior: func(m *mocks.MockPaymentServiceClient) {
				m.EXPECT().InitiateCardBinding(gomock.Any(), gomock.Any()).
					Return(nil, status.Error(codes.Internal, "db error"))
			},
			expectedURL: "",
			expectedErr: ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockGRPCClient := mocks.NewMockPaymentServiceClient(ctrl)
			tt.mockBehavior(mockGRPCClient)

			client := NewPaymentClient(mockGRPCClient)
			url, err := client.InitiateCardBinding(context.Background(), tt.userID, tt.idemKey)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedURL, url)
			}
		})
	}
}

func TestPaymentClient_GetUserCards(t *testing.T) {
	type mockBehavior func(m *mocks.MockPaymentServiceClient)

	tests := []struct {
		name         string
		userID       int64
		mockBehavior mockBehavior
		expectedRes  []PaymentMethod
		expectedErr  error
	}{
		{
			name:   "Успешное получение списка карт",
			userID: 1,
			mockBehavior: func(m *mocks.MockPaymentServiceClient) {
				m.EXPECT().GetUserCards(gomock.Any(), &pbPayment.GetUserCardsRequest{
					UserId: 1,
				}).Return(&pbPayment.GetUserCardsResponse{
					UserId: 1,
					Cards: []*pbPayment.PaymentMethod{
						{
							ExternalId: "ext-1",
							CardType:   "Visa",
							Last4:      "1234",
							IssuerName: "Sber",
							IsDefault:  true,
						},
						{
							ExternalId: "ext-2",
							CardType:   "MasterCard",
							Last4:      "5678",
							IssuerName: "Tinkoff",
							IsDefault:  false,
						},
					},
				}, nil)
			},
			expectedRes: []PaymentMethod{
				{ID: "ext-1", CardType: "Visa", Last4: "1234", IssuerName: "Sber", IsDefault: true},
				{ID: "ext-2", CardType: "MasterCard", Last4: "5678", IssuerName: "Tinkoff", IsDefault: false},
			},
			expectedErr: nil,
		},
		{
			name:   "Ошибка получения списка карт",
			userID: 1,
			mockBehavior: func(m *mocks.MockPaymentServiceClient) {
				m.EXPECT().GetUserCards(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("network error"))
			},
			expectedRes: nil,
			expectedErr: ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockGRPCClient := mocks.NewMockPaymentServiceClient(ctrl)
			tt.mockBehavior(mockGRPCClient)

			client := NewPaymentClient(mockGRPCClient)
			cards, err := client.GetUserCards(context.Background(), tt.userID)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
				assert.Nil(t, cards)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedRes, cards)
			}
		})
	}
}

func TestPaymentClient_SetDefaultCard(t *testing.T) {
	type mockBehavior func(m *mocks.MockPaymentServiceClient)

	tests := []struct {
		name         string
		userID       int64
		cardID       string
		idemKey      string
		mockBehavior mockBehavior
		expectedErr  error
	}{
		{
			name:    "Успешная установка",
			userID:  1,
			cardID:  "ext-1",
			idemKey: "idem-1",
			mockBehavior: func(m *mocks.MockPaymentServiceClient) {
				m.EXPECT().SetDefaultCard(gomock.Any(), &pbPayment.ChangeCardRequest{
					UserId:         1,
					CardId:         "ext-1",
					IdempotencyKey: "idem-1",
				}).Return(&emptypb.Empty{}, nil)
			},
			expectedErr: nil,
		},
		{
			name:    "Карта не найдена (NotFound)",
			userID:  1,
			cardID:  "ext-404",
			idemKey: "idem-2",
			mockBehavior: func(m *mocks.MockPaymentServiceClient) {
				m.EXPECT().SetDefaultCard(gomock.Any(), gomock.Any()).
					Return(nil, status.Error(codes.NotFound, "not found"))
			},
			expectedErr: ErrPaymentMethodNotFound,
		},
		{
			name:    "Внутренняя ошибка (Internal)",
			userID:  1,
			cardID:  "ext-err",
			idemKey: "idem-3",
			mockBehavior: func(m *mocks.MockPaymentServiceClient) {
				m.EXPECT().SetDefaultCard(gomock.Any(), gomock.Any()).
					Return(nil, status.Error(codes.Internal, "db crash"))
			},
			expectedErr: ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockGRPCClient := mocks.NewMockPaymentServiceClient(ctrl)
			tt.mockBehavior(mockGRPCClient)

			client := NewPaymentClient(mockGRPCClient)
			err := client.SetDefaultCard(context.Background(), tt.userID, tt.cardID, tt.idemKey)

			assert.ErrorIs(t, err, tt.expectedErr)
		})
	}
}

func TestPaymentClient_DeleteCard(t *testing.T) {
	type mockBehavior func(m *mocks.MockPaymentServiceClient)

	tests := []struct {
		name         string
		userID       int64
		cardID       string
		idemKey      string
		mockBehavior mockBehavior
		expectedErr  error
	}{
		{
			name:    "Успешное удаление",
			userID:  1,
			cardID:  "ext-1",
			idemKey: "idem-1",
			mockBehavior: func(m *mocks.MockPaymentServiceClient) {
				m.EXPECT().DeleteCard(gomock.Any(), &pbPayment.ChangeCardRequest{
					UserId:         1,
					CardId:         "ext-1",
					IdempotencyKey: "idem-1",
				}).Return(&emptypb.Empty{}, nil)
			},
			expectedErr: nil,
		},
		{
			name:    "Карта не найдена",
			userID:  1,
			cardID:  "ext-404",
			idemKey: "idem-2",
			mockBehavior: func(m *mocks.MockPaymentServiceClient) {
				m.EXPECT().DeleteCard(gomock.Any(), gomock.Any()).
					Return(nil, status.Error(codes.NotFound, "not found"))
			},
			expectedErr: ErrPaymentMethodNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockGRPCClient := mocks.NewMockPaymentServiceClient(ctrl)
			tt.mockBehavior(mockGRPCClient)

			client := NewPaymentClient(mockGRPCClient)
			err := client.DeleteCard(context.Background(), tt.userID, tt.cardID, tt.idemKey)

			assert.ErrorIs(t, err, tt.expectedErr)
		})
	}
}

func TestPaymentClient_ProcessPaymentMethodWebhook(t *testing.T) {
	type mockBehavior func(m *mocks.MockPaymentServiceClient)

	cardInfo := CardInfo{
		First6:      "123456",
		Last4:       "7890",
		ExpiryMonth: "12",
		ExpiryYear:  "2025",
		CardType:    "Visa",
		IssuerName:  "Bank",
	}

	tests := []struct {
		name         string
		id           string
		status       string
		pType        string
		saved        bool
		card         CardInfo
		mockBehavior mockBehavior
		expectedErr  error
	}{
		{
			name:   "Успешная обработка вебхука метода оплаты",
			id:     "pm-123",
			status: "active",
			pType:  "bank_card",
			saved:  true,
			card:   cardInfo,
			mockBehavior: func(m *mocks.MockPaymentServiceClient) {
				m.EXPECT().ProcessPaymentMethodWebhook(gomock.Any(), &pbPayment.ProcessPaymentMethodWebhookRequest{
					Id:     "pm-123",
					Status: "active",
					Saved:  true,
					Type:   "bank_card",
					Card: &pbPayment.CardInfo{
						First6:      "123456",
						Last4:       "7890",
						ExpiryMonth: "12",
						ExpiryYear:  "2025",
						CardType:    "Visa",
						IssuerName:  "Bank",
					},
				}).Return(&emptypb.Empty{}, nil)
			},
			expectedErr: nil,
		},
		{
			name:   "Ошибка обработки",
			id:     "pm-123",
			status: "active",
			pType:  "bank_card",
			saved:  true,
			card:   cardInfo,
			mockBehavior: func(m *mocks.MockPaymentServiceClient) {
				m.EXPECT().ProcessPaymentMethodWebhook(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("timeout"))
			},
			expectedErr: ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockGRPCClient := mocks.NewMockPaymentServiceClient(ctrl)
			tt.mockBehavior(mockGRPCClient)

			client := NewPaymentClient(mockGRPCClient)
			err := client.ProcessPaymentMethodWebhook(context.Background(), tt.id, tt.status, tt.pType, tt.saved, tt.card)

			assert.ErrorIs(t, err, tt.expectedErr)
		})
	}
}

func TestPaymentClient_ProcessPaymentWebhook(t *testing.T) {
	type mockBehavior func(m *mocks.MockPaymentServiceClient)

	tests := []struct {
		name         string
		id           string
		status       string
		mockBehavior mockBehavior
		expectedErr  error
	}{
		{
			name:   "Успешная обработка вебхука платежа",
			id:     "pay-123",
			status: "succeeded",
			mockBehavior: func(m *mocks.MockPaymentServiceClient) {
				m.EXPECT().ProcessPaymentWebhook(gomock.Any(), &pbPayment.ProcessPaymentWebhookRequest{
					Id:     "pay-123",
					Status: "succeeded",
				}).Return(&emptypb.Empty{}, nil)
			},
			expectedErr: nil,
		},
		{
			name:   "Ошибка",
			id:     "pay-err",
			status: "succeeded",
			mockBehavior: func(m *mocks.MockPaymentServiceClient) {
				m.EXPECT().ProcessPaymentWebhook(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("fail"))
			},
			expectedErr: ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockGRPCClient := mocks.NewMockPaymentServiceClient(ctrl)
			tt.mockBehavior(mockGRPCClient)

			client := NewPaymentClient(mockGRPCClient)
			err := client.ProcessPaymentWebhook(context.Background(), tt.id, tt.status)

			assert.ErrorIs(t, err, tt.expectedErr)
		})
	}
}

func TestPaymentClient_RefreshPaymentStatus(t *testing.T) {
	type mockBehavior func(m *mocks.MockPaymentServiceClient)

	tests := []struct {
		name         string
		paymentID    string
		mockBehavior mockBehavior
		expectedStat string
		expectedErr  error
	}{
		{
			name:      "Успешное обновление статуса",
			paymentID: "pay-123",
			mockBehavior: func(m *mocks.MockPaymentServiceClient) {
				m.EXPECT().RefreshPaymentStatus(gomock.Any(), &pbPayment.RefreshPaymentStatusRequest{
					YookassaPaymentId: "pay-123",
				}).Return(&pbPayment.RefreshPaymentStatusResponse{
					Status: "succeeded",
				}, nil)
			},
			expectedStat: "succeeded",
			expectedErr:  nil,
		},
		{
			name:      "Ошибка сети",
			paymentID: "pay-err",
			mockBehavior: func(m *mocks.MockPaymentServiceClient) {
				m.EXPECT().RefreshPaymentStatus(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("network fail"))
			},
			expectedStat: "",
			expectedErr:  ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockGRPCClient := mocks.NewMockPaymentServiceClient(ctrl)
			tt.mockBehavior(mockGRPCClient)

			client := NewPaymentClient(mockGRPCClient)
			statusStr, err := client.RefreshPaymentStatus(context.Background(), tt.paymentID)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedStat, statusStr)
			}
		})
	}
}
