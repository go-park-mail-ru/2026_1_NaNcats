package grpc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/support/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/support/internal/usecase/mocks"
	pb "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/support"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestSupportHandler_CreateTicket(t *testing.T) {
	type mockInit func(m *mocks.MockSupportUseCase)

	ticketPublicID := uuid.New()
	guestID := uuid.New().String()

	tests := []struct {
		name          string
		req           *pb.CreateTicketRequest
		mockInit      mockInit
		expectedCode  codes.Code
		checkResponse bool
	}{
		{
			name: "Успешное создание тикета (авторизован)",
			req: &pb.CreateTicketRequest{
				ContactEmail:   "test@mail.ru",
				CategoryId:     1,
				InitialMessage: "Help",
				ClientId:       42,
				IdempotencyKey: "idem-1",
			},
			mockInit: func(m *mocks.MockSupportUseCase) {
				m.EXPECT().
					CreateTicket(gomock.Any(), gomock.Any()).
					Return(ticketPublicID, nil)
			},
			expectedCode:  codes.OK,
			checkResponse: true,
		},
		{
			name: "Успешное создание тикета (гость)",
			req: &pb.CreateTicketRequest{
				ContactEmail:   "guest@mail.ru",
				CategoryId:     2,
				InitialMessage: "Guest help",
				GuestId:        guestID,
				IdempotencyKey: "idem-2",
			},
			mockInit: func(m *mocks.MockSupportUseCase) {
				m.EXPECT().
					CreateTicket(gomock.Any(), gomock.Any()).
					Return(ticketPublicID, nil)
			},
			expectedCode: codes.OK,
		},
		{
			name: "Ошибка: невалидный UUID гостя",
			req: &pb.CreateTicketRequest{
				GuestId: "not-a-uuid",
			},
			mockInit:     func(m *mocks.MockSupportUseCase) {},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Ошибка: UseCase вернул ошибку",
			req: &pb.CreateTicketRequest{
				ContactEmail: "err@mail.ru",
				GuestId:      guestID,
			},
			mockInit: func(m *mocks.MockSupportUseCase) {
				m.EXPECT().
					CreateTicket(gomock.Any(), gomock.Any()).
					Return(uuid.Nil, status.Error(codes.Internal, "internal error"))
			},
			expectedCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUC := mocks.NewMockSupportUseCase(ctrl)
			tt.mockInit(mockUC)

			h := NewSupportHandler(mockUC)
			resp, err := h.CreateTicket(context.Background(), tt.req)

			st, _ := status.FromError(err)
			assert.Equal(t, tt.expectedCode, st.Code())

			if tt.checkResponse {
				assert.NotNil(t, resp)
				assert.Equal(t, ticketPublicID.String(), resp.PublicId)
			}
		})
	}
}

func TestSupportHandler_SendMessage(t *testing.T) {
	type mockInit func(m *mocks.MockSupportUseCase)

	ticketPublicID := uuid.New().String()

	tests := []struct {
		name         string
		req          *pb.SendMessageRequest
		mockInit     mockInit
		expectedCode codes.Code
	}{
		{
			name: "Успешная отправка сообщения",
			req: &pb.SendMessageRequest{
				TicketPublicId: ticketPublicID,
				AuthorId:       10,
				AuthorRole:     "user",
				Message:        "Hello support",
				IdempotencyKey: "msg-idem-1",
			},
			mockInit: func(m *mocks.MockSupportUseCase) {
				m.EXPECT().
					AddMessage(gomock.Any(), gomock.Any(), gomock.Any(), "user", "Hello support", "msg-idem-1").
					Return(nil)
			},
			expectedCode: codes.OK,
		},
		{
			name: "Ошибка: невалидный UUID тикета",
			req: &pb.SendMessageRequest{
				TicketPublicId: "bad-uuid",
			},
			mockInit:     func(m *mocks.MockSupportUseCase) {},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Ошибка: тикет не найден в UseCase",
			req: &pb.SendMessageRequest{
				TicketPublicId: ticketPublicID,
				Message:        "Still here",
			},
			mockInit: func(m *mocks.MockSupportUseCase) {
				m.EXPECT().
					AddMessage(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(status.Error(codes.NotFound, "ticket not found"))
			},
			expectedCode: codes.NotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUC := mocks.NewMockSupportUseCase(ctrl)
			tt.mockInit(mockUC)

			h := NewSupportHandler(mockUC)
			resp, err := h.SendMessage(context.Background(), tt.req)

			st, _ := status.FromError(err)
			assert.Equal(t, tt.expectedCode, st.Code())

			if tt.expectedCode == codes.OK {
				assert.NotNil(t, resp)
				assert.NotNil(t, resp.CreatedAt)
			}
		})
	}
}

func TestSupportHandler_GetUserTickets(t *testing.T) {
	type mockInit func(m *mocks.MockSupportUseCase)

	guestID := uuid.New()
	clientID := int64(10)
	ticketPublicID := uuid.New()

	tests := []struct {
		name         string
		req          *pb.GetUserTicketsRequest
		mockInit     mockInit
		expectedCode codes.Code
		expectedLen  int
	}{
		{
			name: "Успешное получение тикетов клиента",
			req: &pb.GetUserTicketsRequest{
				ClientId: clientID,
			},
			mockInit: func(m *mocks.MockSupportUseCase) {
				m.EXPECT().
					GetMyTickets(gomock.Any(), gomock.Eq(&clientID), gomock.Nil()).
					Return([]domain.Ticket{
						{ID: 1, PublicID: ticketPublicID, CurrentStatus: "open", CreatedAt: time.Now()},
					}, nil)
			},
			expectedCode: codes.OK,
			expectedLen:  1,
		},
		{
			name: "Успешное получение тикетов гостя",
			req: &pb.GetUserTicketsRequest{
				GuestId: guestID.String(),
			},
			mockInit: func(m *mocks.MockSupportUseCase) {
				m.EXPECT().
					GetMyTickets(gomock.Any(), gomock.Nil(), gomock.Eq(&guestID)).
					Return([]domain.Ticket{}, nil)
			},
			expectedCode: codes.OK,
			expectedLen:  0,
		},
		{
			name: "Ошибка: невалидный UUID гостя",
			req: &pb.GetUserTicketsRequest{
				GuestId: "invalid-uuid",
			},
			mockInit:     func(m *mocks.MockSupportUseCase) {},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Ошибка: UseCase вернул ошибку",
			req: &pb.GetUserTicketsRequest{
				ClientId: clientID,
			},
			mockInit: func(m *mocks.MockSupportUseCase) {
				m.EXPECT().
					GetMyTickets(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, status.Error(codes.Internal, "db error"))
			},
			expectedCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUC := mocks.NewMockSupportUseCase(ctrl)
			tt.mockInit(mockUC)

			h := NewSupportHandler(mockUC)
			resp, err := h.GetUserTickets(context.Background(), tt.req)

			st, _ := status.FromError(err)
			assert.Equal(t, tt.expectedCode, st.Code())

			if tt.expectedCode == codes.OK {
				assert.NotNil(t, resp)
				assert.Len(t, resp.Tickets, tt.expectedLen)
			}
		})
	}
}

func TestSupportHandler_GetTicketEvents(t *testing.T) {
	type mockInit func(m *mocks.MockSupportUseCase)

	ticketPublicID := uuid.New()
	guestID := uuid.New()
	clientID := int64(42)

	tests := []struct {
		name         string
		req          *pb.GetTicketEventsRequest
		mockInit     mockInit
		expectedCode codes.Code
		expectedLen  int
	}{
		{
			name: "Успешное получение событий",
			req: &pb.GetTicketEventsRequest{
				TicketPublicId: ticketPublicID.String(),
				ClientId:       clientID,
				GuestId:        guestID.String(),
			},
			mockInit: func(m *mocks.MockSupportUseCase) {
				m.EXPECT().
					GetTicketEvents(gomock.Any(), gomock.Eq(ticketPublicID), gomock.Eq(&clientID), gomock.Eq(&guestID)).
					Return([]domain.Event{
						{ID: 1, TicketID: 100, AuthorRole: "user", EventType: "message", Payload: []byte(`{"text":"hi"}`), CreatedAt: time.Now()},
					}, nil)
			},
			expectedCode: codes.OK,
			expectedLen:  1,
		},
		{
			name: "Ошибка: невалидный UUID тикета",
			req: &pb.GetTicketEventsRequest{
				TicketPublicId: "not-a-uuid",
			},
			mockInit:     func(m *mocks.MockSupportUseCase) {},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Ошибка: доступ запрещен (UseCase)",
			req: &pb.GetTicketEventsRequest{
				TicketPublicId: ticketPublicID.String(),
			},
			mockInit: func(m *mocks.MockSupportUseCase) {
				m.EXPECT().
					GetTicketEvents(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, status.Error(codes.PermissionDenied, "access denied"))
			},
			expectedCode: codes.PermissionDenied,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUC := mocks.NewMockSupportUseCase(ctrl)
			tt.mockInit(mockUC)

			h := NewSupportHandler(mockUC)
			resp, err := h.GetTicketEvents(context.Background(), tt.req)

			st, _ := status.FromError(err)
			assert.Equal(t, tt.expectedCode, st.Code())

			if tt.expectedCode == codes.OK {
				assert.NotNil(t, resp)
				assert.Len(t, resp.Events, tt.expectedLen)
			}
		})
	}
}

func TestSupportHandler_RateTicket(t *testing.T) {
	type mockInit func(m *mocks.MockSupportUseCase)
	ticketUUID := uuid.New()
	clientID := int64(10)

	tests := []struct {
		name         string
		req          *pb.RateTicketRequest
		mockInit     mockInit
		expectedCode codes.Code
	}{
		{
			name: "Успешная оценка тикета",
			req: &pb.RateTicketRequest{
				TicketPublicId: ticketUUID.String(),
				Rating:         5,
				ClientId:       clientID,
				IdempotencyKey: "rate-key",
			},
			mockInit: func(m *mocks.MockSupportUseCase) {
				m.EXPECT().
					RateTicket(gomock.Any(), ticketUUID, 5, &clientID, "rate-key").
					Return(nil)
			},
			expectedCode: codes.OK,
		},
		{
			name: "Ошибка: неверный формат UUID",
			req: &pb.RateTicketRequest{
				TicketPublicId: "not-a-uuid",
			},
			mockInit:     func(m *mocks.MockSupportUseCase) {},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Ошибка: тикет не найден в UseCase",
			req: &pb.RateTicketRequest{
				TicketPublicId: ticketUUID.String(),
				Rating:         1,
			},
			mockInit: func(m *mocks.MockSupportUseCase) {
				m.EXPECT().
					RateTicket(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(status.Error(codes.NotFound, "TICKET_NOT_FOUND"))
			},
			expectedCode: codes.NotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUC := mocks.NewMockSupportUseCase(ctrl)
			tt.mockInit(mockUC)

			h := NewSupportHandler(mockUC)
			resp, err := h.RateTicket(context.Background(), tt.req)

			st, _ := status.FromError(err)
			assert.Equal(t, tt.expectedCode, st.Code())
			if tt.expectedCode == codes.OK {
				assert.True(t, resp.Success)
			}
		})
	}
}

func TestSupportHandler_GetAssignedTickets(t *testing.T) {
	type mockInit func(m *mocks.MockSupportUseCase)
	agentID := int64(777)

	tests := []struct {
		name         string
		req          *pb.GetAssignedTicketsRequest
		mockInit     mockInit
		expectedCode codes.Code
		expectedLen  int
	}{
		{
			name: "Успешное получение назначенных тикетов",
			req: &pb.GetAssignedTicketsRequest{
				AgentId: agentID,
			},
			mockInit: func(m *mocks.MockSupportUseCase) {
				m.EXPECT().
					GetAssignedTickets(gomock.Any(), agentID).
					Return([]domain.Ticket{
						{ID: 1, PublicID: uuid.New(), CurrentStatus: "open", CreatedAt: time.Now()},
						{ID: 2, PublicID: uuid.New(), CurrentStatus: "in_progress", CreatedAt: time.Now()},
					}, nil)
			},
			expectedCode: codes.OK,
			expectedLen:  2,
		},
		{
			name: "Успех: список пуст",
			req:  &pb.GetAssignedTicketsRequest{AgentId: agentID},
			mockInit: func(m *mocks.MockSupportUseCase) {
				m.EXPECT().GetAssignedTickets(gomock.Any(), agentID).Return([]domain.Ticket{}, nil)
			},
			expectedCode: codes.OK,
			expectedLen:  0,
		},
		{
			name: "Ошибка UseCase",
			req:  &pb.GetAssignedTicketsRequest{AgentId: agentID},
			mockInit: func(m *mocks.MockSupportUseCase) {
				m.EXPECT().GetAssignedTickets(gomock.Any(), agentID).Return(nil, status.Error(codes.Internal, "DB_ERROR"))
			},
			expectedCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUC := mocks.NewMockSupportUseCase(ctrl)
			tt.mockInit(mockUC)

			h := NewSupportHandler(mockUC)
			resp, err := h.GetAssignedTickets(context.Background(), tt.req)

			st, _ := status.FromError(err)
			assert.Equal(t, tt.expectedCode, st.Code())
			if tt.expectedCode == codes.OK {
				assert.Len(t, resp.Tickets, tt.expectedLen)
			}
		})
	}
}

func TestSupportHandler_ChangeTicketStatus(t *testing.T) {
	type mockInit func(m *mocks.MockSupportUseCase)
	ticketUUID := uuid.New()
	agentID := int64(42)

	tests := []struct {
		name         string
		req          *pb.ChangeTicketStatusRequest
		mockInit     mockInit
		expectedCode codes.Code
	}{
		{
			name: "Успешная смена статуса",
			req: &pb.ChangeTicketStatusRequest{
				TicketPublicId: ticketUUID.String(),
				Status:         "closed",
				AgentId:        agentID,
				IdempotencyKey: "status-key",
			},
			mockInit: func(m *mocks.MockSupportUseCase) {
				m.EXPECT().
					ChangeTicketStatus(gomock.Any(), ticketUUID, "closed", agentID, "status-key").
					Return(nil)
			},
			expectedCode: codes.OK,
		},
		{
			name: "Ошибка: кривой UUID",
			req: &pb.ChangeTicketStatusRequest{
				TicketPublicId: "bad-uuid",
			},
			mockInit:     func(m *mocks.MockSupportUseCase) {},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Ошибка: отказ доступа в UseCase",
			req: &pb.ChangeTicketStatusRequest{
				TicketPublicId: ticketUUID.String(),
				Status:         "in_progress",
				AgentId:        agentID,
			},
			mockInit: func(m *mocks.MockSupportUseCase) {
				m.EXPECT().
					ChangeTicketStatus(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(status.Error(codes.PermissionDenied, "NOT_YOUR_TICKET"))
			},
			expectedCode: codes.PermissionDenied,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUC := mocks.NewMockSupportUseCase(ctrl)
			tt.mockInit(mockUC)

			h := NewSupportHandler(mockUC)
			resp, err := h.ChangeTicketStatus(context.Background(), tt.req)

			st, _ := status.FromError(err)
			assert.Equal(t, tt.expectedCode, st.Code())
			if tt.expectedCode == codes.OK {
				assert.True(t, resp.Success)
			}
		})
	}
}

func TestSupportHandler_ReassignTicket(t *testing.T) {
	type mockInit func(m *mocks.MockSupportUseCase)

	ticketID := uuid.New()
	agentID := int64(777)
	authorID := int64(1)
	idemKey := "reassign-idem"

	tests := []struct {
		name         string
		req          *pb.ReassignTicketRequest
		mockInit     mockInit
		expectedCode codes.Code
	}{
		{
			name: "Успешное переназначение тикета",
			req: &pb.ReassignTicketRequest{
				TicketPublicId: ticketID.String(),
				AgentId:        agentID,
				Line:           2,
				AuthorId:       authorID,
				IdempotencyKey: idemKey,
			},
			mockInit: func(m *mocks.MockSupportUseCase) {
				m.EXPECT().
					ReassignTicket(gomock.Any(), ticketID, agentID, 2, authorID, idemKey).
					Return(nil)
			},
			expectedCode: codes.OK,
		},
		{
			name: "Ошибка: невалидный UUID тикета",
			req: &pb.ReassignTicketRequest{
				TicketPublicId: "not-a-uuid",
			},
			mockInit:     func(m *mocks.MockSupportUseCase) {},
			expectedCode: codes.InvalidArgument,
		},
		{
			name: "Ошибка UseCase: тикет не найден",
			req: &pb.ReassignTicketRequest{
				TicketPublicId: ticketID.String(),
				AgentId:        agentID,
			},
			mockInit: func(m *mocks.MockSupportUseCase) {
				m.EXPECT().
					ReassignTicket(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(status.Error(codes.NotFound, "ticket not found"))
			},
			expectedCode: codes.NotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUC := mocks.NewMockSupportUseCase(ctrl)
			tt.mockInit(mockUC)

			h := NewSupportHandler(mockUC)
			resp, err := h.ReassignTicket(context.Background(), tt.req)

			st, _ := status.FromError(err)
			assert.Equal(t, tt.expectedCode, st.Code())

			if tt.expectedCode == codes.OK {
				assert.NotNil(t, resp)
				assert.True(t, resp.Success)
			}
		})
	}
}

func TestSupportHandler_SetAgentStatus(t *testing.T) {
	type mockInit func(m *mocks.MockSupportUseCase)

	agentID := int64(42)
	newStatus := "online"

	tests := []struct {
		name         string
		req          *pb.SetAgentStatusRequest
		mockInit     mockInit
		expectedCode codes.Code
	}{
		{
			name: "Успешное обновление статуса агента",
			req: &pb.SetAgentStatusRequest{
				AgentId: agentID,
				Status:  newStatus,
			},
			mockInit: func(m *mocks.MockSupportUseCase) {
				m.EXPECT().
					SetAgentStatus(gomock.Any(), agentID, newStatus).
					Return(nil)
			},
			expectedCode: codes.OK,
		},
		{
			name: "Ошибка UseCase: невалидный статус",
			req: &pb.SetAgentStatusRequest{
				AgentId: agentID,
				Status:  "invalid_status",
			},
			mockInit: func(m *mocks.MockSupportUseCase) {
				m.EXPECT().
					SetAgentStatus(gomock.Any(), agentID, "invalid_status").
					Return(status.Error(codes.InvalidArgument, "invalid status"))
			},
			expectedCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUC := mocks.NewMockSupportUseCase(ctrl)
			tt.mockInit(mockUC)

			h := NewSupportHandler(mockUC)
			resp, err := h.SetAgentStatus(context.Background(), tt.req)

			st, _ := status.FromError(err)
			assert.Equal(t, tt.expectedCode, st.Code())

			if tt.expectedCode == codes.OK {
				assert.NotNil(t, resp)
				assert.True(t, resp.Success)
			}
		})
	}
}

func TestSupportHandler_GetCategories(t *testing.T) {
	type mockInit func(m *mocks.MockSupportUseCase)

	categories := []domain.Category{
		{ID: 1, Name: "Технический вопрос", Description: "Баги", DefaultLine: 2},
		{ID: 2, Name: "Оплата", Description: "Деньги", DefaultLine: 1},
	}

	tests := []struct {
		name         string
		mockInit     mockInit
		expectedCode codes.Code
		expectedLen  int
	}{
		{
			name: "Успешное получение списка категорий",
			mockInit: func(m *mocks.MockSupportUseCase) {
				m.EXPECT().
					GetCategories(gomock.Any()).
					Return(categories, nil)
			},
			expectedCode: codes.OK,
			expectedLen:  2,
		},
		{
			name: "Ошибка UseCase",
			mockInit: func(m *mocks.MockSupportUseCase) {
				m.EXPECT().
					GetCategories(gomock.Any()).
					Return(nil, errors.New("db error"))
			},
			expectedCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUC := mocks.NewMockSupportUseCase(ctrl)
			tt.mockInit(mockUC)

			h := NewSupportHandler(mockUC)
			resp, err := h.GetCategories(context.Background(), &emptypb.Empty{})

			st, _ := status.FromError(err)
			assert.Equal(t, tt.expectedCode, st.Code())

			if tt.expectedCode == codes.OK {
				assert.NotNil(t, resp)
				assert.Len(t, resp.Categories, tt.expectedLen)
				assert.Equal(t, categories[0].Name, resp.Categories[0].Name)
			}
		})
	}
}

func TestSupportHandler_GetTemplates(t *testing.T) {
	type mockInit func(m *mocks.MockSupportUseCase)

	templates := []domain.Template{
		{ID: 1, Name: "Приветствие", Content: "Здравствуйте, чем я могу вам помочь?"},
	}

	tests := []struct {
		name         string
		mockInit     mockInit
		expectedCode codes.Code
		expectedLen  int
	}{
		{
			name: "Успешное получение списка шаблонов",
			mockInit: func(m *mocks.MockSupportUseCase) {
				m.EXPECT().
					GetTemplates(gomock.Any()).
					Return(templates, nil)
			},
			expectedCode: codes.OK,
			expectedLen:  1,
		},
		{
			name: "Ошибка UseCase",
			mockInit: func(m *mocks.MockSupportUseCase) {
				m.EXPECT().
					GetTemplates(gomock.Any()).
					Return(nil, errors.New("internal error"))
			},
			expectedCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUC := mocks.NewMockSupportUseCase(ctrl)
			tt.mockInit(mockUC)

			h := NewSupportHandler(mockUC)
			resp, err := h.GetTemplates(context.Background(), &emptypb.Empty{})

			st, _ := status.FromError(err)
			assert.Equal(t, tt.expectedCode, st.Code())

			if tt.expectedCode == codes.OK {
				assert.NotNil(t, resp)
				assert.Len(t, resp.Templates, tt.expectedLen)
				assert.Equal(t, templates[0].Content, resp.Templates[0].Content)
			}
		})
	}
}
