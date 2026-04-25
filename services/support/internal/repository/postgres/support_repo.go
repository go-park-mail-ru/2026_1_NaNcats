package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/support/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/support/internal/repository"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/postgres"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type agentProfileDB struct {
	AccountID   int64     `db:"account_id"`
	Status      string    `db:"status"`
	SupportLine int       `db:"support_line"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

func (a agentProfileDB) toDomain() domain.AgentProfile {
	return domain.AgentProfile{
		AccountID:   a.AccountID,
		Status:      a.Status,
		SupportLine: a.SupportLine,
		CreatedAt:   a.CreatedAt,
		UpdatedAt:   a.UpdatedAt,
	}
}

type ticketDB struct {
	ID               int64           `db:"id"`
	PublicID         string          `db:"public_id"`
	ClientAccountID  *int64          `db:"client_account_id"`
	GuestID          *string         `db:"guest_id"`
	ContactEmail     string          `db:"contact_email"`
	CategoryID       int64           `db:"category_id"`
	CurrentStatus    string          `db:"current_status"`
	SupportLine      int             `db:"support_line"`
	AssigneeID       *int64          `db:"assignee_id"`
	ResolutionRating *int            `db:"resolution_rating"`
	ClientMeta       json.RawMessage `db:"client_meta"`
	CreatorRole      string          `db:"creator_role"`
	CreatedAt        time.Time       `db:"created_at"`
	UpdatedAt        time.Time       `db:"updated_at"`
}

func (t ticketDB) toDomain() domain.Ticket {
	return domain.Ticket{
		ID:               t.ID,
		PublicID:         t.PublicID,
		ClientID:         t.ClientAccountID,
		GuestID:          t.GuestID,
		ContactEmail:     t.ContactEmail,
		CategoryID:       int(t.CategoryID),
		CurrentStatus:    t.CurrentStatus,
		SupportLine:      t.SupportLine,
		AssigneeID:       t.AssigneeID,
		ResolutionRating: t.ResolutionRating,
		ClientMeta:       t.ClientMeta,
		CreatorRole:      t.CreatorRole,
		CreatedAt:        t.CreatedAt,
		UpdatedAt:        t.UpdatedAt,
	}
}

type eventDB struct {
	ID              int64           `db:"id"`
	TicketID        int64           `db:"ticket_id"`
	AuthorAccountID *int64          `db:"author_account_id"`
	AuthorRole      string          `db:"author_role"`
	EventType       string          `db:"event_type"`
	Payload         json.RawMessage `db:"payload"`
	CreatedAt       time.Time       `db:"created_at"`
}

func (e eventDB) toDomain() domain.Event {
	return domain.Event{
		ID:         e.ID,
		TicketID:   e.TicketID,
		AuthorID:   e.AuthorAccountID,
		AuthorRole: e.AuthorRole,
		EventType:  e.EventType,
		Payload:    e.Payload,
		CreatedAt:  e.CreatedAt,
	}
}

type supportRepo struct {
	pool postgres.PgxPool
}

func NewSupportRepo(pool postgres.PgxPool) repository.SupportRepository {
	return &supportRepo{pool: pool}
}

// CreateTicket создает тикет и первое сообщение в одной транзакции
func (r *supportRepo) CreateTicket(ctx context.Context, input domain.CreateTicketInput) (string, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)

	query := `
		INSERT INTO "support_ticket" (
			client_account_id, guest_id, contact_email, category_id, 
			creator_role, client_meta, idempotency_key
		) VALUES ($1, NULLIF($2, '')::uuid, $3, $4, $5, $6, $7)
		RETURNING id, public_id;
	`
	var ticketID int64
	var publicID string

	err = tx.QueryRow(ctx, query,
		input.ClientID,
		input.GuestID,
		input.ContactEmail,
		input.CategoryID,
		input.CreatorRole,
		input.ClientMeta,
		input.IdempotencyKey,
	).Scan(&ticketID, &publicID)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return "", fmt.Errorf("ticket already exists: %w", err)
		}
		return "", err
	}

	// Создаем начальное событие (первое сообщение клиента)
	eventPayload, _ := json.Marshal(map[string]string{"text": input.FirstMessage})
	eventQuery := `
		INSERT INTO "support_event" (ticket_id, author_account_id, author_role, event_type, payload)
		VALUES ($1, $2, $3, 'ticket_created', $4)
	`
	_, err = tx.Exec(ctx, eventQuery, ticketID, input.ClientID, input.CreatorRole, eventPayload)
	if err != nil {
		return "", err
	}

	return publicID, tx.Commit(ctx)
}

func (r *supportRepo) AddEvent(ctx context.Context, input domain.AddEventInput) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// 1. Вставляем событие
	query := `
		INSERT INTO "support_event" (ticket_id, author_account_id, author_role, event_type, payload, idempotency_key)
		VALUES ($1, $2, $3, $4, $5, NULLIF($6, ''))
	`
	_, err = tx.Exec(ctx, query,
		input.TicketID,
		input.AuthorID,
		input.AuthorRole,
		input.EventType,
		input.Payload,
		input.IdempotencyKey,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return nil // Идемпотентность: событие уже есть
		}
		return err
	}

	// 2. Обновляем время тикета (замена триггера)
	updateTimeQuery := `UPDATE "support_ticket" SET updated_at = NOW() WHERE id = $1`
	_, err = tx.Exec(ctx, updateTimeQuery, input.TicketID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *supportRepo) UpdateTicketStatus(ctx context.Context, ticketID int64, status string) error {
	query := `
		UPDATE "support_ticket" 
		SET current_status = $1, updated_at = NOW() 
		WHERE id = $2
	`
	tag, err := r.pool.Exec(ctx, query, status, ticketID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("ticket not found")
	}
	return nil
}

func (r *supportRepo) AssignTicket(ctx context.Context, ticketID int64, agentID int64, line int) error {
	query := `
		UPDATE "support_ticket" 
		SET assignee_id = $1, support_line = $2, updated_at = NOW() 
		WHERE id = $3
	`
	_, err := r.pool.Exec(ctx, query, agentID, line, ticketID)
	return err
}

func (r *supportRepo) GetTicketByPublicID(ctx context.Context, publicID string) (domain.Ticket, error) {
	query := `SELECT * FROM "support_ticket" WHERE public_id = $1`
	rows, _ := r.pool.Query(ctx, query, publicID)

	dbT, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[ticketDB])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Ticket{}, errors.New("ticket not found")
		}
		return domain.Ticket{}, err
	}
	return dbT.toDomain(), nil
}

func (r *supportRepo) GetTicketsByClientID(ctx context.Context, clientID int64) ([]domain.Ticket, error) {
	query := `
		SELECT * FROM "support_ticket" 
		WHERE client_account_id = $1 
		ORDER BY updated_at DESC;
	`
	rows, err := r.pool.Query(ctx, query, clientID)
	if err != nil {
		return nil, fmt.Errorf("query client tickets: %w", err)
	}
	defer rows.Close()

	dbTickets, err := pgx.CollectRows(rows, pgx.RowToStructByName[ticketDB])
	if err != nil {
		return nil, fmt.Errorf("collect client tickets: %w", err)
	}

	result := make([]domain.Ticket, 0, len(dbTickets))
	for _, t := range dbTickets {
		result = append(result, t.toDomain())
	}

	return result, nil
}

func (r *supportRepo) GetTicketsByGuestID(ctx context.Context, guestID string) ([]domain.Ticket, error) {
	query := `
		SELECT * FROM "support_ticket" 
		WHERE guest_id = $1 
		ORDER BY updated_at DESC;
	`
	rows, err := r.pool.Query(ctx, query, guestID)
	if err != nil {
		return nil, fmt.Errorf("query guest tickets: %w", err)
	}
	defer rows.Close()

	dbTickets, err := pgx.CollectRows(rows, pgx.RowToStructByName[ticketDB])
	if err != nil {
		return nil, fmt.Errorf("collect guest tickets: %w", err)
	}

	result := make([]domain.Ticket, 0, len(dbTickets))
	for _, t := range dbTickets {
		result = append(result, t.toDomain())
	}

	return result, nil
}

func (r *supportRepo) GetAssignedTickets(ctx context.Context, agentID int64) ([]domain.Ticket, error) {
	query := `
		SELECT * FROM "support_ticket" 
		WHERE assignee_id = $1 AND current_status != 'closed'
		ORDER BY updated_at DESC;
	`
	rows, err := r.pool.Query(ctx, query, agentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	dbTickets, err := pgx.CollectRows(rows, pgx.RowToStructByName[ticketDB])
	if err != nil {
		return nil, err
	}

	result := make([]domain.Ticket, 0, len(dbTickets))
	for _, t := range dbTickets {
		result = append(result, t.toDomain())
	}

	return result, nil
}

func (r *supportRepo) GetEventsByTicketID(ctx context.Context, ticketID int64) ([]domain.Event, error) {
	query := `SELECT * FROM "support_event" WHERE ticket_id = $1 ORDER BY created_at ASC`
	rows, _ := r.pool.Query(ctx, query, ticketID)

	dbEvents, err := pgx.CollectRows(rows, pgx.RowToStructByName[eventDB])
	if err != nil {
		return nil, err
	}

	res := make([]domain.Event, 0, len(dbEvents))
	for _, e := range dbEvents {
		res = append(res, e.toDomain())
	}
	return res, nil
}

func (r *supportRepo) UpdateAgentStatus(ctx context.Context, agentID int64, status string) error {
	query := `
		INSERT INTO "support_agent_profile" (account_id, status, updated_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (account_id) DO UPDATE SET status = $2, updated_at = NOW()
	`
	_, err := r.pool.Exec(ctx, query, agentID, status)
	return err
}

func (r *supportRepo) GetActiveCategories(ctx context.Context) ([]domain.Category, error) {
	query := `SELECT id, name, description FROM "support_category" WHERE is_active = true`
	rows, _ := r.pool.Query(ctx, query)

	categories, err := pgx.CollectRows(rows, pgx.RowToStructByPos[domain.Category])
	return categories, err
}

func (r *supportRepo) GetAgentProfile(ctx context.Context, agentID int64) (domain.AgentProfile, error) {
	query := `
		SELECT account_id, status, support_line, created_at, updated_at 
		FROM "support_agent_profile" 
		WHERE account_id = $1
	`

	rows, err := r.pool.Query(ctx, query, agentID)
	if err != nil {
		return domain.AgentProfile{}, fmt.Errorf("query agent profile: %w", err)
	}

	dbAgent, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[agentProfileDB])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.AgentProfile{}, errors.New("agent profile not found")
		}
		return domain.AgentProfile{}, fmt.Errorf("collect agent profile row: %w", err)
	}

	return dbAgent.toDomain(), nil
}

func (r *supportRepo) SetTicketRating(ctx context.Context, ticketID int64, rating int) error {
	query := `
		UPDATE "support_ticket" 
		SET resolution_rating = $1, updated_at = NOW() 
		WHERE id = $2
	`
	tag, err := r.pool.Exec(ctx, query, rating, ticketID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("ticket not found")
	}
	return nil
}
