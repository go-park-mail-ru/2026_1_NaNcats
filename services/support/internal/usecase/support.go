package usecase

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/support/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/support/internal/repository"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/errutil"
	"github.com/mailru/easyjson"
	"google.golang.org/grpc/codes"
)

//go:generate mockgen -destination=mocks/support_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/services/support/internal/usecase SupportUseCase
type SupportUseCase interface {
	// Для пользователей / гостей
	CreateTicket(ctx context.Context, input domain.CreateTicketInput) (string, error)
	GetMyTickets(ctx context.Context, clientID *int64, guestID *string) ([]domain.Ticket, error)
	AddMessage(ctx context.Context, ticketPublicID string, authorID *int64, authorRole string, text, idempotencyKey string) error
	GetTicketChat(ctx context.Context, ticketPublicID string) ([]domain.Event, error)
	RateTicket(ctx context.Context, ticketPublicID string, rating int, authorID *int64, idempotencyKey string) error

	// Для операторов
	GetAssignedTickets(ctx context.Context, agentID int64) ([]domain.Ticket, error)
	ChangeTicketStatus(ctx context.Context, ticketPublicID string, status string, agentID int64, idempotencyKey string) error
	ReassignTicket(ctx context.Context, ticketPublicID string, agentID int64, line int, authorID int64, idempotencyKey string) error
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

func (u *supportUseCase) CreateTicket(ctx context.Context, input domain.CreateTicketInput) (string, error) {
	// Проверка авторизации
	if input.ClientID == nil && (input.GuestID == nil || *input.GuestID == "") {
		return "", domain.ErrUnauthorized
	}

	// Умная маршрутизация: достаем категорию, чтобы узнать default_line
	categories, err := u.repo.GetActiveCategories(ctx)
	if err != nil {
		return "", errutil.Internal("failed to fetch categories for routing", err)
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
		return "", errutil.Wrap("INTERNAL_SERVER_ERROR", "failed to create ticket", err, codes.Internal)
	}

	return publicID, nil
}

func (u *supportUseCase) GetMyTickets(ctx context.Context, clientID *int64, guestID *string) ([]domain.Ticket, error) {
	if clientID != nil {
		return u.repo.GetTicketsByClientID(ctx, *clientID)
	}
	if guestID != nil && *guestID != "" {
		return u.repo.GetTicketsByGuestID(ctx, *guestID)
	}
	return nil, domain.ErrUnauthorized
}

func (u *supportUseCase) AddMessage(ctx context.Context, ticketPublicID string, authorID *int64, authorRole string, text, idempotencyKey string) error {
	if text == "" {
		return domain.ErrInvalidMessageInput
	}

	ticket, err := u.repo.GetTicketByPublicID(ctx, ticketPublicID)
	if err != nil {
		return domain.ErrTicketNotFound
	}

	if ticket.CurrentStatus == "closed" && authorRole == "user" {
		err = u.repo.UpdateTicketStatus(ctx, ticket.ID, "open", nil, "system", idempotencyKey+"_reopen")
		if err != nil {
			return errutil.Internal("failed to auto-reopen ticket", err)
		}
	}

	payload := domain.MessagePayload{Text: text}
	payloadBytes, err := easyjson.Marshal(payload)
	if err != nil {
		return errutil.Internal("failed to marshal message", err)
	}

	event := domain.AddEventInput{
		TicketID:       ticket.ID,
		AuthorID:       authorID,
		AuthorRole:     authorRole,
		EventType:      "message",
		Payload:        payloadBytes,
		IdempotencyKey: idempotencyKey,
	}

	return u.repo.AddEvent(ctx, event)
}

func (u *supportUseCase) GetTicketChat(ctx context.Context, ticketPublicID string) ([]domain.Event, error) {
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

func (u *supportUseCase) GetTicketEvents(ctx context.Context, publicID string, clientID *int64, guestID *string) ([]domain.Event, error) {
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

func (u *supportUseCase) RateTicket(ctx context.Context, publicID string, rating int, clientID *int64, idempotencyKey string) error {
	if rating < 1 || rating > 5 {
		return domain.ErrInvalidRatingInput
	}

	ticket, err := u.repo.GetTicketByPublicID(ctx, publicID)
	if err != nil {
		return errutil.Wrap("TICKET_NOT_FOUND", "ticket not found", err, codes.NotFound)
	}

	if ticket.CurrentStatus != "closed" {
		return domain.ErrInvalidState
	}

	err = u.repo.SetTicketRating(ctx, ticket.ID, rating, clientID, idempotencyKey)
	if err != nil {
		return errutil.Internal("failed to set rating", err)
	}

	return nil
}

func (u *supportUseCase) ChangeTicketStatus(ctx context.Context, ticketPublicID string, status string, agentID int64, idempotencyKey string) error {
	ticket, err := u.repo.GetTicketByPublicID(ctx, ticketPublicID)
	if err != nil {
		return errutil.Wrap("TICKET_NOT_FOUND", "ticket not found", err, codes.NotFound)
	}

	// Защита от лишних действий
	if ticket.CurrentStatus == status {
		return nil
	}

	err = u.repo.UpdateTicketStatus(ctx, ticket.ID, status, &agentID, "support", idempotencyKey)
	if err != nil {
		return errutil.Internal("failed to update status", err)
	}
	return nil
}

func (u *supportUseCase) ReassignTicket(ctx context.Context, ticketPublicID string, newAgentID int64, line int, authorID int64, idempotencyKey string) error {
	ticket, err := u.repo.GetTicketByPublicID(ctx, ticketPublicID)
	if err != nil {
		return domain.ErrTicketNotFound
	}

	if err := u.repo.AssignTicket(ctx, ticket.ID, newAgentID, line, &authorID, "support", idempotencyKey); err != nil {
		return errutil.Internal("failed to reassign ticket", err)
	}

	return nil
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
