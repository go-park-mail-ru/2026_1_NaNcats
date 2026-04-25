package domain

import "time"

type Ticket struct {
	ID               int64
	PublicID         string
	ClientID         *int64
	GuestID          *string
	ContactEmail     string
	CategoryID       int
	CurrentStatus    string
	SupportLine      int
	AssigneeID       *int64
	ResolutionRating *int
	ClientMeta       []byte
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
	Payload    []byte
	CreatedAt  time.Time
}

type CreateTicketInput struct {
	ClientID     *int64
	GuestID      *string
	ContactEmail string
	CategoryID   int
	Message      string
	ClientMeta   []byte
}

type SendMessageInput struct {
	TicketID   int64
	AuthorID   *int64
	AuthorRole string
	Message    string
}
