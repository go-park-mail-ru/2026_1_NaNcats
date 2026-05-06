package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/support/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/support/internal/repository/mocks"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/errutil"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestSupportUseCase_CreateTicket(t *testing.T) {
	type mockInit func(m *mocks.MockSupportRepository)

	validGuestID := uuid.New()
	validPublicID := uuid.New()
	clientID := int64(42)

	tests := []struct {
		name          string
		input         domain.CreateTicketInput
		mockInit      mockInit
		expectedID    uuid.UUID
		expectedError error
	}{
		{
			name: "Успешное создание тикета (авторизованный пользователь)",
			input: domain.CreateTicketInput{
				ClientID:     &clientID,
				CategoryID:   2,
				FirstMessage: "Problem",
			},
			mockInit: func(m *mocks.MockSupportRepository) {
				m.EXPECT().GetActiveCategories(gomock.Any()).Return([]domain.Category{
					{ID: 2, DefaultLine: 2},
				}, nil)
				// Проверяем, что SupportLine выставилась из категории
				m.EXPECT().CreateTicket(gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, in domain.CreateTicketInput) (uuid.UUID, error) {
						if in.SupportLine != 2 {
							return uuid.Nil, errors.New("wrong line")
						}
						return validPublicID, nil
					})
			},
			expectedID: validPublicID,
		},
		{
			name: "Успешное создание тикета (гость)",
			input: domain.CreateTicketInput{
				GuestID:      &validGuestID,
				CategoryID:   1,
				FirstMessage: "Guest help",
			},
			mockInit: func(m *mocks.MockSupportRepository) {
				m.EXPECT().GetActiveCategories(gomock.Any()).Return([]domain.Category{}, nil)
				m.EXPECT().CreateTicket(gomock.Any(), gomock.Any()).Return(validPublicID, nil)
			},
			expectedID: validPublicID,
		},
		{
			name: "Ошибка: не авторизован (нет ClientID и GuestID)",
			input: domain.CreateTicketInput{
				CategoryID: 1,
			},
			mockInit:      func(m *mocks.MockSupportRepository) {},
			expectedError: domain.ErrUnauthorized,
		},
		{
			name: "Ошибка БД при получении категорий",
			input: domain.CreateTicketInput{
				ClientID:   &clientID,
				CategoryID: 1,
			},
			mockInit: func(m *mocks.MockSupportRepository) {
				m.EXPECT().GetActiveCategories(gomock.Any()).Return(nil, errors.New("db error"))
			},
			expectedError: errutil.Internal("failed to fetch categories for routing", errors.New("db error")),
		},
		{
			name: "Ошибка БД при сохранении тикета",
			input: domain.CreateTicketInput{
				ClientID:   &clientID,
				CategoryID: 1,
			},
			mockInit: func(m *mocks.MockSupportRepository) {
				m.EXPECT().GetActiveCategories(gomock.Any()).Return([]domain.Category{}, nil)
				m.EXPECT().CreateTicket(gomock.Any(), gomock.Any()).Return(uuid.Nil, errors.New("save error"))
			},
			expectedError: errutil.Internal("failed to create ticket in db", errors.New("save error")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockSupportRepository(ctrl)
			tt.mockInit(repo)

			uc := NewSupportUseCase(repo)
			res, err := uc.CreateTicket(context.Background(), tt.input)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedID, res)
			}
		})
	}
}

func TestSupportUseCase_GetStats(t *testing.T) {
	type mockInit func(m *mocks.MockSupportRepository)

	agentID := int64(7)
	stats := domain.SupportStats{TotalTickets: 100}

	tests := []struct {
		name          string
		agentID       int64
		mockInit      mockInit
		expectedRes   domain.SupportStats
		expectedError error
	}{
		{
			name:    "Успешное получение статистики (агент 2 линии)",
			agentID: agentID,
			mockInit: func(m *mocks.MockSupportRepository) {
				m.EXPECT().GetAgentProfile(gomock.Any(), agentID).Return(domain.AgentProfile{
					SupportLine: 2,
				}, nil)
				m.EXPECT().GetStats(gomock.Any()).Return(stats, nil)
			},
			expectedRes: stats,
		},
		{
			name:    "Ошибка: профиль агента не найден",
			agentID: agentID,
			mockInit: func(m *mocks.MockSupportRepository) {
				m.EXPECT().GetAgentProfile(gomock.Any(), agentID).Return(domain.AgentProfile{}, errors.New("not found"))
			},
			expectedError: domain.ErrPermissionDenied,
		},
		{
			name:    "Ошибка: низкая линия поддержки (линия 1)",
			agentID: agentID,
			mockInit: func(m *mocks.MockSupportRepository) {
				m.EXPECT().GetAgentProfile(gomock.Any(), agentID).Return(domain.AgentProfile{
					SupportLine: 1,
				}, nil)
			},
			expectedError: domain.ErrPermissionDenied,
		},
		{
			name:    "Ошибка БД при получении статистики",
			agentID: agentID,
			mockInit: func(m *mocks.MockSupportRepository) {
				m.EXPECT().GetAgentProfile(gomock.Any(), agentID).Return(domain.AgentProfile{
					SupportLine: 3,
				}, nil)
				m.EXPECT().GetStats(gomock.Any()).Return(domain.SupportStats{}, errors.New("stats error"))
			},
			expectedError: errutil.Internal("failed to fetch live stats", errors.New("stats error")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockSupportRepository(ctrl)
			tt.mockInit(repo)

			uc := NewSupportUseCase(repo)
			res, err := uc.GetStats(context.Background(), tt.agentID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedRes, res)
			}
		})
	}
}

func TestSupportUseCase_GetMyTickets(t *testing.T) {
	type mockInit func(m *mocks.MockSupportRepository)

	clientID := int64(10)
	guestID := uuid.New()
	tickets := []domain.Ticket{{ID: 1}}

	tests := []struct {
		name          string
		clientID      *int64
		guestID       *uuid.UUID
		mockInit      mockInit
		expectedError error
	}{
		{
			name:     "Успех: получение по clientID",
			clientID: &clientID,
			mockInit: func(m *mocks.MockSupportRepository) {
				m.EXPECT().GetTicketsByClientID(gomock.Any(), clientID).Return(tickets, nil)
			},
		},
		{
			name:    "Успех: получение по guestID",
			guestID: &guestID,
			mockInit: func(m *mocks.MockSupportRepository) {
				m.EXPECT().GetTicketsByGuestID(gomock.Any(), guestID).Return(tickets, nil)
			},
		},
		{
			name:          "Ошибка: нет идентификаторов",
			mockInit:      func(m *mocks.MockSupportRepository) {},
			expectedError: domain.ErrUnauthorized,
		},
		{
			name:     "Ошибка репозитория",
			clientID: &clientID,
			mockInit: func(m *mocks.MockSupportRepository) {
				m.EXPECT().GetTicketsByClientID(gomock.Any(), clientID).Return(nil, errors.New("db fail"))
			},
			expectedError: errors.New("db fail"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockSupportRepository(ctrl)
			tt.mockInit(repo)

			uc := NewSupportUseCase(repo)
			res, err := uc.GetMyTickets(context.Background(), tt.clientID, tt.guestID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Len(t, res, 1)
			}
		})
	}
}

func TestSupportUseCase_AddMessage(t *testing.T) {
	type mockInit func(m *mocks.MockSupportRepository)

	ticketID := uuid.New()
	internalID := int64(100)
	authorID := int64(1)
	idemKey := "key"

	tests := []struct {
		name          string
		role          string
		text          string
		mockInit      mockInit
		expectedError error
	}{
		{
			name: "Успешная отправка (обычный случай)",
			role: "user",
			text: "Hello",
			mockInit: func(m *mocks.MockSupportRepository) {
				m.EXPECT().GetTicketByPublicID(gomock.Any(), ticketID).Return(domain.Ticket{ID: internalID, CurrentStatus: "open"}, nil)
				m.EXPECT().AddMessageEvent(gomock.Any(), internalID, &authorID, "user", "Hello", idemKey).Return(nil)
			},
		},
		{
			name:          "Ошибка: пустое сообщение",
			text:          "",
			mockInit:      func(m *mocks.MockSupportRepository) {},
			expectedError: domain.ErrInvalidMessageInput,
		},
		{
			name: "Ошибка: тикет не найден",
			text: "Hi",
			mockInit: func(m *mocks.MockSupportRepository) {
				m.EXPECT().GetTicketByPublicID(gomock.Any(), ticketID).Return(domain.Ticket{}, domain.ErrTicketNotFound)
			},
			expectedError: domain.ErrTicketNotFound,
		},
		{
			name: "Автоматическое переоткрытие тикета пользователем",
			role: "user",
			text: "I still have a problem",
			mockInit: func(m *mocks.MockSupportRepository) {
				m.EXPECT().GetTicketByPublicID(gomock.Any(), ticketID).Return(domain.Ticket{ID: internalID, CurrentStatus: "closed"}, nil)
				// Ожидаем вызов смены статуса
				m.EXPECT().UpdateTicketStatus(gomock.Any(), internalID, "open", nil, "system", idemKey+"_reopen").Return(nil)
				m.EXPECT().AddMessageEvent(gomock.Any(), internalID, &authorID, "user", "I still have a problem", idemKey).Return(nil)
			},
		},
		{
			name: "Нет переоткрытия, если пишет саппорт",
			role: "support",
			text: "Closing note",
			mockInit: func(m *mocks.MockSupportRepository) {
				m.EXPECT().GetTicketByPublicID(gomock.Any(), ticketID).Return(domain.Ticket{ID: internalID, CurrentStatus: "closed"}, nil)
				m.EXPECT().AddMessageEvent(gomock.Any(), internalID, &authorID, "support", "Closing note", idemKey).Return(nil)
			},
		},
		{
			name: "Ошибка при авто-переоткрытии",
			role: "user",
			text: "Help",
			mockInit: func(m *mocks.MockSupportRepository) {
				m.EXPECT().GetTicketByPublicID(gomock.Any(), ticketID).Return(domain.Ticket{ID: internalID, CurrentStatus: "closed"}, nil)
				m.EXPECT().UpdateTicketStatus(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("fail"))
			},
			expectedError: errutil.Internal("failed to auto-reopen ticket", errors.New("fail")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockSupportRepository(ctrl)
			tt.mockInit(repo)

			uc := NewSupportUseCase(repo)
			err := uc.AddMessage(context.Background(), ticketID, &authorID, tt.role, tt.text, idemKey)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSupportUseCase_GetTicketChat(t *testing.T) {
	type mockInit func(m *mocks.MockSupportRepository)

	ticketID := uuid.New()
	internalID := int64(50)

	tests := []struct {
		name          string
		mockInit      mockInit
		expectedError error
	}{
		{
			name: "Успешное получение чата",
			mockInit: func(m *mocks.MockSupportRepository) {
				m.EXPECT().GetTicketByPublicID(gomock.Any(), ticketID).Return(domain.Ticket{ID: internalID}, nil)
				m.EXPECT().GetEventsByTicketID(gomock.Any(), internalID).Return([]domain.Event{{ID: 1}}, nil)
			},
		},
		{
			name: "Тикет не найден",
			mockInit: func(m *mocks.MockSupportRepository) {
				m.EXPECT().GetTicketByPublicID(gomock.Any(), ticketID).Return(domain.Ticket{}, domain.ErrTicketNotFound)
			},
			expectedError: domain.ErrTicketNotFound,
		},
		{
			name: "Ошибка БД при получении событий",
			mockInit: func(m *mocks.MockSupportRepository) {
				m.EXPECT().GetTicketByPublicID(gomock.Any(), ticketID).Return(domain.Ticket{ID: internalID}, nil)
				m.EXPECT().GetEventsByTicketID(gomock.Any(), internalID).Return(nil, errors.New("query fail"))
			},
			expectedError: errutil.Internal("failed to fetch ticket events from db", errors.New("query fail")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockSupportRepository(ctrl)
			tt.mockInit(repo)

			uc := NewSupportUseCase(repo)
			res, err := uc.GetTicketChat(context.Background(), ticketID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Len(t, res, 1)
			}
		})
	}
}

func TestSupportUseCase_GetTicketEvents(t *testing.T) {
	type mockInit func(m *mocks.MockSupportRepository)

	ticketPublicID := uuid.New()
	internalID := int64(1)
	clientID := int64(42)
	otherClientID := int64(43)
	guestID := uuid.New()
	events := []domain.Event{{ID: 10, TicketID: internalID}}

	tests := []struct {
		name          string
		publicID      uuid.UUID
		clientID      *int64
		guestID       *uuid.UUID
		mockInit      mockInit
		expectedRes   []domain.Event
		expectedError error
	}{
		{
			name:     "Успешное получение событий (владелец ClientID)",
			publicID: ticketPublicID,
			clientID: &clientID,
			mockInit: func(m *mocks.MockSupportRepository) {
				m.EXPECT().GetTicketByPublicID(gomock.Any(), ticketPublicID).Return(domain.Ticket{ID: internalID, ClientID: &clientID}, nil)
				m.EXPECT().GetEventsByTicketID(gomock.Any(), internalID).Return(events, nil)
			},
			expectedRes: events,
		},
		{
			name:     "Успешное получение событий (владелец GuestID)",
			publicID: ticketPublicID,
			guestID:  &guestID,
			mockInit: func(m *mocks.MockSupportRepository) {
				m.EXPECT().GetTicketByPublicID(gomock.Any(), ticketPublicID).Return(domain.Ticket{ID: internalID, GuestID: &guestID}, nil)
				m.EXPECT().GetEventsByTicketID(gomock.Any(), internalID).Return(events, nil)
			},
			expectedRes: events,
		},
		{
			name:     "Ошибка: доступ запрещен (не владелец)",
			publicID: ticketPublicID,
			clientID: &otherClientID,
			mockInit: func(m *mocks.MockSupportRepository) {
				m.EXPECT().GetTicketByPublicID(gomock.Any(), ticketPublicID).Return(domain.Ticket{ID: internalID, ClientID: &clientID}, nil)
			},
			expectedError: domain.ErrPermissionDenied,
		},
		{
			name:     "Ошибка: тикет не найден",
			publicID: ticketPublicID,
			mockInit: func(m *mocks.MockSupportRepository) {
				m.EXPECT().GetTicketByPublicID(gomock.Any(), ticketPublicID).Return(domain.Ticket{}, domain.ErrTicketNotFound)
			},
			expectedError: domain.ErrTicketNotFound,
		},
		{
			name:     "Ошибка БД при получении событий",
			publicID: ticketPublicID,
			clientID: &clientID,
			mockInit: func(m *mocks.MockSupportRepository) {
				m.EXPECT().GetTicketByPublicID(gomock.Any(), ticketPublicID).Return(domain.Ticket{ID: internalID, ClientID: &clientID}, nil)
				m.EXPECT().GetEventsByTicketID(gomock.Any(), internalID).Return(nil, errors.New("db error"))
			},
			expectedError: errutil.Internal("failed to get ticket events", errors.New("db error")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockSupportRepository(ctrl)
			tt.mockInit(repo)

			uc := NewSupportUseCase(repo)
			res, err := uc.GetTicketEvents(context.Background(), tt.publicID, tt.clientID, tt.guestID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedRes, res)
			}
		})
	}
}

func TestSupportUseCase_RateTicket(t *testing.T) {
	type mockInit func(m *mocks.MockSupportRepository)

	ticketPublicID := uuid.New()
	internalID := int64(1)
	clientID := int64(42)
	idemKey := "idem-123"

	tests := []struct {
		name          string
		rating        int
		mockInit      mockInit
		expectedError error
	}{
		{
			name:   "Успешная оценка тикета",
			rating: 5,
			mockInit: func(m *mocks.MockSupportRepository) {
				m.EXPECT().GetTicketByPublicID(gomock.Any(), ticketPublicID).Return(domain.Ticket{ID: internalID, CurrentStatus: "closed"}, nil)
				m.EXPECT().SetTicketRating(gomock.Any(), internalID, 5, &clientID, idemKey).Return(nil)
			},
		},
		{
			name:          "Ошибка: неверный рейтинг (< 1)",
			rating:        0,
			mockInit:      func(m *mocks.MockSupportRepository) {},
			expectedError: domain.ErrInvalidRatingInput,
		},
		{
			name:          "Ошибка: неверный рейтинг (> 5)",
			rating:        6,
			mockInit:      func(m *mocks.MockSupportRepository) {},
			expectedError: domain.ErrInvalidRatingInput,
		},
		{
			name:   "Ошибка: тикет не закрыт",
			rating: 4,
			mockInit: func(m *mocks.MockSupportRepository) {
				m.EXPECT().GetTicketByPublicID(gomock.Any(), ticketPublicID).Return(domain.Ticket{ID: internalID, CurrentStatus: "open"}, nil)
			},
			expectedError: domain.ErrInvalidState,
		},
		{
			name:   "Ошибка: тикет не найден",
			rating: 5,
			mockInit: func(m *mocks.MockSupportRepository) {
				m.EXPECT().GetTicketByPublicID(gomock.Any(), ticketPublicID).Return(domain.Ticket{}, domain.ErrTicketNotFound)
			},
			expectedError: domain.ErrTicketNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockSupportRepository(ctrl)
			tt.mockInit(repo)

			uc := NewSupportUseCase(repo)
			err := uc.RateTicket(context.Background(), ticketPublicID, tt.rating, &clientID, idemKey)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSupportUseCase_ChangeTicketStatus(t *testing.T) {
	type mockInit func(m *mocks.MockSupportRepository)

	ticketPublicID := uuid.New()
	internalID := int64(1)
	agentID := int64(777)
	idemKey := "idem-status"

	tests := []struct {
		name          string
		newStatus     string
		mockInit      mockInit
		expectedError error
	}{
		{
			name:      "Успешное изменение статуса",
			newStatus: "in_progress",
			mockInit: func(m *mocks.MockSupportRepository) {
				m.EXPECT().GetTicketByPublicID(gomock.Any(), ticketPublicID).Return(domain.Ticket{ID: internalID, CurrentStatus: "open"}, nil)
				m.EXPECT().UpdateTicketStatus(gomock.Any(), internalID, "in_progress", &agentID, "support", idemKey).Return(nil)
			},
		},
		{
			name:      "Идемпотентность: статус уже установлен",
			newStatus: "open",
			mockInit: func(m *mocks.MockSupportRepository) {
				m.EXPECT().GetTicketByPublicID(gomock.Any(), ticketPublicID).Return(domain.Ticket{ID: internalID, CurrentStatus: "open"}, nil)
			},
		},
		{
			name:      "Ошибка: тикет не найден",
			newStatus: "closed",
			mockInit: func(m *mocks.MockSupportRepository) {
				m.EXPECT().GetTicketByPublicID(gomock.Any(), ticketPublicID).Return(domain.Ticket{}, domain.ErrTicketNotFound)
			},
			expectedError: domain.ErrTicketNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockSupportRepository(ctrl)
			tt.mockInit(repo)

			uc := NewSupportUseCase(repo)
			err := uc.ChangeTicketStatus(context.Background(), ticketPublicID, tt.newStatus, agentID, idemKey)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSupportUseCase_ReassignTicket(t *testing.T) {
	type mockInit func(m *mocks.MockSupportRepository)

	ticketPublicID := uuid.New()
	internalID := int64(10)
	newAgentID := int64(55)
	authorID := int64(1)
	line := 2
	idemKey := "reassign-key"

	tests := []struct {
		name          string
		mockInit      mockInit
		expectedError error
	}{
		{
			name: "Успешное переназначение тикета",
			mockInit: func(m *mocks.MockSupportRepository) {
				m.EXPECT().GetTicketByPublicID(gomock.Any(), ticketPublicID).
					Return(domain.Ticket{ID: internalID}, nil)
				m.EXPECT().AssignTicket(gomock.Any(), internalID, newAgentID, line, &authorID, "support", idemKey).
					Return(nil)
			},
		},
		{
			name: "Ошибка: тикет не найден",
			mockInit: func(m *mocks.MockSupportRepository) {
				m.EXPECT().GetTicketByPublicID(gomock.Any(), ticketPublicID).
					Return(domain.Ticket{}, errors.New("not found"))
			},
			expectedError: domain.ErrTicketNotFound,
		},
		{
			name: "Ошибка репозитория при назначении",
			mockInit: func(m *mocks.MockSupportRepository) {
				m.EXPECT().GetTicketByPublicID(gomock.Any(), ticketPublicID).
					Return(domain.Ticket{ID: internalID}, nil)
				m.EXPECT().AssignTicket(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New("db error"))
			},
			expectedError: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockSupportRepository(ctrl)
			tt.mockInit(repo)

			uc := NewSupportUseCase(repo)
			err := uc.ReassignTicket(context.Background(), ticketPublicID, newAgentID, line, authorID, idemKey)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSupportUseCase_GetAssignedTickets(t *testing.T) {
	type mockInit func(m *mocks.MockSupportRepository)

	agentID := int64(777)
	tickets := []domain.Ticket{{ID: 1}, {ID: 2}}

	tests := []struct {
		name          string
		mockInit      mockInit
		expectedRes   []domain.Ticket
		expectedError error
	}{
		{
			name: "Успешное получение назначенных тикетов",
			mockInit: func(m *mocks.MockSupportRepository) {
				m.EXPECT().GetAssignedTickets(gomock.Any(), agentID).Return(tickets, nil)
			},
			expectedRes: tickets,
		},
		{
			name: "Ошибка репозитория (Internal)",
			mockInit: func(m *mocks.MockSupportRepository) {
				m.EXPECT().GetAssignedTickets(gomock.Any(), agentID).Return(nil, errors.New("query failed"))
			},
			expectedError: errutil.Internal("failed to get assigned tickets", errors.New("query failed")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockSupportRepository(ctrl)
			tt.mockInit(repo)

			uc := NewSupportUseCase(repo)
			res, err := uc.GetAssignedTickets(context.Background(), agentID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedRes, res)
			}
		})
	}
}

func TestSupportUseCase_SetAgentStatus(t *testing.T) {
	type mockInit func(m *mocks.MockSupportRepository)

	agentID := int64(1)

	tests := []struct {
		name          string
		status        string
		mockInit      mockInit
		expectedError error
	}{
		{
			name:   "Успешная установка статуса online",
			status: "online",
			mockInit: func(m *mocks.MockSupportRepository) {
				m.EXPECT().UpdateAgentStatus(gomock.Any(), agentID, "online").Return(nil)
			},
		},
		{
			name:   "Успешная установка статуса offline",
			status: "offline",
			mockInit: func(m *mocks.MockSupportRepository) {
				m.EXPECT().UpdateAgentStatus(gomock.Any(), agentID, "offline").Return(nil)
			},
		},
		{
			name:          "Ошибка: невалидный статус",
			status:        "away",
			mockInit:      func(m *mocks.MockSupportRepository) {},
			expectedError: domain.ErrInvalidStatusInput,
		},
		{
			name:   "Ошибка репозитория при обновлении",
			status: "online",
			mockInit: func(m *mocks.MockSupportRepository) {
				m.EXPECT().UpdateAgentStatus(gomock.Any(), agentID, "online").Return(errors.New("db error"))
			},
			expectedError: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockSupportRepository(ctrl)
			tt.mockInit(repo)

			uc := NewSupportUseCase(repo)
			err := uc.SetAgentStatus(context.Background(), agentID, tt.status)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSupportUseCase_GetCategories(t *testing.T) {
	type mockInit func(m *mocks.MockSupportRepository)

	categories := []domain.Category{
		{ID: 1, Name: "Технический вопрос", DefaultLine: 1},
		{ID: 2, Name: "Оплата", DefaultLine: 2},
	}

	tests := []struct {
		name          string
		mockInit      mockInit
		expectedRes   []domain.Category
		expectedError error
	}{
		{
			name: "Успешное получение категорий",
			mockInit: func(m *mocks.MockSupportRepository) {
				m.EXPECT().GetActiveCategories(gomock.Any()).Return(categories, nil)
			},
			expectedRes: categories,
		},
		{
			name: "Ошибка репозитория",
			mockInit: func(m *mocks.MockSupportRepository) {
				m.EXPECT().GetActiveCategories(gomock.Any()).Return(nil, errors.New("db error"))
			},
			expectedError: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockSupportRepository(ctrl)
			tt.mockInit(repo)

			uc := NewSupportUseCase(repo)
			res, err := uc.GetCategories(context.Background())

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedRes, res)
			}
		})
	}
}

func TestSupportUseCase_GetTemplates(t *testing.T) {
	type mockInit func(m *mocks.MockSupportRepository)

	templates := []domain.Template{
		{ID: 1, Name: "Приветствие", Content: "Здравствуйте! Чем могу помочь?"},
		{ID: 2, Name: "Прощание", Content: "Всего доброго, обращайтесь еще!"},
	}

	tests := []struct {
		name          string
		mockInit      mockInit
		expectedRes   []domain.Template
		expectedError error
	}{
		{
			name: "Успешное получение шаблонов",
			mockInit: func(m *mocks.MockSupportRepository) {
				m.EXPECT().GetTemplates(gomock.Any()).Return(templates, nil)
			},
			expectedRes: templates,
		},
		{
			name: "Ошибка репозитория",
			mockInit: func(m *mocks.MockSupportRepository) {
				m.EXPECT().GetTemplates(gomock.Any()).Return(nil, errors.New("db failure"))
			},
			expectedError: errors.New("db failure"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mocks.NewMockSupportRepository(ctrl)
			tt.mockInit(repo)

			uc := NewSupportUseCase(repo)
			res, err := uc.GetTemplates(context.Background())

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedRes, res)
			}
		})
	}
}
