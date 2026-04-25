package domain

import (
	"encoding/json"
	"time"
)

type Ticket struct {
	ID               int64
	PublicID         string
	ClientID         *int64
	GuestID          *string
	ContactEmail     string
	CategoryID       int64
	CurrentStatus    string
	SupportLine      int
	AssigneeID       *int64
	ResolutionRating *int
	ClientMeta       json.RawMessage
	CreatorRole      string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type Event struct {
	ID         int64
	TicketID   int64
	AuthorID   *int64
	AuthorRole string
	EventType  string
	Payload    json.RawMessage
	CreatedAt  time.Time
}
type AgentProfile struct {
	AccountID   int64
	Status      string
	SupportLine int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
type CreateTicketInput struct {
	ClientID       *int64
	GuestID        *string
	ContactEmail   string
	CategoryID     int64
	FirstMessage   string
	ClientMeta     json.RawMessage
	CreatorRole    string
	IdempotencyKey string
}

type AddEventInput struct {
	TicketID       int64
	AuthorID       *int64
	AuthorRole     string
	EventType      string
	Payload        json.RawMessage
	IdempotencyKey string
}

type Category struct {
	ID          int64  `db:"id"`
	Name        string `db:"name"`
	Description string `db:"description"`
	IsActive    bool   `db:"is_active"`
}

type Template struct {
	ID      int64  `db:"id"`
	Name    string `db:"name"`
	Content string `db:"content"`
}

type MessagePayload struct {
	Text string
}

type StatusChangedPayload struct {
	OldStatus string
	NewStatus string
	Reason    string
}
