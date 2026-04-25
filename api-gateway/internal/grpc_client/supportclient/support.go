package supportclient

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	pbSupport "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/support"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrUnauthorized   = errors.New("unauthorized action")
	ErrInternal       = errors.New("internal server error")
	ErrTicketNotFound = errors.New("ticket not found")
)

type CreateTicketInput struct {
	ClientID     *int64
	GuestID      *string
	ContactEmail string
	CategoryID   int64
	FirstMessage string
	ClientMeta   json.RawMessage
}

type SendMessageInput struct {
	TicketID   int64
	AuthorID   *int64
	AuthorRole string
	Message    string
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
	SendMessage(ctx context.Context, input SendMessageInput, idempotencyKey string) (int64, error)
	GetUserTickets(ctx context.Context, clientID *int64, guestID *string) ([]Ticket, error)
}

type supportClient struct {
	client pbSupport.SupportServiceClient
}

func NewSupportClient(cl pbSupport.SupportServiceClient) SupportClient {
	return &supportClient{
		client: cl,
	}
}

func (c *supportClient) CreateTicket(ctx context.Context, input CreateTicketInput, idempotencyKey string) (string, error) {
	req := &pbSupport.CreateTicketRequest{
		ContactEmail:   input.ContactEmail,
		CategoryId:     input.CategoryID,
		InitialMessage: input.FirstMessage,
		ClientMeta:     string(input.ClientMeta),
		IdempotencyKey: idempotencyKey,
	}

	if input.ClientID != nil {
		req.ClientId = *input.ClientID
	}
	if input.GuestID != nil {
		req.GuestId = *input.GuestID
	}

	resp, err := c.client.CreateTicket(ctx, req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.Unauthenticated {
			return "", ErrUnauthorized
		}
		return "", ErrInternal
	}

	return resp.PublicId, nil
}

func (c *supportClient) SendMessage(ctx context.Context, input SendMessageInput, idempotencyKey string) (int64, error) {
	req := &pbSupport.SendMessageRequest{
		TicketId:       input.TicketID,
		AuthorRole:     input.AuthorRole,
		Message:        input.Message,
		IdempotencyKey: idempotencyKey,
	}

	if input.AuthorID != nil {
		req.AuthorId = *input.AuthorID
	}

	resp, err := c.client.SendMessage(ctx, req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			return 0, ErrTicketNotFound
		}
		return 0, ErrInternal
	}

	return resp.Id, nil
}

func (c *supportClient) GetUserTickets(ctx context.Context, clientID *int64, guestID *string) ([]Ticket, error) {
	req := &pbSupport.GetUserTicketsRequest{}

	if clientID != nil {
		req.ClientId = *clientID
	}
	if guestID != nil {
		req.GuestId = *guestID
	}

	resp, err := c.client.GetUserTickets(ctx, req)
	if err != nil {
		return nil, ErrInternal
	}

	tickets := make([]Ticket, 0, len(resp.Tickets))
	for _, pbT := range resp.Tickets {
		tickets = append(tickets, Ticket{
			ID:            pbT.Id,
			PublicID:      pbT.PublicId,
			CurrentStatus: pbT.CurrentStatus,
			CreatedAt:     pbT.CreatedAt.AsTime(),
		})
	}

	return tickets, nil
}
