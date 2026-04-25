package usecase

import (
	"context"
	"encoding/json"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/support/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/errutil"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	"google.golang.org/grpc/codes"
)

// Статусы тикетов
const (
	StatusOpen       = "OPEN"
	StatusInProgress = "IN_PROGRESS"
	StatusClosed     = "CLOSED"
)

// Роли авторов
const (
	RoleUser    = "USER"
	RoleSupport = "SUPPORT"
	RoleSystem  = "SYSTEM"
)

// Типы событий
const (
	EventTicketCreated = "TICKET_CREATED"
	EventMessage       = "MESSAGE"
	EventStatusChanged = "STATUS_CHANGED"
	EventReassigned    = "REASSIGNED"
)

//go:generate mockgen -destination=mocks/support_repo_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/services/support/internal/usecase SupportRepository
type SupportRepository interface {
	CreateTicketWithEvent(ctx context.Context, ticket domain.Ticket, event domain.Event) (domain.Ticket, error)
	CreateEvent(ctx context.Context, event domain.Event) (domain.Event, error)
	GetTicketsByUser(ctx context.Context, clientID *int64, guestID *string) ([]domain.Ticket, error)
	GetEventsByTicketID(ctx context.Context, ticketID int64) ([]domain.Event, error)
	GetTicketByID(ctx context.Context, ticketID int64) (domain.Ticket, error)
}

//go:generate mockgen -destination=mocks/support_usecase_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/services/support/internal/usecase SupportUseCase
type SupportUseCase interface {
	CreateTicket(ctx context.Context, req domain.CreateTicketInput) (domain.Ticket, error)
	SendMessage(ctx context.Context, req domain.SendMessageInput) (domain.Event, error)
	GetUserTickets(ctx context.Context, clientID *int64, guestID *string) ([]domain.Ticket, error)
	GetTicketMessages(ctx context.Context, ticketID int64) ([]domain.Event, error)
}

type supportUseCase struct {
	supportRepo SupportRepository
	logger      logger.Logger
}

func NewSupportUseCase(sr SupportRepository, l logger.Logger) SupportUseCase {
	return &supportUseCase{
		supportRepo: sr,
		logger:      l,
	}
}

func (s *supportUseCase) CreateTicket(ctx context.Context, req domain.CreateTicketInput) (domain.Ticket, error) {
	// Валидируем
	if req.ClientID == nil && req.GuestID == nil {
		return domain.Ticket{}, errutil.New("UNAUTHORIZED", "either client_id or guest_id must be provided", codes.Unauthenticated)
	}
	if req.Message == "" {
		return domain.Ticket{}, errutil.New("INVALID_ARGUMENT", "message cannot be empty", codes.InvalidArgument)
	}

	// Собираем тикет
	ticket := domain.Ticket{
		ClientID:      req.ClientID,
		GuestID:       req.GuestID,
		ContactEmail:  req.ContactEmail,
		CategoryID:    req.CategoryID,
		CurrentStatus: StatusOpen,
		SupportLine:   1, // автоматом на первую линию отправляем | TODO: сделать возможность устанавливать на вторую линию, если тикет создается сотрудником на другую проблему для передачи на вторую линию
		ClientMeta:    req.ClientMeta,
		CreatorRole:   RoleUser,
	}

	// Формируем payload для первого события
	payloadMap := map[string]string{"text": req.Message}
	payloadBytes, err := json.Marshal(payloadMap)
	if err != nil {
		return domain.Ticket{}, errutil.Wrap("INTERNAL_ERROR", "failed to marshal event payload", err, codes.Internal)
	}

	event := domain.Event{
		AuthorID:   req.ClientID,
		AuthorRole: RoleUser,
		EventType:  EventTicketCreated,
		Payload:    payloadBytes,
	}

	// Сохраняем в репозиторий в одной транзакции
	createdTicket, err := s.supportRepo.CreateTicketWithEvent(ctx, ticket, event)
	if err != nil {
		return domain.Ticket{}, errutil.Wrap("INTERNAL_SERVER_ERROR", "failed to create ticket and event in db", err, codes.Internal)
	}

	return createdTicket, nil
}

func (s *supportUseCase) SendMessage(ctx context.Context, req domain.SendMessageInput) (domain.Event, error) {
	if req.Message == "" {
		return domain.Event{}, errutil.New("INVALID_ARGUMENT", "message cannot be empty", codes.InvalidArgument)
	}

	// Проверяем, существует ли тикет
	ticket, err := s.supportRepo.GetTicketByID(ctx, req.TicketID)
	if err != nil {
		return domain.Event{}, errutil.Wrap("NOT_FOUND", "ticket not found", err, codes.NotFound)
	}

	if ticket.CurrentStatus == StatusClosed {
		return domain.Event{}, errutil.New("TICKET_CLOSED", "cannot send message to a closed ticket", codes.FailedPrecondition)
	}

	// Формируем payload
	payloadMap := map[string]string{"text": req.Message}
	payloadBytes, err := json.Marshal(payloadMap)
	if err != nil {
		return domain.Event{}, errutil.Wrap("INTERNAL_ERROR", "failed to marshal event payload", err, codes.Internal)
	}

	event := domain.Event{
		TicketID:   req.TicketID,
		AuthorID:   req.AuthorID,
		AuthorRole: req.AuthorRole,
		EventType:  EventMessage,
		Payload:    payloadBytes,
	}

	createdEvent, err := s.supportRepo.CreateEvent(ctx, event)
	if err != nil {
		return domain.Event{}, errutil.Wrap("INTERNAL_SERVER_ERROR", "failed to create message event", err, codes.Internal)
	}

	return createdEvent, nil
}

func (s *supportUseCase) GetUserTickets(ctx context.Context, clientID *int64, guestID *string) ([]domain.Ticket, error) {
	if clientID == nil && guestID == nil {
		return nil, errutil.New("UNAUTHORIZED", "either client_id or guest_id must be provided", codes.Unauthenticated)
	}

	tickets, err := s.supportRepo.GetTicketsByUser(ctx, clientID, guestID)
	if err != nil {
		return nil, errutil.Wrap("INTERNAL_SERVER_ERROR", "failed to get user tickets", err, codes.Internal)
	}

	return tickets, nil
}

func (s *supportUseCase) GetTicketMessages(ctx context.Context, ticketID int64) ([]domain.Event, error) {
	events, err := s.supportRepo.GetEventsByTicketID(ctx, ticketID)
	if err != nil {
		return nil, errutil.Wrap("INTERNAL_SERVER_ERROR", "failed to get ticket events", err, codes.Internal)
	}

	return events, nil
}
