package postgres

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/support/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pashagolub/pgxmock/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSupportRepo_CreateTicket(t *testing.T) {
	ctx := context.Background()
	ticketPublicID := uuid.New()
	idempotencyKey := "unique-key-123"

	input := domain.CreateTicketInput{
		ContactEmail:   "test@mail.ru",
		CategoryID:     1,
		SupportLine:    1,
		FirstMessage:   "Hello",
		CreatorRole:    "user",
		IdempotencyKey: idempotencyKey,
	}

	tests := []struct {
		name          string
		mockInit      func(m pgxmock.PgxPoolIface)
		expectedID    uuid.UUID
		expectedError string
	}{
		{
			name: "Успешное создание тикета",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()
				m.ExpectQuery(`INSERT INTO "support_ticket"`).
					WithArgs(
						input.ClientID, input.GuestID, input.ContactEmail,
						input.CategoryID, input.SupportLine, input.CreatorRole,
						input.ClientMeta, input.IdempotencyKey,
					).
					WillReturnRows(pgxmock.NewRows([]string{"id", "public_id"}).AddRow(int64(1), ticketPublicID))

				m.ExpectExec(`INSERT INTO "support_event"`).
					WithArgs(int64(1), input.ClientID, input.CreatorRole, pgxmock.AnyArg()).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))

				m.ExpectCommit()
			},
			expectedID: ticketPublicID,
		},
		{
			name: "Ошибка: нарушение уникальности (идемпотентность)",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()
				m.ExpectQuery(`INSERT INTO "support_ticket"`).
					WithArgs(
						input.ClientID, input.GuestID, input.ContactEmail,
						input.CategoryID, input.SupportLine, input.CreatorRole,
						input.ClientMeta, input.IdempotencyKey,
					).
					WillReturnError(&pgconn.PgError{Code: pgerrcode.UniqueViolation})
				m.ExpectRollback()
			},
			expectedID:    uuid.Nil,
			expectedError: "ticket with this idempotency key already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			repo := NewSupportRepo(mock)
			tt.mockInit(mock)

			res, err := repo.CreateTicket(ctx, input)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedID, res)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestSupportRepo_AddEvent(t *testing.T) {
	ctx := context.Background()
	input := domain.AddEventInput{
		TicketID:       1,
		AuthorRole:     "support",
		EventType:      "message",
		Payload:        []byte(`{"text": "reply"}`),
		IdempotencyKey: "evt-key",
	}

	tests := []struct {
		name          string
		mockInit      func(m pgxmock.PgxPoolIface)
		expectedError bool
	}{
		{
			name: "Успешное добавление события",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()
				m.ExpectExec(`INSERT INTO "support_event"`).
					WithArgs(input.TicketID, input.AuthorID, input.AuthorRole, input.EventType, input.Payload, input.IdempotencyKey).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
				m.ExpectExec(`UPDATE "support_ticket" SET updated_at = NOW\(\) WHERE id = \$1`).
					WithArgs(input.TicketID).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
				m.ExpectCommit()
			},
		},
		{
			name: "Идемпотентность: событие уже существует",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()
				m.ExpectExec(`INSERT INTO "support_event"`).
					WithArgs(input.TicketID, input.AuthorID, input.AuthorRole, input.EventType, input.Payload, input.IdempotencyKey).
					WillReturnError(&pgconn.PgError{Code: pgerrcode.UniqueViolation})
				m.ExpectRollback()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			repo := NewSupportRepo(mock)
			tt.mockInit(mock)

			err = repo.AddEvent(ctx, input)
			assert.NoError(t, err)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestSupportRepo_UpdateTicketStatus(t *testing.T) {
	ctx := context.Background()
	var ticketID int64 = 1
	newStatus := "closed"

	tests := []struct {
		name          string
		mockInit      func(m pgxmock.PgxPoolIface)
		expectedError string
	}{
		{
			name: "Успешное обновление статуса",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()
				m.ExpectQuery(`SELECT current_status FROM "support_ticket" WHERE id = \$1 FOR UPDATE`).
					WithArgs(ticketID).
					WillReturnRows(pgxmock.NewRows([]string{"current_status"}).AddRow("open"))
				m.ExpectExec(`UPDATE "support_ticket" SET current_status = \$1, updated_at = NOW\(\) WHERE id = \$2`).
					WithArgs(newStatus, ticketID).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))

				// saveEventTx лоигка
				m.ExpectExec(`INSERT INTO "support_event"`).
					WithArgs(ticketID, pgxmock.AnyArg(), "system", "status_changed", pgxmock.AnyArg(), "").
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
				m.ExpectExec(`UPDATE "support_ticket" SET updated_at = NOW\(\) WHERE id = \$1`).
					WithArgs(ticketID).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))

				m.ExpectCommit()
			},
		},
		{
			name: "Тикет не найден",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()
				m.ExpectQuery(`SELECT current_status`).
					WithArgs(ticketID).
					WillReturnError(pgx.ErrNoRows)
				m.ExpectRollback()
			},
			expectedError: "ticket not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, _ := pgxmock.NewPool()
			defer mock.Close()

			repo := NewSupportRepo(mock)
			tt.mockInit(mock)

			err := repo.UpdateTicketStatus(ctx, ticketID, newStatus, nil, "system", "")

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestSupportRepo_AssignTicket(t *testing.T) {
	ctx := context.Background()
	var ticketID int64 = 10
	var agentID int64 = 5
	line := 2

	tests := []struct {
		name     string
		mockInit func(m pgxmock.PgxPoolIface)
		wantErr  bool
	}{
		{
			name: "Успешное переназначение",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()
				m.ExpectQuery(`SELECT assignee_id, support_line FROM "support_ticket" WHERE id = \$1`).
					WithArgs(ticketID).
					WillReturnRows(pgxmock.NewRows([]string{"assignee_id", "support_line"}).AddRow(nil, 1))

				// saveEventTx calls
				m.ExpectExec(`INSERT INTO "support_event"`).
					WithArgs(ticketID, pgxmock.AnyArg(), "support", "reassigned", pgxmock.AnyArg(), "").
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
				m.ExpectExec(`UPDATE "support_ticket" SET updated_at = NOW\(\) WHERE id = \$1`).
					WithArgs(ticketID).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))

				m.ExpectCommit()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, _ := pgxmock.NewPool()
			defer mock.Close()

			repo := NewSupportRepo(mock)
			tt.mockInit(mock)

			err := repo.AssignTicket(ctx, ticketID, agentID, line, nil, "support", "")
			assert.NoError(t, err)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestSupportRepo_GetTicketByPublicID(t *testing.T) {
	ctx := context.Background()
	publicID := uuid.New()
	columns := []string{
		"id", "public_id", "client_account_id", "guest_id", "contact_email", "category_id",
		"current_status", "support_line", "assignee_id", "resolution_rating",
		"client_meta", "creator_role", "idempotency_key", "created_at", "updated_at",
	}

	tests := []struct {
		name          string
		mockInit      func(m pgxmock.PgxPoolIface)
		expectedError string
	}{
		{
			name: "Успешное получение тикета",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(`SELECT (.+) FROM "support_ticket" WHERE public_id = \$1`).
					WithArgs(publicID).
					WillReturnRows(pgxmock.NewRows(columns).
						AddRow(int64(1), publicID, nil, nil, "user@mail.ru", int64(1), "open", 1, nil, nil, []byte("{}"), "user", nil, time.Now(), time.Now()))
			},
		},
		{
			name: "Тикет не найден",
			mockInit: func(m pgxmock.PgxPoolIface) {
				// Чтобы CollectOneRow вернул ErrNoRows, нужно вернуть пустой набор строк, а не ошибку из Query
				m.ExpectQuery(`SELECT (.+) FROM "support_ticket" WHERE public_id = \$1`).
					WithArgs(publicID).
					WillReturnRows(pgxmock.NewRows(columns))
			},
			expectedError: "ticket not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, _ := pgxmock.NewPool()
			defer mock.Close()

			repo := NewSupportRepo(mock)
			tt.mockInit(mock)

			_, err := repo.GetTicketByPublicID(ctx, publicID)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Equal(t, tt.expectedError, err.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestSupportRepo_GetTicketsByClientID(t *testing.T) {
	ctx := context.Background()
	var clientID int64 = 123
	columns := []string{
		"id", "public_id", "client_account_id", "guest_id", "contact_email", "category_id",
		"current_status", "support_line", "assignee_id", "resolution_rating",
		"client_meta", "creator_role", "idempotency_key", "created_at", "updated_at",
	}

	tests := []struct {
		name     string
		mockInit func(m pgxmock.PgxPoolIface)
		count    int
		wantErr  bool
	}{
		{
			name: "Получение списка тикетов клиента",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(`SELECT (.+) FROM "support_ticket" WHERE client_account_id = \$1`).
					WithArgs(clientID).
					WillReturnRows(pgxmock.NewRows(columns).
						AddRow(int64(1), uuid.New(), &clientID, nil, "a@b.com", int64(1), "open", 1, nil, nil, []byte("{}"), "user", nil, time.Now(), time.Now()).
						AddRow(int64(2), uuid.New(), &clientID, nil, "a@b.com", int64(1), "closed", 1, nil, nil, []byte("{}"), "user", nil, time.Now(), time.Now()))
			},
			count: 2,
		},
		{
			name: "Ошибка запроса",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(`SELECT`).WillReturnError(fmt.Errorf("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, _ := pgxmock.NewPool()
			defer mock.Close()

			repo := NewSupportRepo(mock)
			tt.mockInit(mock)

			res, err := repo.GetTicketsByClientID(ctx, clientID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, res, tt.count)
			}
		})
	}
}

func TestSupportRepo_GetTicketsByGuestID(t *testing.T) {
	ctx := context.Background()
	guestID := uuid.New()

	tests := []struct {
		name     string
		mockInit func(m pgxmock.PgxPoolIface)
		wantErr  bool
	}{
		{
			name: "Успешное получение тикетов гостя",
			mockInit: func(m pgxmock.PgxPoolIface) {
				// Используем пустой результат для простоты
				m.ExpectQuery(`SELECT (.+) FROM "support_ticket" WHERE guest_id = \$1`).
					WithArgs(guestID).
					WillReturnRows(pgxmock.NewRows([]string{"id"}))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, _ := pgxmock.NewPool()
			defer mock.Close()

			repo := NewSupportRepo(mock)
			tt.mockInit(mock)

			_, err := repo.GetTicketsByGuestID(ctx, guestID)
			assert.NoError(t, err)
		})
	}
}

func TestSupportRepo_GetAssignedTickets(t *testing.T) {
	ctx := context.Background()
	var agentID int64 = 777

	tests := []struct {
		name     string
		mockInit func(m pgxmock.PgxPoolIface)
		wantErr  bool
	}{
		{
			name: "Получение назначенных на агента тикетов",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(`SELECT (.+) FROM "support_ticket" WHERE assignee_id = \$1 AND current_status != 'closed'`).
					WithArgs(agentID).
					WillReturnRows(pgxmock.NewRows([]string{"id"}))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, _ := pgxmock.NewPool()
			defer mock.Close()

			repo := NewSupportRepo(mock)
			tt.mockInit(mock)

			_, err := repo.GetAssignedTickets(ctx, agentID)
			assert.NoError(t, err)
		})
	}
}

func TestSupportRepo_GetEventsByTicketID(t *testing.T) {
	ctx := context.Background()
	var ticketID int64 = 10
	columns := []string{"id", "ticket_id", "author_account_id", "author_role", "event_type", "payload", "idempotency_key", "created_at"}

	tests := []struct {
		name     string
		mockInit func(m pgxmock.PgxPoolIface)
		count    int
		wantErr  bool
	}{
		{
			name: "Успешное получение истории событий",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(`SELECT (.+) FROM "support_event" WHERE ticket_id = \$1 ORDER BY created_at ASC`).
					WithArgs(ticketID).
					WillReturnRows(pgxmock.NewRows(columns).
						AddRow(int64(100), ticketID, nil, "user", "ticket_created", []byte("{}"), nil, time.Now()).
						AddRow(int64(101), ticketID, nil, "support", "message", []byte(`{"text":"hi"}`), nil, time.Now()))
			},
			count: 2,
		},
		{
			name: "Ошибка при получении событий",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(`SELECT`).WillReturnError(errors.New("scan error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, _ := pgxmock.NewPool()
			defer mock.Close()

			repo := NewSupportRepo(mock)
			tt.mockInit(mock)

			res, err := repo.GetEventsByTicketID(ctx, ticketID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, res, tt.count)
				if tt.count > 0 {
					assert.Equal(t, "user", res[0].AuthorRole)
				}
			}
		})
	}
}

func TestSupportRepo_UpdateAgentStatus(t *testing.T) {
	ctx := context.Background()
	var agentID int64 = 1
	status := "online"

	type mockInit func(m pgxmock.PgxPoolIface)
	tests := []struct {
		name     string
		mockInit mockInit
		wantErr  bool
	}{
		{
			name: "Успешный апдейт/инсерт статуса",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectExec(`INSERT INTO "support_agent_profile"`).
					WithArgs(agentID, status).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
			},
			wantErr: false,
		},
		{
			name: "Ошибка базы данных",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectExec(`INSERT INTO "support_agent_profile"`).
					WithArgs(agentID, status).
					WillReturnError(errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			repo := NewSupportRepo(mock)
			tt.mockInit(mock)

			err = repo.UpdateAgentStatus(ctx, agentID, status)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestSupportRepo_GetActiveCategories(t *testing.T) {
	ctx := context.Background()
	columns := []string{"id", "name", "description", "default_line", "is_active"}

	tests := []struct {
		name     string
		mockInit func(m pgxmock.PgxPoolIface)
		count    int
		wantErr  bool
	}{
		{
			name: "Успешное получение списка категорий",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(`SELECT (.+) FROM "support_category" WHERE is_active = true`).
					WillReturnRows(pgxmock.NewRows(columns).
						AddRow(int64(1), "Tech", "Tech support", 2, true).
						AddRow(int64(2), "Billing", "Payment issues", 1, true))
			},
			count: 2,
		},
		{
			name: "Пустой список",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(`SELECT`).WillReturnRows(pgxmock.NewRows(columns))
			},
			count: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, _ := pgxmock.NewPool()
			defer mock.Close()

			repo := NewSupportRepo(mock)
			tt.mockInit(mock)

			res, err := repo.GetActiveCategories(ctx)
			assert.NoError(t, err)
			assert.Len(t, res, tt.count)
		})
	}
}

func TestSupportRepo_GetAgentProfile(t *testing.T) {
	ctx := context.Background()
	var agentID int64 = 42
	columns := []string{"account_id", "status", "support_line", "created_at", "updated_at"}

	tests := []struct {
		name          string
		mockInit      func(m pgxmock.PgxPoolIface)
		expectedError string
	}{
		{
			name: "Успешное получение профиля",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(`SELECT (.+) FROM "support_agent_profile" WHERE account_id = \$1`).
					WithArgs(agentID).
					WillReturnRows(pgxmock.NewRows(columns).
						AddRow(agentID, "online", 2, time.Now(), time.Now()))
			},
		},
		{
			name: "Профиль не найден",
			mockInit: func(m pgxmock.PgxPoolIface) {
				// Пустые строки для срабатывания CollectOneRow -> ErrNoRows
				m.ExpectQuery(`SELECT (.+) FROM "support_agent_profile" WHERE account_id = \$1`).
					WithArgs(agentID).
					WillReturnRows(pgxmock.NewRows(columns))
			},
			expectedError: "agent profile not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, _ := pgxmock.NewPool()
			defer mock.Close()

			repo := NewSupportRepo(mock)
			tt.mockInit(mock)

			_, err := repo.GetAgentProfile(ctx, agentID)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestSupportRepo_SetTicketRating(t *testing.T) {
	ctx := context.Background()
	var ticketID int64 = 100
	var authorID int64 = 5
	rating := 5
	idemKey := "rating-key"

	tests := []struct {
		name          string
		mockInit      func(m pgxmock.PgxPoolIface)
		expectedError string
	}{
		{
			name: "Успешная установка рейтинга",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()
				// Апдейт рейтинга в мастере
				m.ExpectExec(`UPDATE "support_ticket" SET resolution_rating = \$1, updated_at = NOW\(\) WHERE id = \$2`).
					WithArgs(rating, ticketID).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))

				// Вставка события (в этом методе 4 аргумента, т.к. роль и тип захардкожены в SQL)
				m.ExpectExec(`INSERT INTO "support_event"`).
					WithArgs(ticketID, &authorID, pgxmock.AnyArg(), idemKey).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))

				m.ExpectCommit()
			},
		},
		{
			name: "Тикет не найден",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()
				m.ExpectExec(`UPDATE "support_ticket" SET resolution_rating = \$1`).
					WithArgs(rating, ticketID).
					WillReturnResult(pgxmock.NewResult("UPDATE", 0))
				m.ExpectRollback()
			},
			expectedError: "ticket not found",
		},
		{
			name: "Ошибка при вставке события (откат транзакции)",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()
				m.ExpectExec(`UPDATE "support_ticket" SET resolution_rating = \$1`).
					WithArgs(rating, ticketID).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))

				m.ExpectExec(`INSERT INTO "support_event"`).
					WithArgs(ticketID, &authorID, pgxmock.AnyArg(), idemKey).
					WillReturnError(errors.New("fatal insert error"))
				m.ExpectRollback()
			},
			expectedError: "fatal insert error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			repo := NewSupportRepo(mock)
			tt.mockInit(mock)

			err = repo.SetTicketRating(ctx, ticketID, rating, &authorID, idemKey)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestSupportRepo_GetTemplates(t *testing.T) {
	ctx := context.Background()

	type mockInit func(m pgxmock.PgxPoolIface)
	tests := []struct {
		name     string
		mockInit mockInit
		want     []domain.Template
		wantErr  bool
	}{
		{
			name: "Успешное получение шаблонов",
			mockInit: func(m pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "name", "content"}).
					AddRow(int64(1), "Приветствие", "Здравствуйте!").
					AddRow(int64(2), "Прощание", "До свидания!")
				m.ExpectQuery(`SELECT id, name, content FROM "support_template"`).
					WillReturnRows(rows)
			},
			want: []domain.Template{
				{ID: 1, Name: "Приветствие", Content: "Здравствуйте!"},
				{ID: 2, Name: "Прощание", Content: "До свидания!"},
			},
			wantErr: false,
		},
		{
			name: "Ошибка выполнения запроса",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(`SELECT id, name, content`).
					WillReturnError(errors.New("db error"))
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			repo := NewSupportRepo(mock)
			tt.mockInit(mock)

			got, err := repo.GetTemplates(ctx)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestSupportRepo_AddMessageEvent(t *testing.T) {
	ctx := context.Background()
	var ticketID int64 = 1
	var authorID int64 = 10
	authorRole := "support"
	text := "Test message"
	idemKey := "msg-key"

	tests := []struct {
		name          string
		mockInit      func(m pgxmock.PgxPoolIface)
		expectedError string
	}{
		{
			name: "Успешная отправка сообщения",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()
				m.ExpectExec(`INSERT INTO "support_event"`).
					WithArgs(ticketID, &authorID, authorRole, "message", pgxmock.AnyArg(), idemKey).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))

				m.ExpectExec(`UPDATE "support_ticket" SET updated_at = NOW\(\) WHERE id = \$1`).
					WithArgs(ticketID).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))

				m.ExpectCommit()
			},
		},
		{
			name: "Ошибка: нарушение уникальности ключа идемпотентности",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()
				// В коде при UniqueViolation возвращается nil (игнорируем дубликат)
				m.ExpectExec(`INSERT INTO "support_event"`).
					WithArgs(ticketID, &authorID, authorRole, "message", pgxmock.AnyArg(), idemKey).
					WillReturnError(&pgconn.PgError{Code: pgerrcode.UniqueViolation})
				m.ExpectCommit()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			repo := NewSupportRepo(mock)
			tt.mockInit(mock)

			err = repo.AddMessageEvent(ctx, ticketID, &authorID, authorRole, text, idemKey)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestSupportRepo_AddStatusChangedEvent(t *testing.T) {
	ctx := context.Background()
	var ticketID int64 = 1
	authorRole := "system"
	oldStatus := "open"
	newStatus := "closed"
	reason := "resolved"
	idemKey := "status-key"

	tests := []struct {
		name          string
		mockInit      func(m pgxmock.PgxPoolIface)
		expectedError string
	}{
		{
			name: "Успешная смена статуса с событием",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()
				m.ExpectExec(`UPDATE "support_ticket" SET current_status = \$1`).
					WithArgs(newStatus, ticketID).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))

				m.ExpectExec(`INSERT INTO "support_event"`).
					WithArgs(ticketID, pgxmock.AnyArg(), authorRole, "status_changed", pgxmock.AnyArg(), idemKey).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))

				m.ExpectExec(`UPDATE "support_ticket" SET updated_at = NOW\(\) WHERE id = \$1`).
					WithArgs(ticketID).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))

				m.ExpectCommit()
			},
		},
		{
			name: "Ошибка при обновлении статуса",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()
				m.ExpectExec(`UPDATE "support_ticket" SET current_status`).
					WithArgs(newStatus, ticketID).
					WillReturnError(errors.New("db error"))
				m.ExpectRollback()
			},
			expectedError: "update current_status: db error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, _ := pgxmock.NewPool()
			defer mock.Close()

			repo := NewSupportRepo(mock)
			tt.mockInit(mock)

			err := repo.AddStatusChangedEvent(ctx, ticketID, nil, authorRole, oldStatus, newStatus, reason, idemKey)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestSupportRepo_GetStats(t *testing.T) {
	ctx := context.Background()

	type mockInit func(m pgxmock.PgxPoolIface)
	tests := []struct {
		name          string
		mockInit      mockInit
		want          domain.SupportStats
		expectedError bool
	}{
		{
			name: "Успешное получение статистики",
			mockInit: func(m pgxmock.PgxPoolIface) {
				// Первый запрос: общие цифры
				m.ExpectQuery(`SELECT count\(\*\), (.+) FROM "support_ticket"`).
					WillReturnRows(pgxmock.NewRows([]string{"count", "avg_rating", "avg_time"}).
						AddRow(int64(10), 4.5, int64(3600)))

				// Второй запрос: группировка по статусам
				m.ExpectQuery(`SELECT current_status, count\(\*\) FROM "support_ticket"`).
					WillReturnRows(pgxmock.NewRows([]string{"status", "count"}).
						AddRow("open", int64(3)).
						AddRow("closed", int64(7)))

				// Третий запрос: группировка по категориям
				m.ExpectQuery(`SELECT c.name, count\(t.id\) FROM "support_category"`).
					WillReturnRows(pgxmock.NewRows([]string{"name", "count"}).
						AddRow("Техническая", int64(6)).
						AddRow("Оплата", int64(4)))
			},
			want: domain.SupportStats{
				TotalTickets:         10,
				AverageRating:        4.5,
				AvgResolutionTimeSec: 3600,
				ByStatus:             map[string]int64{"open": 3, "closed": 7},
				ByCategory:           map[string]int64{"Техническая": 6, "Оплата": 4},
			},
		},
		{
			name: "Ошибка на первом запросе",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(`SELECT count\(\*\)`).WillReturnError(errors.New("db crash"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			repo := NewSupportRepo(mock)
			tt.mockInit(mock)

			res, err := repo.GetStats(ctx)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want.TotalTickets, res.TotalTickets)
				assert.Equal(t, tt.want.ByStatus, res.ByStatus)
				assert.Equal(t, tt.want.ByCategory, res.ByCategory)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestSupportRepo_CreateAgentProfile(t *testing.T) {
	ctx := context.Background()
	var agentID int64 = 42

	type mockInit func(m pgxmock.PgxPoolIface)
	tests := []struct {
		name     string
		mockInit mockInit
		wantErr  bool
	}{
		{
			name: "Успешное создание профиля агента",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectExec(`INSERT INTO "support_agent_profile"`).
					WithArgs(agentID).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
			},
		},
		{
			name: "Ошибка БД",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectExec(`INSERT INTO "support_agent_profile"`).
					WithArgs(agentID).
					WillReturnError(errors.New("insert failed"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			repo := NewSupportRepo(mock)
			tt.mockInit(mock)

			err = repo.CreateAgentProfile(ctx, agentID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "insert failed")
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestSupportRepo_DeleteAgentProfile(t *testing.T) {
	ctx := context.Background()
	var agentID int64 = 42

	type mockInit func(m pgxmock.PgxPoolIface)
	tests := []struct {
		name     string
		mockInit mockInit
		wantErr  bool
	}{
		{
			name: "Успешное удаление профиля",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectExec(`DELETE FROM "support_agent_profile" WHERE account_id = \$1`).
					WithArgs(agentID).
					WillReturnResult(pgxmock.NewResult("DELETE", 1))
			},
		},
		{
			name: "Ошибка при удалении",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectExec(`DELETE FROM "support_agent_profile" WHERE account_id = \$1`).
					WithArgs(agentID).
					WillReturnError(errors.New("delete failed"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			repo := NewSupportRepo(mock)
			tt.mockInit(mock)

			err = repo.DeleteAgentProfile(ctx, agentID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "delete failed")
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
