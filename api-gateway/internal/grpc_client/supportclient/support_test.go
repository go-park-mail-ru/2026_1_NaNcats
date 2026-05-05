package supportclient

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	pbSupport "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/support"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/support/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

//go:generate mockgen -destination=../../../../shared/proto/support/mocks/support_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/support SupportServiceClient

func TestSupportClient_CreateTicket(t *testing.T) {
	type mockInit func(m *mocks.MockSupportServiceClient)

	clientID := int64(1)
	guestID := "guest-uuid"
	input := CreateTicketInput{
		ClientID:     &clientID,
		ContactEmail: "test@mail.ru",
		CategoryID:   2,
		FirstMessage: "Help!",
		ClientMeta:   json.RawMessage(`{"os":"linux"}`),
	}

	tests := []struct {
		name     string
		input    CreateTicketInput
		mockInit mockInit
		wantErr  error
	}{
		{
			name:  "Успешное создание тикета",
			input: input,
			mockInit: func(m *mocks.MockSupportServiceClient) {
				m.EXPECT().CreateTicket(gomock.Any(), &pbSupport.CreateTicketRequest{
					ClientId:       clientID,
					ContactEmail:   "test@mail.ru",
					CategoryId:     2,
					InitialMessage: "Help!",
					ClientMeta:     `{"os":"linux"}`,
					IdempotencyKey: "key-1",
				}).Return(&pbSupport.CreateTicketResponse{PublicId: "ticket-uuid"}, nil)
			},
			wantErr: nil,
		},
		{
			name: "Ошибка: неавторизован",
			input: CreateTicketInput{
				GuestID: &guestID,
			},
			mockInit: func(m *mocks.MockSupportServiceClient) {
				m.EXPECT().CreateTicket(gomock.Any(), gomock.Any()).
					Return(nil, status.Error(codes.Unauthenticated, "unauth"))
			},
			wantErr: ErrUnauthorized,
		},
		{
			name:  "Внутренняя ошибка сервиса",
			input: input,
			mockInit: func(m *mocks.MockSupportServiceClient) {
				m.EXPECT().CreateTicket(gomock.Any(), gomock.Any()).
					Return(nil, status.Error(codes.Internal, "crash"))
			},
			wantErr: ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSvc := mocks.NewMockSupportServiceClient(ctrl)
			tt.mockInit(mockSvc)

			client := NewSupportClient(mockSvc)
			res, err := client.CreateTicket(context.Background(), tt.input, "key-1")

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Empty(t, res)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, res)
			}
		})
	}
}

func TestSupportClient_SendMessage(t *testing.T) {
	type mockInit func(m *mocks.MockSupportServiceClient)

	authorID := int64(100)
	input := SendMessageInput{
		TicketPublicID: "ticket-uuid",
		AuthorID:       &authorID,
		AuthorRole:     "user",
		Message:        "Still waiting",
	}

	tests := []struct {
		name     string
		input    SendMessageInput
		mockInit mockInit
		wantErr  error
	}{
		{
			name:  "Успешная отправка сообщения",
			input: input,
			mockInit: func(m *mocks.MockSupportServiceClient) {
				m.EXPECT().SendMessage(gomock.Any(), &pbSupport.SendMessageRequest{
					TicketPublicId: "ticket-uuid",
					AuthorId:       authorID,
					AuthorRole:     "user",
					Message:        "Still waiting",
					IdempotencyKey: "msg-key",
				}).Return(&pbSupport.EventResponse{Id: 500}, nil)
			},
			wantErr: nil,
		},
		{
			name:  "Ошибка: тикет не найден",
			input: input,
			mockInit: func(m *mocks.MockSupportServiceClient) {
				m.EXPECT().SendMessage(gomock.Any(), gomock.Any()).
					Return(nil, status.Error(codes.NotFound, "not found"))
			},
			wantErr: ErrTicketNotFound,
		},
		{
			name:  "Ошибка: системный сбой",
			input: input,
			mockInit: func(m *mocks.MockSupportServiceClient) {
				m.EXPECT().SendMessage(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("rpc error"))
			},
			wantErr: ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSvc := mocks.NewMockSupportServiceClient(ctrl)
			tt.mockInit(mockSvc)

			client := NewSupportClient(mockSvc)
			id, err := client.SendMessage(context.Background(), tt.input, "msg-key")

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Zero(t, id)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, int64(500), id)
			}
		})
	}
}

func TestSupportClient_GetUserTickets(t *testing.T) {
	type mockInit func(m *mocks.MockSupportServiceClient)

	clientID := int64(1)
	guestID := "guest-uuid"
	now := time.Now().Truncate(time.Second)

	tests := []struct {
		name     string
		clientID *int64
		guestID  *string
		mockInit mockInit
		wantErr  error
	}{
		{
			name:     "Успешное получение тикетов пользователя",
			clientID: &clientID,
			mockInit: func(m *mocks.MockSupportServiceClient) {
				m.EXPECT().GetUserTickets(gomock.Any(), &pbSupport.GetUserTicketsRequest{
					ClientId: clientID,
				}).Return(&pbSupport.TicketListResponse{
					Tickets: []*pbSupport.TicketResponse{
						{
							Id:            10,
							PublicId:      "uuid-10",
							CategoryId:    2,
							CurrentStatus: "open",
							SupportLine:   1,
							CreatedAt:     timestamppb.New(now),
						},
					},
				}, nil)
			},
			wantErr: nil,
		},
		{
			name:    "Успешное получение тикетов гостя",
			guestID: &guestID,
			mockInit: func(m *mocks.MockSupportServiceClient) {
				m.EXPECT().GetUserTickets(gomock.Any(), &pbSupport.GetUserTicketsRequest{
					GuestId: guestID,
				}).Return(&pbSupport.TicketListResponse{Tickets: []*pbSupport.TicketResponse{}}, nil)
			},
			wantErr: nil,
		},
		{
			name:     "Ошибка на стороне gRPC сервиса",
			clientID: &clientID,
			mockInit: func(m *mocks.MockSupportServiceClient) {
				m.EXPECT().GetUserTickets(gomock.Any(), gomock.Any()).
					Return(nil, status.Error(codes.DataLoss, "data lost"))
			},
			wantErr: ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSvc := mocks.NewMockSupportServiceClient(ctrl)
			tt.mockInit(mockSvc)

			client := NewSupportClient(mockSvc)
			res, err := client.GetUserTickets(context.Background(), tt.clientID, tt.guestID)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, res)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, res)
			}
		})
	}
}

func TestSupportClient_GetTicketEvents(t *testing.T) {
	type mockInit func(m *mocks.MockSupportServiceClient)

	clientID := int64(1)
	guestID := "guest-uuid"
	now := time.Now().Truncate(time.Second)

	tests := []struct {
		name     string
		publicID string
		clientID *int64
		guestID  *string
		mockInit mockInit
		wantErr  error
	}{
		{
			name:     "Успешное получение событий пользователем",
			publicID: "ticket-uuid",
			clientID: &clientID,
			mockInit: func(m *mocks.MockSupportServiceClient) {
				m.EXPECT().GetTicketEvents(gomock.Any(), &pbSupport.GetTicketEventsRequest{
					TicketPublicId: "ticket-uuid",
					ClientId:       clientID,
				}).Return(&pbSupport.EventListResponse{
					Events: []*pbSupport.Event{
						{
							Id:         100,
							TicketId:   1,
							AuthorRole: "user",
							EventType:  "message",
							Payload:    `{"text":"hello"}`,
							CreatedAt:  timestamppb.New(now),
						},
					},
				}, nil)
			},
			wantErr: nil,
		},
		{
			name:     "Успешное получение событий гостем",
			publicID: "ticket-uuid",
			guestID:  &guestID,
			mockInit: func(m *mocks.MockSupportServiceClient) {
				m.EXPECT().GetTicketEvents(gomock.Any(), &pbSupport.GetTicketEventsRequest{
					TicketPublicId: "ticket-uuid",
					GuestId:        guestID,
				}).Return(&pbSupport.EventListResponse{Events: []*pbSupport.Event{}}, nil)
			},
			wantErr: nil,
		},
		{
			name:     "Тикет не найден",
			publicID: "missing",
			mockInit: func(m *mocks.MockSupportServiceClient) {
				m.EXPECT().GetTicketEvents(gomock.Any(), gomock.Any()).
					Return(nil, status.Error(codes.NotFound, "not found"))
			},
			wantErr: ErrTicketNotFound,
		},
		{
			name:     "Доступ запрещен",
			publicID: "other-user-ticket",
			mockInit: func(m *mocks.MockSupportServiceClient) {
				m.EXPECT().GetTicketEvents(gomock.Any(), gomock.Any()).
					Return(nil, status.Error(codes.PermissionDenied, "denied"))
			},
			wantErr: ErrUnauthorized,
		},
		{
			name:     "Ошибка сервера",
			publicID: "uuid",
			mockInit: func(m *mocks.MockSupportServiceClient) {
				m.EXPECT().GetTicketEvents(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("rpc error"))
			},
			wantErr: ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSvc := mocks.NewMockSupportServiceClient(ctrl)
			tt.mockInit(mockSvc)

			client := NewSupportClient(mockSvc)
			res, err := client.GetTicketEvents(context.Background(), tt.publicID, tt.clientID, tt.guestID)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, res)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, res)
			}
		})
	}
}

func TestSupportClient_RateTicket(t *testing.T) {
	type mockInit func(m *mocks.MockSupportServiceClient)

	clientID := int64(1)

	tests := []struct {
		name     string
		mockInit mockInit
		wantErr  error
	}{
		{
			name: "Успешная оценка",
			mockInit: func(m *mocks.MockSupportServiceClient) {
				m.EXPECT().RateTicket(gomock.Any(), &pbSupport.RateTicketRequest{
					TicketPublicId: "uuid",
					Rating:         5,
					ClientId:       clientID,
					IdempotencyKey: "key",
				}).Return(&pbSupport.SuccessResponse{Success: true}, nil)
			},
			wantErr: nil,
		},
		{
			name: "Ошибка: тикет не найден",
			mockInit: func(m *mocks.MockSupportServiceClient) {
				m.EXPECT().RateTicket(gomock.Any(), gomock.Any()).
					Return(nil, status.Error(codes.NotFound, "not found"))
			},
			wantErr: ErrTicketNotFound,
		},
		{
			name: "Ошибка: системный сбой",
			mockInit: func(m *mocks.MockSupportServiceClient) {
				m.EXPECT().RateTicket(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("fail"))
			},
			wantErr: ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSvc := mocks.NewMockSupportServiceClient(ctrl)
			tt.mockInit(mockSvc)

			client := NewSupportClient(mockSvc)
			err := client.RateTicket(context.Background(), "uuid", 5, &clientID, "key")

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSupportClient_GetAssignedTickets(t *testing.T) {
	type mockInit func(m *mocks.MockSupportServiceClient)

	tests := []struct {
		name     string
		mockInit mockInit
		wantErr  error
	}{
		{
			name: "Успешное получение назначенных тикетов",
			mockInit: func(m *mocks.MockSupportServiceClient) {
				m.EXPECT().GetAssignedTickets(gomock.Any(), &pbSupport.GetAssignedTicketsRequest{
					AgentId: 42,
				}).Return(&pbSupport.TicketListResponse{
					Tickets: []*pbSupport.TicketResponse{
						{Id: 1, PublicId: "p1", CreatedAt: timestamppb.Now()},
					},
				}, nil)
			},
			wantErr: nil,
		},
		{
			name: "Ошибка gRPC",
			mockInit: func(m *mocks.MockSupportServiceClient) {
				m.EXPECT().GetAssignedTickets(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("crash"))
			},
			wantErr: ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSvc := mocks.NewMockSupportServiceClient(ctrl)
			tt.mockInit(mockSvc)

			client := NewSupportClient(mockSvc)
			res, err := client.GetAssignedTickets(context.Background(), 42)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, res)
			} else {
				assert.NoError(t, err)
				assert.Len(t, res, 1)
			}
		})
	}
}

func TestSupportClient_ChangeTicketStatus(t *testing.T) {
	type mockInit func(m *mocks.MockSupportServiceClient)

	tests := []struct {
		name     string
		mockInit mockInit
		wantErr  error
	}{
		{
			name: "Успешное изменение статуса",
			mockInit: func(m *mocks.MockSupportServiceClient) {
				m.EXPECT().ChangeTicketStatus(gomock.Any(), &pbSupport.ChangeTicketStatusRequest{
					TicketPublicId: "ticket-1",
					Status:         "closed",
					AgentId:        42,
					IdempotencyKey: "key-1",
				}).Return(&pbSupport.SuccessResponse{Success: true}, nil)
			},
			wantErr: nil,
		},
		{
			name: "Ошибка: тикет не найден",
			mockInit: func(m *mocks.MockSupportServiceClient) {
				m.EXPECT().ChangeTicketStatus(gomock.Any(), gomock.Any()).
					Return(nil, status.Error(codes.NotFound, "not found"))
			},
			wantErr: ErrTicketNotFound,
		},
		{
			name: "Ошибка: системный сбой",
			mockInit: func(m *mocks.MockSupportServiceClient) {
				m.EXPECT().ChangeTicketStatus(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("rpc error"))
			},
			wantErr: ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSvc := mocks.NewMockSupportServiceClient(ctrl)
			tt.mockInit(mockSvc)

			client := NewSupportClient(mockSvc)
			err := client.ChangeTicketStatus(context.Background(), "ticket-1", "closed", 42, "key-1")

			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestSupportClient_ReassignTicket(t *testing.T) {
	type mockInit func(m *mocks.MockSupportServiceClient)

	tests := []struct {
		name     string
		mockInit mockInit
		wantErr  error
	}{
		{
			name: "Успешное переназначение тикета",
			mockInit: func(m *mocks.MockSupportServiceClient) {
				m.EXPECT().ReassignTicket(gomock.Any(), &pbSupport.ReassignTicketRequest{
					TicketPublicId: "ticket-1",
					AgentId:        100,
					Line:           2,
					AuthorId:       42,
					IdempotencyKey: "key-re",
				}).Return(&pbSupport.SuccessResponse{Success: true}, nil)
			},
			wantErr: nil,
		},
		{
			name: "Ошибка: тикет не найден при переназначении",
			mockInit: func(m *mocks.MockSupportServiceClient) {
				m.EXPECT().ReassignTicket(gomock.Any(), gomock.Any()).
					Return(nil, status.Error(codes.NotFound, "not found"))
			},
			wantErr: ErrTicketNotFound,
		},
		{
			name: "Ошибка: внутренняя ошибка gRPC",
			mockInit: func(m *mocks.MockSupportServiceClient) {
				m.EXPECT().ReassignTicket(gomock.Any(), gomock.Any()).
					Return(nil, status.Error(codes.Internal, "fail"))
			},
			wantErr: ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSvc := mocks.NewMockSupportServiceClient(ctrl)
			tt.mockInit(mockSvc)

			client := NewSupportClient(mockSvc)
			err := client.ReassignTicket(context.Background(), "ticket-1", 100, 2, 42, "key-re")

			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestSupportClient_SetAgentStatus(t *testing.T) {
	type mockInit func(m *mocks.MockSupportServiceClient)

	tests := []struct {
		name     string
		mockInit mockInit
		wantErr  error
	}{
		{
			name: "Успешная смена статуса агента",
			mockInit: func(m *mocks.MockSupportServiceClient) {
				m.EXPECT().SetAgentStatus(gomock.Any(), &pbSupport.SetAgentStatusRequest{
					AgentId: 42,
					Status:  "online",
				}).Return(&pbSupport.SuccessResponse{Success: true}, nil)
			},
			wantErr: nil,
		},
		{
			name: "Ошибка: сбой сервиса",
			mockInit: func(m *mocks.MockSupportServiceClient) {
				m.EXPECT().SetAgentStatus(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("connection failed"))
			},
			wantErr: ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSvc := mocks.NewMockSupportServiceClient(ctrl)
			tt.mockInit(mockSvc)

			client := NewSupportClient(mockSvc)
			err := client.SetAgentStatus(context.Background(), 42, "online")

			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestSupportClient_GetCategories(t *testing.T) {
	type mockInit func(m *mocks.MockSupportServiceClient)

	tests := []struct {
		name     string
		mockInit mockInit
		want     []Category
		wantErr  error
	}{
		{
			name: "Успешное получение категорий",
			mockInit: func(m *mocks.MockSupportServiceClient) {
				m.EXPECT().GetCategories(gomock.Any(), &emptypb.Empty{}).Return(&pbSupport.CategoryListResponse{
					Categories: []*pbSupport.Category{
						{Id: 1, Name: "Технический вопрос", Description: "Баги", DefaultLine: 2},
					},
				}, nil)
			},
			want: []Category{
				{ID: 1, Name: "Технический вопрос", Description: "Баги", DefaultLine: 2},
			},
			wantErr: nil,
		},
		{
			name: "Ошибка gRPC при получении категорий",
			mockInit: func(m *mocks.MockSupportServiceClient) {
				m.EXPECT().GetCategories(gomock.Any(), gomock.Any()).Return(nil, errors.New("rpc fail"))
			},
			want:    nil,
			wantErr: ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSvc := mocks.NewMockSupportServiceClient(ctrl)
			tt.mockInit(mockSvc)

			client := NewSupportClient(mockSvc)
			res, err := client.GetCategories(context.Background())

			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, res)
		})
	}
}

func TestSupportClient_GetTemplates(t *testing.T) {
	type mockInit func(m *mocks.MockSupportServiceClient)

	tests := []struct {
		name     string
		mockInit mockInit
		want     []Template
		wantErr  error
	}{
		{
			name: "Успешное получение шаблонов",
			mockInit: func(m *mocks.MockSupportServiceClient) {
				m.EXPECT().GetTemplates(gomock.Any(), &emptypb.Empty{}).Return(&pbSupport.TemplateListResponse{
					Templates: []*pbSupport.Template{
						{Id: 1, Name: "Приветствие", Content: "Здравствуйте!"},
					},
				}, nil)
			},
			want: []Template{
				{ID: 1, Name: "Приветствие", Content: "Здравствуйте!"},
			},
			wantErr: nil,
		},
		{
			name: "Ошибка gRPC при получении шаблонов",
			mockInit: func(m *mocks.MockSupportServiceClient) {
				m.EXPECT().GetTemplates(gomock.Any(), gomock.Any()).Return(nil, errors.New("rpc fail"))
			},
			want:    nil,
			wantErr: ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSvc := mocks.NewMockSupportServiceClient(ctrl)
			tt.mockInit(mockSvc)

			client := NewSupportClient(mockSvc)
			res, err := client.GetTemplates(context.Background())

			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, res)
		})
	}
}

func TestSupportClient_GetStats(t *testing.T) {
	type mockInit func(m *mocks.MockSupportServiceClient)

	stats := SupportStats{
		TotalTickets:         10,
		ByStatus:             map[string]int64{"open": 5},
		ByCategory:           map[string]int64{"bug": 5},
		AverageRating:        4.5,
		AvgResolutionTimeSec: 3600,
	}

	tests := []struct {
		name     string
		mockInit mockInit
		want     SupportStats
		wantErr  error
	}{
		{
			name: "Успешное получение статистики",
			mockInit: func(m *mocks.MockSupportServiceClient) {
				m.EXPECT().GetStats(gomock.Any(), &emptypb.Empty{}).Return(&pbSupport.SupportStatsResponse{
					TotalTickets:         10,
					ByStatus:             map[string]int64{"open": 5},
					ByCategory:           map[string]int64{"bug": 5},
					AverageRating:        4.5,
					AvgResolutionTimeSec: 3600,
				}, nil)
			},
			want:    stats,
			wantErr: nil,
		},
		{
			name: "Ошибка gRPC при получении статистики",
			mockInit: func(m *mocks.MockSupportServiceClient) {
				m.EXPECT().GetStats(gomock.Any(), gomock.Any()).Return(nil, errors.New("rpc fail"))
			},
			want:    SupportStats{},
			wantErr: ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSvc := mocks.NewMockSupportServiceClient(ctrl)
			tt.mockInit(mockSvc)

			client := NewSupportClient(mockSvc)
			res, err := client.GetStats(context.Background())

			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, res)
		})
	}
}
