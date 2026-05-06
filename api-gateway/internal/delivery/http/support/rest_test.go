package support

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/grpc_client/supportclient"
	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/grpc_client/supportclient/mocks"
	"github.com/go-park-mail-ru/2026_1_NaNcats/api-gateway/internal/middleware"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestSupportHandler_CreateTicket(t *testing.T) {
	type mockInit func(m *mocks.MockSupportClient)

	tests := []struct {
		name           string
		userID         any
		idemKey        string
		body           any
		cookie         *http.Cookie
		mockInit       mockInit
		expectedStatus int
	}{
		{
			name:    "Успешное создание тикета авторизованным пользователем",
			userID:  int64(1),
			idemKey: "idem-key-1",
			body: CreateTicketRequest{
				ContactEmail: "user@mail.ru",
				CategoryID:   1,
				FirstMessage: "Test",
				ClientMeta:   `{"os":"ios"}`,
			},
			mockInit: func(m *mocks.MockSupportClient) {
				m.EXPECT().CreateTicket(gomock.Any(), gomock.Any(), "idem-key-1").
					Return("ticket-uuid", nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:    "Успешное создание тикета гостем",
			userID:  nil,
			idemKey: "idem-key-2",
			body: CreateTicketRequest{
				ContactEmail: "guest@mail.ru",
				CategoryID:   2,
			},
			cookie: &http.Cookie{Name: "guest_id", Value: "guest-uuid"},
			mockInit: func(m *mocks.MockSupportClient) {
				m.EXPECT().CreateTicket(gomock.Any(), gomock.Any(), "idem-key-2").
					Return("ticket-uuid-guest", nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Ошибка: отсутствует Idempotency-Key",
			idemKey:        "",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Ошибка: невалидный JSON",
			idemKey:        "key",
			body:           "{invalid",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:    "Ошибка: сбой gRPC клиента",
			userID:  int64(1),
			idemKey: "key",
			body:    CreateTicketRequest{CategoryID: 1},
			mockInit: func(m *mocks.MockSupportClient) {
				m.EXPECT().CreateTicket(gomock.Any(), gomock.Any(), gomock.Any()).
					Return("", errors.New("grpc error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClient := mocks.NewMockSupportClient(ctrl)
			if tt.mockInit != nil {
				tt.mockInit(mockClient)
			}

			nopLogger := logger.NewNopLogger()
			handler := NewSupportHandler(mockClient, nil, nopLogger)

			var buf bytes.Buffer
			if s, ok := tt.body.(string); ok {
				buf.WriteString(s)
			} else {
				_ = json.NewEncoder(&buf).Encode(tt.body)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/support/tickets", &buf)
			if tt.userID != nil {
				req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, tt.userID))
			}
			if tt.idemKey != "" {
				req.Header.Set("Idempotency-Key", tt.idemKey)
			}
			if tt.cookie != nil {
				req.AddCookie(tt.cookie)
			}

			rec := httptest.NewRecorder()
			handler.CreateTicket(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

func TestSupportHandler_GetMyTickets(t *testing.T) {
	type mockInit func(m *mocks.MockSupportClient)

	tests := []struct {
		name           string
		userID         any
		cookie         *http.Cookie
		mockInit       mockInit
		expectedStatus int
	}{
		{
			name:   "Успешное получение тикетов авторизованным пользователем",
			userID: int64(1),
			mockInit: func(m *mocks.MockSupportClient) {
				m.EXPECT().GetUserTickets(gomock.Any(), gomock.Any(), nil).
					Return([]supportclient.Ticket{
						{ID: 1, PublicID: "uuid-1", CreatedAt: time.Now()},
					}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "Успешное получение тикетов гостем",
			userID: nil,
			cookie: &http.Cookie{Name: "guest_id", Value: "guest-uuid"},
			mockInit: func(m *mocks.MockSupportClient) {
				m.EXPECT().GetUserTickets(gomock.Any(), nil, gomock.Any()).
					Return([]supportclient.Ticket{}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "Ошибка: сбой gRPC при получении",
			userID: int64(1),
			mockInit: func(m *mocks.MockSupportClient) {
				m.EXPECT().GetUserTickets(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, errors.New("fail"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClient := mocks.NewMockSupportClient(ctrl)
			if tt.mockInit != nil {
				tt.mockInit(mockClient)
			}

			nopLogger := logger.NewNopLogger()
			handler := NewSupportHandler(mockClient, nil, nopLogger)

			req := httptest.NewRequest(http.MethodGet, "/api/support/tickets", nil)
			if tt.userID != nil {
				req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, tt.userID))
			}
			if tt.cookie != nil {
				req.AddCookie(tt.cookie)
			}

			rec := httptest.NewRecorder()
			handler.GetMyTickets(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

func TestSupportHandler_GetTicketEvents(t *testing.T) {
	type mockInit func(m *mocks.MockSupportClient)

	tests := []struct {
		name           string
		ticketID       string
		userID         any
		cookie         *http.Cookie
		mockInit       mockInit
		expectedStatus int
	}{
		{
			name:     "Успешное получение событий авторизованным пользователем",
			ticketID: "uuid-1",
			userID:   int64(1),
			mockInit: func(m *mocks.MockSupportClient) {
				m.EXPECT().GetTicketEvents(gomock.Any(), "uuid-1", gomock.Any(), nil).
					Return([]supportclient.Event{
						{ID: 1, TicketID: 1, AuthorRole: "user", CreatedAt: time.Now()},
					}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:     "Ошибка: тикет не найден",
			ticketID: "uuid-404",
			userID:   int64(1),
			mockInit: func(m *mocks.MockSupportClient) {
				m.EXPECT().GetTicketEvents(gomock.Any(), "uuid-404", gomock.Any(), nil).
					Return(nil, supportclient.ErrTicketNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:     "Ошибка: доступ запрещен",
			ticketID: "uuid-no-access",
			userID:   int64(1),
			mockInit: func(m *mocks.MockSupportClient) {
				m.EXPECT().GetTicketEvents(gomock.Any(), "uuid-no-access", gomock.Any(), nil).
					Return(nil, supportclient.ErrUnauthorized)
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:     "Внутренняя ошибка сервиса",
			ticketID: "uuid-1",
			userID:   int64(1),
			mockInit: func(m *mocks.MockSupportClient) {
				m.EXPECT().GetTicketEvents(gomock.Any(), "uuid-1", gomock.Any(), nil).
					Return(nil, errors.New("grpc fail"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClient := mocks.NewMockSupportClient(ctrl)
			if tt.mockInit != nil {
				tt.mockInit(mockClient)
			}

			handler := NewSupportHandler(mockClient, nil, logger.NewNopLogger())

			req := httptest.NewRequest(http.MethodGet, "/api/support/tickets/"+tt.ticketID+"/events", nil)
			req.SetPathValue("id", tt.ticketID)
			if tt.userID != nil {
				req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, tt.userID))
			}
			if tt.cookie != nil {
				req.AddCookie(tt.cookie)
			}

			rec := httptest.NewRecorder()
			handler.GetTicketEvents(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

func TestSupportHandler_RateTicket(t *testing.T) {
	type mockInit func(m *mocks.MockSupportClient)

	tests := []struct {
		name           string
		ticketID       string
		userID         any
		idemKey        string
		body           any
		mockInit       mockInit
		expectedStatus int
	}{
		{
			name:     "Успешная оценка тикета",
			ticketID: "uuid-1",
			userID:   int64(1),
			idemKey:  "key-1",
			body:     RateTicketRequest{Rating: 5},
			mockInit: func(m *mocks.MockSupportClient) {
				m.EXPECT().RateTicket(gomock.Any(), "uuid-1", 5, gomock.Any(), "key-1").
					Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Ошибка: отсутствует Idempotency-Key",
			ticketID:       "uuid-1",
			idemKey:        "",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Ошибка: некорректный JSON",
			ticketID:       "uuid-1",
			userID:         int64(1),
			idemKey:        "key-1",
			body:           "invalid",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:     "Ошибка на стороне сервиса",
			ticketID: "uuid-1",
			userID:   int64(1),
			idemKey:  "key-1",
			body:     RateTicketRequest{Rating: 4},
			mockInit: func(m *mocks.MockSupportClient) {
				m.EXPECT().RateTicket(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New("fail"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClient := mocks.NewMockSupportClient(ctrl)
			if tt.mockInit != nil {
				tt.mockInit(mockClient)
			}

			handler := NewSupportHandler(mockClient, nil, logger.NewNopLogger())

			var buf bytes.Buffer
			if s, ok := tt.body.(string); ok {
				buf.WriteString(s)
			} else {
				_ = json.NewEncoder(&buf).Encode(tt.body)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/support/tickets/"+tt.ticketID+"/rate", &buf)
			req.SetPathValue("id", tt.ticketID)
			if tt.userID != nil {
				req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, tt.userID))
			}
			if tt.idemKey != "" {
				req.Header.Set("Idempotency-Key", tt.idemKey)
			}

			rec := httptest.NewRecorder()
			handler.RateTicket(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

func TestSupportHandler_GetAssignedTickets(t *testing.T) {
	type mockInit func(m *mocks.MockSupportClient)

	tests := []struct {
		name           string
		userID         any
		mockInit       mockInit
		expectedStatus int
	}{
		{
			name:   "Успешное получение назначенных тикетов",
			userID: int64(10),
			mockInit: func(m *mocks.MockSupportClient) {
				m.EXPECT().GetAssignedTickets(gomock.Any(), int64(10)).
					Return([]supportclient.Ticket{
						{ID: 100, PublicID: "uuid-100", CreatedAt: time.Now()},
					}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Ошибка: не авторизован",
			userID:         nil,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:   "Ошибка gRPC при получении списка",
			userID: int64(10),
			mockInit: func(m *mocks.MockSupportClient) {
				m.EXPECT().GetAssignedTickets(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("internal error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClient := mocks.NewMockSupportClient(ctrl)
			if tt.mockInit != nil {
				tt.mockInit(mockClient)
			}

			handler := NewSupportHandler(mockClient, nil, logger.NewNopLogger())

			req := httptest.NewRequest(http.MethodGet, "/api/admin/support/tickets", nil)
			if tt.userID != nil {
				req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, tt.userID))
			}

			rec := httptest.NewRecorder()
			handler.GetAssignedTickets(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

func TestSupportHandler_ChangeTicketStatus(t *testing.T) {
	type mockInit func(m *mocks.MockSupportClient)

	tests := []struct {
		name           string
		userID         any
		ticketID       string
		idemKey        string
		body           any
		mockInit       mockInit
		expectedStatus int
	}{
		{
			name:     "Успешная смена статуса тикета",
			userID:   int64(10),
			ticketID: "ticket-uuid",
			idemKey:  "idem-1",
			body:     ChangeStatusRequest{Status: "in_progress"},
			mockInit: func(m *mocks.MockSupportClient) {
				m.EXPECT().ChangeTicketStatus(gomock.Any(), "ticket-uuid", "in_progress", int64(10), "idem-1").
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
			name:           "Ошибка: отсутствует Idempotency-Key",
			userID:         int64(10),
			idemKey:        "",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Ошибка: невалидный JSON",
			userID:         int64(10),
			idemKey:        "idem-2",
			body:           "invalid-json",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:     "Ошибка: сбой gRPC сервиса",
			userID:   int64(10),
			ticketID: "ticket-uuid",
			idemKey:  "idem-3",
			body:     ChangeStatusRequest{Status: "closed"},
			mockInit: func(m *mocks.MockSupportClient) {
				m.EXPECT().ChangeTicketStatus(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New("internal fail"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClient := mocks.NewMockSupportClient(ctrl)
			if tt.mockInit != nil {
				tt.mockInit(mockClient)
			}

			handler := NewSupportHandler(mockClient, nil, logger.NewNopLogger())

			var buf bytes.Buffer
			_ = json.NewEncoder(&buf).Encode(tt.body)

			req := httptest.NewRequest(http.MethodPatch, "/api/admin/support/tickets/"+tt.ticketID+"/status", &buf)
			if tt.userID != nil {
				req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, tt.userID))
			}
			if tt.ticketID != "" {
				req.SetPathValue("id", tt.ticketID)
			}
			if tt.idemKey != "" {
				req.Header.Set("Idempotency-Key", tt.idemKey)
			}

			rec := httptest.NewRecorder()
			handler.ChangeTicketStatus(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

func TestSupportHandler_ReassignTicket(t *testing.T) {
	type mockInit func(m *mocks.MockSupportClient)

	tests := []struct {
		name           string
		userID         any
		ticketID       string
		idemKey        string
		body           any
		mockInit       mockInit
		expectedStatus int
	}{
		{
			name:     "Успешное переназначение тикета",
			userID:   int64(10),
			ticketID: "ticket-uuid",
			idemKey:  "idem-reassign",
			body:     ReassignTicketRequest{AgentID: 20, Line: 2},
			mockInit: func(m *mocks.MockSupportClient) {
				m.EXPECT().ReassignTicket(gomock.Any(), "ticket-uuid", int64(20), 2, int64(10), "idem-reassign").
					Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:     "Ошибка gRPC при переназначении",
			userID:   int64(10),
			ticketID: "ticket-uuid",
			idemKey:  "idem-fail",
			body:     ReassignTicketRequest{AgentID: 20, Line: 2},
			mockInit: func(m *mocks.MockSupportClient) {
				m.EXPECT().ReassignTicket(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(errors.New("fail"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClient := mocks.NewMockSupportClient(ctrl)
			if tt.mockInit != nil {
				tt.mockInit(mockClient)
			}

			handler := NewSupportHandler(mockClient, nil, logger.NewNopLogger())

			var buf bytes.Buffer
			_ = json.NewEncoder(&buf).Encode(tt.body)

			req := httptest.NewRequest(http.MethodPost, "/api/admin/support/tickets/"+tt.ticketID+"/reassign", &buf)
			if tt.userID != nil {
				req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, tt.userID))
			}
			req.SetPathValue("id", tt.ticketID)
			req.Header.Set("Idempotency-Key", tt.idemKey)

			rec := httptest.NewRecorder()
			handler.ReassignTicket(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

func TestSupportHandler_SetAgentStatus(t *testing.T) {
	type mockInit func(m *mocks.MockSupportClient)

	tests := []struct {
		name           string
		userID         any
		body           any
		mockInit       mockInit
		expectedStatus int
	}{
		{
			name:   "Успешная установка статуса агента",
			userID: int64(100),
			body:   SetAgentStatusRequest{Status: "online"},
			mockInit: func(m *mocks.MockSupportClient) {
				m.EXPECT().SetAgentStatus(gomock.Any(), int64(100), "online").
					Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "Ошибка UseCase при установке статуса",
			userID: int64(100),
			body:   SetAgentStatusRequest{Status: "offline"},
			mockInit: func(m *mocks.MockSupportClient) {
				m.EXPECT().SetAgentStatus(gomock.Any(), int64(100), "offline").
					Return(errors.New("db error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClient := mocks.NewMockSupportClient(ctrl)
			if tt.mockInit != nil {
				tt.mockInit(mockClient)
			}

			handler := NewSupportHandler(mockClient, nil, logger.NewNopLogger())

			var buf bytes.Buffer
			_ = json.NewEncoder(&buf).Encode(tt.body)

			req := httptest.NewRequest(http.MethodPatch, "/api/admin/support/agent/status", &buf)
			if tt.userID != nil {
				req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, tt.userID))
			}

			rec := httptest.NewRecorder()
			handler.SetAgentStatus(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

func TestSupportHandler_GetCategories(t *testing.T) {
	type mockInit func(m *mocks.MockSupportClient)

	tests := []struct {
		name           string
		mockInit       mockInit
		expectedStatus int
	}{
		{
			name: "Успешное получение категорий",
			mockInit: func(m *mocks.MockSupportClient) {
				m.EXPECT().GetCategories(gomock.Any()).Return([]supportclient.Category{
					{ID: 1, Name: "Технический вопрос", DefaultLine: 1},
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Ошибка gRPC при получении категорий",
			mockInit: func(m *mocks.MockSupportClient) {
				m.EXPECT().GetCategories(gomock.Any()).Return(nil, errors.New("rpc error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClient := mocks.NewMockSupportClient(ctrl)
			tt.mockInit(mockClient)

			handler := NewSupportHandler(mockClient, nil, logger.NewNopLogger())

			req := httptest.NewRequest(http.MethodGet, "/api/support/categories", nil)
			rec := httptest.NewRecorder()

			handler.GetCategories(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

func TestSupportHandler_GetTemplates(t *testing.T) {
	type mockInit func(m *mocks.MockSupportClient)

	tests := []struct {
		name           string
		mockInit       mockInit
		expectedStatus int
	}{
		{
			name: "Успешное получение шаблонов",
			mockInit: func(m *mocks.MockSupportClient) {
				m.EXPECT().GetTemplates(gomock.Any()).Return([]supportclient.Template{
					{ID: 1, Name: "Приветствие", Content: "Здравствуйте!"},
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Ошибка gRPC при получении шаблонов",
			mockInit: func(m *mocks.MockSupportClient) {
				m.EXPECT().GetTemplates(gomock.Any()).Return(nil, errors.New("fail"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClient := mocks.NewMockSupportClient(ctrl)
			tt.mockInit(mockClient)

			handler := NewSupportHandler(mockClient, nil, logger.NewNopLogger())

			req := httptest.NewRequest(http.MethodGet, "/api/admin/support/templates", nil)
			rec := httptest.NewRecorder()

			handler.GetTemplates(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

func TestSupportHandler_GetSupportStats(t *testing.T) {
	type mockInit func(m *mocks.MockSupportClient)

	tests := []struct {
		name           string
		role           any
		mockInit       mockInit
		expectedStatus int
	}{
		{
			name: "Успешное получение статистики админом",
			role: "admin",
			mockInit: func(m *mocks.MockSupportClient) {
				m.EXPECT().GetStats(gomock.Any()).Return(supportclient.SupportStats{TotalTickets: 100}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Успешное получение статистики сотрудником саппорта",
			role: "support",
			mockInit: func(m *mocks.MockSupportClient) {
				m.EXPECT().GetStats(gomock.Any()).Return(supportclient.SupportStats{TotalTickets: 50}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Ошибка доступа: роль обычного пользователя",
			role:           "user",
			mockInit:       func(m *mocks.MockSupportClient) {},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Ошибка доступа: роль отсутствует в контексте",
			role:           nil,
			mockInit:       func(m *mocks.MockSupportClient) {},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "Ошибка gRPC при получении статистики",
			role: "admin",
			mockInit: func(m *mocks.MockSupportClient) {
				m.EXPECT().GetStats(gomock.Any()).Return(supportclient.SupportStats{}, errors.New("internal"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClient := mocks.NewMockSupportClient(ctrl)
			tt.mockInit(mockClient)

			handler := NewSupportHandler(mockClient, nil, logger.NewNopLogger())

			req := httptest.NewRequest(http.MethodGet, "/api/admin/support/stats", nil)
			if tt.role != nil {
				req = req.WithContext(context.WithValue(req.Context(), middleware.RoleKey, tt.role))
			}

			rec := httptest.NewRecorder()
			handler.GetSupportStats(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

func TestSupportHandler_getOrSetGuestID(t *testing.T) {
	handler := NewSupportHandler(nil, nil, logger.NewNopLogger())

	t.Run("Возврат существующего guest_id из кук", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.AddCookie(&http.Cookie{Name: "guest_id", Value: "existing-uuid"})
		rec := httptest.NewRecorder()

		id := handler.getOrSetGuestID(rec, req)

		assert.Equal(t, "existing-uuid", id)
		assert.Empty(t, rec.Header().Get("Set-Cookie"))
	})

	t.Run("Генерация и установка нового guest_id", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		id := handler.getOrSetGuestID(rec, req)

		assert.NotEmpty(t, id)
		assert.Contains(t, rec.Header().Get("Set-Cookie"), "guest_id="+id)
	})
}
