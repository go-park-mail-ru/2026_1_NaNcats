package repository

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/support/internal/domain"
)

type SupportRepository interface {
	// Тикеты
	CreateTicket(ctx context.Context, input domain.CreateTicketInput) (string, error)
	GetTicketByPublicID(ctx context.Context, publicID string) (domain.Ticket, error)
	GetTicketsByClientID(ctx context.Context, clientID int64) ([]domain.Ticket, error)
	GetTicketsByGuestID(ctx context.Context, guestID string) ([]domain.Ticket, error)
	GetAssignedTickets(ctx context.Context, agentID int64) ([]domain.Ticket, error)

	// Управление состоянием (с обновлением updated_at)
	UpdateTicketStatus(ctx context.Context, ticketID int64, status string) error
	AssignTicket(ctx context.Context, ticketID int64, agentID int64, line int) error
	SetTicketRating(ctx context.Context, ticketID int64, rating int) error

	// События (чат и логи)
	AddEvent(ctx context.Context, input domain.AddEventInput) error
	GetEventsByTicketID(ctx context.Context, ticketID int64) ([]domain.Event, error)

	// Профили агентов
	UpdateAgentStatus(ctx context.Context, agentID int64, status string) error
	GetAgentProfile(ctx context.Context, agentID int64) (domain.AgentProfile, error)

	// Справочники
	GetActiveCategories(ctx context.Context) ([]domain.Category, error)
	// Быстрые готовые ответы
	GetTemplates(ctx context.Context) ([]domain.Template, error)
}
