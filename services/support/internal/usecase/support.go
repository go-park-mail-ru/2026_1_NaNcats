package usecase

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/support/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/support/internal/repository"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/errutil"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
)

//go:generate mockgen -destination=mocks/support_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/services/support/internal/usecase SupportUseCase
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
	if input.ClientID == nil && (input.GuestID == nil || *input.GuestID == uuid.Nil) {
		return uuid.Nil, domain.ErrUnauthorized
	}

	// Умная маршрутизация: достаем категорию, чтобы узнать default_line
	categories, err := u.repo.GetActiveCategories(ctx)
	if err != nil {
		return uuid.Nil, errutil.Internal("failed to fetch categories for routing", err)
	}

	// Ищем нашу категорию и выставляем линию
	input.SupportLine = 1
	for _, cat := range categories {
		if cat.ID == input.CategoryID {
			input.SupportLine = cat.DefaultLine
			break
		}
	}

	// Создание
	publicID, err := u.repo.CreateTicket(ctx, input)
	if err != nil {
		return uuid.Nil, errutil.Wrap("INTERNAL_SERVER_ERROR", "failed to create ticket", err, codes.Internal)
	}

	return publicID, nil
}

func (u *supportUseCase) GetStats(ctx context.Context, agentID int64) (domain.SupportStats, error) {
	agent, err := u.repo.GetAgentProfile(ctx, agentID)
	if err != nil {
		return domain.SupportStats{}, domain.ErrPermissionDenied
	}
	if agent.SupportLine < 2 { // статистику видят только L2 и выше
		return domain.SupportStats{}, domain.ErrPermissionDenied
	}

	liveStats, err := u.repo.GetStats(ctx)
	if err != nil {
		return domain.SupportStats{}, errutil.Internal("failed to fetch live stats", err)
	}

	return liveStats, nil
}

func (u *supportUseCase) GetMyTickets(ctx context.Context, clientID *int64, guestID *uuid.UUID) ([]domain.Ticket, error) {
	if clientID != nil {
		return u.repo.GetTicketsByClientID(ctx, *clientID)
	}
	if guestID != nil && *guestID != uuid.Nil {
		return u.repo.GetTicketsByGuestID(ctx, *guestID)
	}
	return nil, domain.ErrUnauthorized
}

func (u *supportUseCase) AddMessage(ctx context.Context, ticketPublicID uuid.UUID, authorID *int64, authorRole string, text, idempotencyKey string) error {
	if text == "" {
		return domain.ErrInvalidMessageInput
	}

	ticket, err := u.repo.GetTicketByPublicID(ctx, ticketPublicID)
	if err != nil {
		return domain.ErrTicketNotFound
	}

	if ticket.CurrentStatus == "closed" && authorRole == "user" {
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
	ticket, err := u.repo.GetTicketByPublicID(ctx, ticketPublicID)
	if err != nil {
		return nil, domain.ErrTicketNotFound
	}

	events, err := u.repo.GetEventsByTicketID(ctx, ticket.ID)
	if err != nil {
		return nil, errutil.Wrap("INTERNAL_SERVER_ERROR", "failed to get events", err, codes.Internal)
	}

	return events, nil
}

func (u *supportUseCase) GetTicketEvents(ctx context.Context, publicID uuid.UUID, clientID *int64, guestID *uuid.UUID) ([]domain.Event, error) {
	ticket, err := u.repo.GetTicketByPublicID(ctx, publicID)
	if err != nil {
		return nil, domain.ErrTicketNotFound
	}

	isOwner := (clientID != nil && ticket.ClientID != nil && *ticket.ClientID == *clientID) ||
		(guestID != nil && ticket.GuestID != nil && *ticket.GuestID == *guestID)

	if !isOwner {
		return nil, domain.ErrPermissionDenied
	}

	events, err := u.repo.GetEventsByTicketID(ctx, ticket.ID)
	if err != nil {
		return nil, errutil.Internal("failed to get ticket events", err)
	}

	return events, nil
}

func (u *supportUseCase) RateTicket(ctx context.Context, publicID uuid.UUID, rating int, clientID *int64, idempotencyKey string) error {
	if rating < 1 || rating > 5 {
		return domain.ErrInvalidRatingInput
	}

	ticket, err := u.repo.GetTicketByPublicID(ctx, publicID)
	if err != nil {
		return domain.ErrTicketNotFound
	}

	if ticket.CurrentStatus != "closed" {
		return domain.ErrInvalidState
	}

	return u.repo.SetTicketRating(ctx, ticket.ID, rating, clientID, idempotencyKey)
}

func (u *supportUseCase) ChangeTicketStatus(ctx context.Context, ticketPublicID uuid.UUID, newStatus string, agentID int64, idempotencyKey string) error {
	ticket, err := u.repo.GetTicketByPublicID(ctx, ticketPublicID)
	if err != nil {
		return domain.ErrTicketNotFound
	}

	if ticket.CurrentStatus == newStatus {
		return nil
	}

	return u.repo.UpdateTicketStatus(
		ctx,
		ticket.ID,
		newStatus,
		&agentID,
		"support",
		idempotencyKey,
	)
}

func (u *supportUseCase) ReassignTicket(ctx context.Context, ticketPublicID uuid.UUID, newAgentID int64, line int, authorID int64, idempotencyKey string) error {
	ticket, err := u.repo.GetTicketByPublicID(ctx, ticketPublicID)
	if err != nil {
		return domain.ErrTicketNotFound
	}

	return u.repo.AssignTicket(
		ctx,
		ticket.ID,
		newAgentID,
		line,
		&authorID, // автор действия (кто переназначил)
		"support",
		idempotencyKey,
	)
}

func (u *supportUseCase) GetAssignedTickets(ctx context.Context, agentID int64) ([]domain.Ticket, error) {
	tickets, err := u.repo.GetAssignedTickets(ctx, agentID)
	if err != nil {
		return nil, errutil.Internal("failed to get assigned tickets", err)
	}
	return tickets, nil
}

func (u *supportUseCase) SetAgentStatus(ctx context.Context, agentID int64, status string) error {
	if status != "online" && status != "offline" {
		return domain.ErrInvalidStatusInput
	}
	return u.repo.UpdateAgentStatus(ctx, agentID, status)

}

func (u *supportUseCase) GetCategories(ctx context.Context) ([]domain.Category, error) {
	return u.repo.GetActiveCategories(ctx)
}

func (u *supportUseCase) GetTemplates(ctx context.Context) ([]domain.Template, error) {
	return u.repo.GetTemplates(ctx)
}
