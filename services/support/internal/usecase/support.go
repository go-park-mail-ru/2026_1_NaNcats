package usecase

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/support/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/support/internal/repository"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/errutil"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

//go:generate mockgen -destination=mocks/support_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/services/support/internal/usecase SupportUseCase
//go:generate gowrap gen -i SupportUseCase -t ../../../../shared/templates/tracing.tmpl -o support_tracing_mw.go -v TracerName=support-service
type SupportUseCase interface {
	// Для пользователей / гостей
	CreateTicket(ctx context.Context, input domain.CreateTicketInput) (uuid.UUID, error)
	GetMyTickets(ctx context.Context, clientID *int64, guestID *uuid.UUID) ([]domain.Ticket, error)
	AddMessage(ctx context.Context, ticketPublicID uuid.UUID, authorID *int64, authorRole string, text, idempotencyKey string) error
	GetTicketChat(ctx context.Context, ticketPublicID uuid.UUID) ([]domain.Event, error)
	RateTicket(ctx context.Context, ticketPublicID uuid.UUID, rating int, authorID *int64, idempotencyKey string) error
	GetTicketEvents(ctx context.Context, publicID uuid.UUID, clientID *int64, guestID *uuid.UUID) ([]domain.Event, error)

	// Для операторов
	GetAssignedTickets(ctx context.Context, agentID int64) ([]domain.Ticket, error)
	ChangeTicketStatus(ctx context.Context, ticketPublicID uuid.UUID, status string, agentID int64, idempotencyKey string) error
	ReassignTicket(ctx context.Context, ticketPublicID uuid.UUID, agentID int64, line int, authorID int64, idempotencyKey string) error
	SetAgentStatus(ctx context.Context, agentID int64, status string) error

	// Справочники
	GetCategories(ctx context.Context) ([]domain.Category, error)
	GetTemplates(ctx context.Context) ([]domain.Template, error)
}

type supportUseCase struct {
	repo repository.SupportRepository
}

func NewSupportUseCase(r repository.SupportRepository) SupportUseCase {
	return &supportUseCase{repo: r}
}

func (u *supportUseCase) CreateTicket(ctx context.Context, input domain.CreateTicketInput) (uuid.UUID, error) {
	// Проверка авторизации
	span := trace.SpanFromContext(ctx)
	if input.ClientID != nil {
		span.SetAttributes(attribute.Int64("user.id", *input.ClientID))
	}
	if input.GuestID != nil {
		span.SetAttributes(attribute.String("guest.id", input.GuestID.String()))
	}
	span.SetAttributes(attribute.Int64("category.id", input.CategoryID))

	if input.ClientID == nil && (input.GuestID == nil || *input.GuestID == uuid.Nil) {
		return uuid.Nil, domain.ErrUnauthorized
	}

	// Достаем категорию, чтобы узнать default_line
	categories, err := u.repo.GetActiveCategories(ctx)
	if err != nil {
		return uuid.Nil, errutil.Internal("failed to fetch categories for routing", err)
	}

	// Ищем категорию и выставляем линию
	input.SupportLine = 1
	for _, cat := range categories {
		if cat.ID == input.CategoryID {
			input.SupportLine = cat.DefaultLine
			break
		}
	}
	span.SetAttributes(attribute.Int("support_line.assigned", input.SupportLine))

	publicID, err := u.repo.CreateTicket(ctx, input)
	if err != nil {
		return uuid.Nil, errutil.Internal("failed to create ticket in db", err)
	}

	span.SetAttributes(attribute.String("ticket.public_id", publicID.String()))
	return publicID, nil
}

func (u *supportUseCase) GetStats(ctx context.Context, agentID int64) (domain.SupportStats, error) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.Int64("agent.id", agentID))

	agent, err := u.repo.GetAgentProfile(ctx, agentID)
	if err != nil {
		return domain.SupportStats{}, domain.ErrPermissionDenied
	}
	span.SetAttributes(attribute.Int("agent.line", agent.SupportLine))

	if agent.SupportLine < 2 {
		span.AddEvent("insufficient_agent_line_for_stats")
		return domain.SupportStats{}, domain.ErrPermissionDenied
	}

	liveStats, err := u.repo.GetStats(ctx)
	if err != nil {
		return domain.SupportStats{}, errutil.Internal("failed to fetch live stats", err)
	}

	return liveStats, nil
}

func (u *supportUseCase) GetMyTickets(ctx context.Context, clientID *int64, guestID *uuid.UUID) ([]domain.Ticket, error) {
	span := trace.SpanFromContext(ctx)

	if clientID != nil {
		span.SetAttributes(attribute.Int64("user.id", *clientID))
		return u.repo.GetTicketsByClientID(ctx, *clientID)
	}
	if guestID != nil && *guestID != uuid.Nil {
		span.SetAttributes(attribute.String("guest.id", guestID.String()))
		return u.repo.GetTicketsByGuestID(ctx, *guestID)
	}

	return nil, domain.ErrUnauthorized
}

func (u *supportUseCase) AddMessage(ctx context.Context, ticketPublicID uuid.UUID, authorID *int64, authorRole string, text, idempotencyKey string) error {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.String("ticket.public_id", ticketPublicID.String()),
		attribute.String("author.role", authorRole),
	)
	if authorID != nil {
		span.SetAttributes(attribute.Int64("author.id", *authorID))
	}

	if text == "" {
		return domain.ErrInvalidMessageInput
	}

	ticket, err := u.repo.GetTicketByPublicID(ctx, ticketPublicID)
	if err != nil {
		return domain.ErrTicketNotFound
	}
	span.SetAttributes(attribute.Int64("ticket.internal_id", ticket.ID))

	if ticket.CurrentStatus == "closed" && authorRole == "user" {
		span.AddEvent("auto_reopening_closed_ticket")
		err = u.repo.UpdateTicketStatus(
			ctx,
			ticket.ID,
			"open",
			nil,
			"system",
			idempotencyKey+"_reopen",
		)
		if err != nil {
			return errutil.Internal("failed to auto-reopen ticket", err)
		}
	}

	return u.repo.AddMessageEvent(ctx, ticket.ID, authorID, authorRole, text, idempotencyKey)
}

func (u *supportUseCase) GetTicketChat(ctx context.Context, ticketPublicID uuid.UUID) ([]domain.Event, error) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.String("ticket.public_id", ticketPublicID.String()))

	ticket, err := u.repo.GetTicketByPublicID(ctx, ticketPublicID)
	if err != nil {
		return nil, domain.ErrTicketNotFound
	}

	events, err := u.repo.GetEventsByTicketID(ctx, ticket.ID)
	if err != nil {
		return nil, errutil.Internal("failed to fetch ticket events from db", err)
	}

	span.SetAttributes(attribute.Int("chat.events_count", len(events)))
	return events, nil
}

func (u *supportUseCase) GetTicketEvents(ctx context.Context, publicID uuid.UUID, clientID *int64, guestID *uuid.UUID) ([]domain.Event, error) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.String("ticket.public_id", publicID.String()))

	if clientID != nil {
		span.SetAttributes(attribute.Int64("client.id", *clientID))
	}
	if guestID != nil {
		span.SetAttributes(attribute.String("guest.id", guestID.String()))
	}

	ticket, err := u.repo.GetTicketByPublicID(ctx, publicID)
	if err != nil {
		return nil, domain.ErrTicketNotFound
	}

	isOwner := (clientID != nil && ticket.ClientID != nil && *ticket.ClientID == *clientID) ||
		(guestID != nil && ticket.GuestID != nil && *ticket.GuestID == *guestID)

	span.SetAttributes(attribute.Bool("auth.is_owner", isOwner))

	if !isOwner {
		return nil, domain.ErrPermissionDenied
	}

	events, err := u.repo.GetEventsByTicketID(ctx, ticket.ID)
	if err != nil {
		return nil, errutil.Internal("failed to get ticket events", err)
	}

	span.SetAttributes(attribute.Int("events.count", len(events)))
	return events, nil
}

func (u *supportUseCase) RateTicket(ctx context.Context, publicID uuid.UUID, rating int, clientID *int64, idempotencyKey string) error {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.String("ticket.public_id", publicID.String()),
		attribute.Int("ticket.rating_value", rating),
	)

	if rating < 1 || rating > 5 {
		return domain.ErrInvalidRatingInput
	}

	ticket, err := u.repo.GetTicketByPublicID(ctx, publicID)
	if err != nil {
		return domain.ErrTicketNotFound
	}

	if ticket.CurrentStatus != "closed" {
		span.AddEvent("rating_rejected_ticket_not_closed")
		return domain.ErrInvalidState
	}

	return u.repo.SetTicketRating(ctx, ticket.ID, rating, clientID, idempotencyKey)
}

func (u *supportUseCase) ChangeTicketStatus(ctx context.Context, ticketPublicID uuid.UUID, newStatus string, agentID int64, idempotencyKey string) error {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.String("ticket.public_id", ticketPublicID.String()),
		attribute.String("ticket.status.new", newStatus),
		attribute.Int64("agent.id", agentID),
	)

	ticket, err := u.repo.GetTicketByPublicID(ctx, ticketPublicID)
	if err != nil {
		return domain.ErrTicketNotFound
	}

	if ticket.CurrentStatus == newStatus {
		span.AddEvent("status_already_set")
		return nil
	}

	return u.repo.UpdateTicketStatus(ctx, ticket.ID, newStatus, &agentID, "support", idempotencyKey)
}

func (u *supportUseCase) ReassignTicket(ctx context.Context, ticketPublicID uuid.UUID, newAgentID int64, line int, authorID int64, idempotencyKey string) error {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.String("ticket.public_id", ticketPublicID.String()),
		attribute.Int64("agent.new_id", newAgentID),
		attribute.Int("support.line_new", line),
		attribute.Int64("author.id", authorID),
	)

	ticket, err := u.repo.GetTicketByPublicID(ctx, ticketPublicID)
	if err != nil {
		return domain.ErrTicketNotFound
	}

	return u.repo.AssignTicket(ctx, ticket.ID, newAgentID, line, &authorID, "support", idempotencyKey)
}

func (u *supportUseCase) GetAssignedTickets(ctx context.Context, agentID int64) ([]domain.Ticket, error) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.Int64("agent.id", agentID))

	tickets, err := u.repo.GetAssignedTickets(ctx, agentID)
	if err != nil {
		return nil, errutil.Internal("failed to get assigned tickets", err)
	}

	span.SetAttributes(attribute.Int("tickets.count", len(tickets)))
	return tickets, nil
}

func (u *supportUseCase) SetAgentStatus(ctx context.Context, agentID int64, status string) error {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.Int64("agent.id", agentID),
		attribute.String("agent.status.new", status),
	)

	if status != "online" && status != "offline" {
		return domain.ErrInvalidStatusInput
	}
	return u.repo.UpdateAgentStatus(ctx, agentID, status)
}

func (u *supportUseCase) GetCategories(ctx context.Context) ([]domain.Category, error) {
	categories, err := u.repo.GetActiveCategories(ctx)
	if err == nil {
		trace.SpanFromContext(ctx).SetAttributes(attribute.Int("categories.count", len(categories)))
	}
	return categories, err
}

func (u *supportUseCase) GetTemplates(ctx context.Context) ([]domain.Template, error) {
	templates, err := u.repo.GetTemplates(ctx)
	if err == nil {
		trace.SpanFromContext(ctx).SetAttributes(attribute.Int("templates.count", len(templates)))
	}
	return templates, err
}
