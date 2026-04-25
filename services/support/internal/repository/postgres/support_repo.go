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
	"github.com/mailru/easyjson"
)

//go:generate easyjson $GOFILE

//easyjson:json
type messagePayloadDTO struct {
	Text string `json:"text"`
}

//easyjson:json
type statusChangedPayloadDTO struct {
	OldStatus string `json:"old_status"`
	NewStatus string `json:"new_status"`
	Reason    string `json:"reason,omitempty"`
}

//easyjson:json
type reassignedPayloadDTO struct {
	OldAssigneeID *int64 `json:"old_assignee_id,omitempty"`
	NewAssigneeID *int64 `json:"new_assignee_id,omitempty"`
	OldLine       int    `json:"old_line,omitempty"`
	NewLine       int    `json:"new_line,omitempty"`
}

// Смена рейтинга
//
//easyjson:json
type ratedPayloadDTO struct {
	Rating int `json:"rating"`
}

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
		CategoryID:       t.CategoryID,
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

func mapDBTicketsToDomain(dbTickets []ticketDB) []domain.Ticket {
	result := make([]domain.Ticket, 0, len(dbTickets))
	for _, t := range dbTickets {
		result = append(result, t.toDomain())
	}
	return result
}

type supportRepo struct {
	pool postgres.PgxPool
}

func NewSupportRepo(pool postgres.PgxPool) repository.SupportRepository {
	return &supportRepo{pool: pool}
}

func (r *supportRepo) CreateTicket(ctx context.Context, input domain.CreateTicketInput) (string, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	ticketQuery := `
		INSERT INTO "support_ticket" (
			client_account_id, guest_id, contact_email, category_id, 
			creator_role, client_meta, idempotency_key,
			created_at, updated_at
		) VALUES ($1, NULLIF($2, '')::uuid, $3, $4, $5, $6, $7, NOW(), NOW())
		RETURNING id, public_id;
	`

	var ticketID int64
	var publicID string

	err = tx.QueryRow(ctx, ticketQuery,
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
			return "", fmt.Errorf("ticket with this idempotency key already exists: %w", err)
		}
		return "", fmt.Errorf("insert support ticket: %w", err)
	}

	payloadDTO := messagePayloadDTO{
		Text: input.FirstMessage,
	}

	eventPayload, err := easyjson.Marshal(payloadDTO)
	if err != nil {
		return "", fmt.Errorf("marshal initial message payload: %w", err)
	}

	eventQuery := `
		INSERT INTO "support_event" (
			ticket_id, author_account_id, author_role, 
			event_type, payload, created_at
		) VALUES ($1, $2, $3, 'ticket_created', $4, NOW())
	`
	_, err = tx.Exec(ctx, eventQuery,
		ticketID,
		input.ClientID,
		input.CreatorRole,
		eventPayload,
	)
	if err != nil {
		return "", fmt.Errorf("insert initial support event: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return "", fmt.Errorf("commit transaction: %w", err)
	}

	return publicID, nil
}

func (r *supportRepo) AddEvent(ctx context.Context, input domain.AddEventInput) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

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
			return nil
		}
		return err
	}

	updateTimeQuery := `UPDATE "support_ticket" SET updated_at = NOW() WHERE id = $1`
	_, err = tx.Exec(ctx, updateTimeQuery, input.TicketID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *supportRepo) UpdateTicketStatus(ctx context.Context, ticketID int64, status string, authorID *int64, authorRole string, idempotencyKey string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var oldStatus string
	err = tx.QueryRow(ctx, `SELECT current_status FROM "support_ticket" WHERE id = $1 FOR UPDATE`, ticketID).Scan(&oldStatus) // Добавил FOR UPDATE для надежности
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New("ticket not found")
		}
		return err
	}

	query := `UPDATE "support_ticket" SET current_status = $1, updated_at = NOW() WHERE id = $2`
	result, err := tx.Exec(ctx, query, status, ticketID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return errors.New("failed to update ticket status")
	}

	// Пишем в историю
	payload, err := easyjson.Marshal(statusChangedPayloadDTO{
		OldStatus: oldStatus,
		NewStatus: status,
	})
	if err != nil {
		return fmt.Errorf("marshal status change: %w", err)
	}

	eventQuery := `
		INSERT INTO "support_event" (ticket_id, author_account_id, author_role, event_type, payload, idempotency_key, created_at)
		VALUES ($1, $2, $3, 'status_changed', $4, NULLIF($5, ''), NOW())
		ON CONFLICT (idempotency_key) DO NOTHING;
	`
	_, err = tx.Exec(ctx, eventQuery, ticketID, authorID, authorRole, payload, idempotencyKey)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *supportRepo) AssignTicket(ctx context.Context, ticketID int64, agentID int64, line int, authorID *int64, authorRole string, idempotencyKey string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var oldAssignee *int64
	var oldLine int
	err = tx.QueryRow(ctx, `SELECT assignee_id, support_line FROM "support_ticket" WHERE id = $1`, ticketID).Scan(&oldAssignee, &oldLine)
	if err != nil {
		return err
	}

	query := `UPDATE "support_ticket" SET assignee_id = $1, support_line = $2, updated_at = NOW() WHERE id = $3`
	if _, err := tx.Exec(ctx, query, agentID, line, ticketID); err != nil {
		return err
	}

	payload, err := easyjson.Marshal(reassignedPayloadDTO{
		OldAssigneeID: oldAssignee,
		NewAssigneeID: &agentID,
		OldLine:       oldLine,
		NewLine:       line,
	})
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}

	eventQuery := `
		INSERT INTO "support_event" (ticket_id, author_account_id, author_role, event_type, payload, idempotency_key, created_at)
		VALUES ($1, $2, $3, 'reassigned', $4, NULLIF($5, ''), NOW())
		ON CONFLICT (idempotency_key) DO NOTHING;
	`
	_, err = tx.Exec(ctx, eventQuery, ticketID, authorID, authorRole, payload, idempotencyKey)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *supportRepo) GetTicketByPublicID(ctx context.Context, publicID string) (domain.Ticket, error) {
	query := `SELECT * FROM "support_ticket" WHERE public_id = $1`
	rows, err := r.pool.Query(ctx, query, publicID)
	if err != nil {
		return domain.Ticket{}, err
	}

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

	return mapDBTicketsToDomain(dbTickets), nil
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

	return mapDBTicketsToDomain(dbTickets), nil
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

	return mapDBTicketsToDomain(dbTickets), nil
}

func (r *supportRepo) GetEventsByTicketID(ctx context.Context, ticketID int64) ([]domain.Event, error) {
	query := `SELECT * FROM "support_event" WHERE ticket_id = $1 ORDER BY created_at ASC`
	rows, err := r.pool.Query(ctx, query, ticketID)
	if err != nil {
		return nil, err
	}

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
	query := `SELECT id, name, description, default_line, is_active FROM "support_category" WHERE is_active = true`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return pgx.CollectRows(rows, pgx.RowToStructByName[domain.Category])
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

func (r *supportRepo) SetTicketRating(ctx context.Context, ticketID int64, rating int, authorID *int64, idempotencyKey string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	query := `UPDATE "support_ticket" SET resolution_rating = $1, updated_at = NOW() WHERE id = $2`
	tag, err := tx.Exec(ctx, query, rating, ticketID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("ticket not found")
	}

	// Добавляем событие в чат
	payload, _ := easyjson.Marshal(ratedPayloadDTO{Rating: rating})
	eventQuery := `
        INSERT INTO "support_event" (ticket_id, author_account_id, author_role, event_type, payload, idempotency_key, created_at)
        VALUES ($1, $2, 'user', 'rated', $3, NULLIF($4, ''), NOW())
        ON CONFLICT (idempotency_key) DO NOTHING;
    `
	_, err = tx.Exec(ctx, eventQuery, ticketID, authorID, payload, idempotencyKey)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *supportRepo) GetTemplates(ctx context.Context) ([]domain.Template, error) {
	query := `
		SELECT id, name, content 
		FROM "support_template" 
		ORDER BY name ASC;
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query support templates: %w", err)
	}
	defer rows.Close()

	templates, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.Template])
	if err != nil {
		return nil, fmt.Errorf("collect support templates: %w", err)
	}

	return templates, nil
}
