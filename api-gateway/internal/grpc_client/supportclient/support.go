package supportclient

import (
	"context"
	"encoding/json"
	"time"
)

type CreateTicketInput struct {
	ClientID     *int64
	GuestID      *string
	ContactEmail string
	CategoryID   int
	FirstMessage string
	ClientMeta   json.RawMessage
}

type Ticket struct {
	ID            int64
	PublicID      string
	CategoryID    int64
	CurrentStatus string
	SupportLine   int
	CreatedAt     time.Time
}

type SupportClient interface {
	CreateTicket(ctx context.Context, input CreateTicketInput, idempotencyKey string) (string, error)
	GetTickets(ctx context.Context, clientID *int64, guestID *string) ([]Ticket, error)
}
