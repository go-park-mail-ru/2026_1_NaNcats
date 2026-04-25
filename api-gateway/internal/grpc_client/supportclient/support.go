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
	ErrUnauthorized = errors.New("unauthorized action")
	ErrInternal     = errors.New("internal server error")
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
	GetTickets(ctx context.Context, clientID *int64, guestID *string) ([]Ticket, error)
	SendMessage(ctx context.Context, input SendMessageInput, idempotencyKey string) (int64, error)
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
