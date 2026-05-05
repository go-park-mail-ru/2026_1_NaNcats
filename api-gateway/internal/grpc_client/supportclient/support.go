package supportclient

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	pbSupport "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/support"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

var (
	ErrUnauthorized   = errors.New("unauthorized action")
	ErrInternal       = errors.New("internal server error")
	ErrTicketNotFound = errors.New("ticket not found")
)

// DTO

type CreateTicketInput struct {
	ClientID     *int64
	GuestID      *string
	ContactEmail string
	CategoryID   int64
	FirstMessage string
	ClientMeta   json.RawMessage
}

type SendMessageInput struct {
	TicketPublicID string
	AuthorID       *int64
	AuthorRole     string
	Message        string
}

type Ticket struct {
	ID               int64
	PublicID         string
	CategoryID       int64
	CurrentStatus    string
	SupportLine      int
	AssigneeID       int64
	ResolutionRating int
	CreatedAt        time.Time
}

type SupportStats struct {
	TotalTickets         int64
	ByStatus             map[string]int64
	ByCategory           map[string]int64
	AverageRating        float64
	AvgResolutionTimeSec int64
}

type Event struct {
	ID         int64
	TicketID   int64
	AuthorID   int64
	AuthorRole string
	EventType  string
	Payload    json.RawMessage
	CreatedAt  time.Time
}

type Category struct {
	ID          int64
	Name        string
	Description string
	DefaultLine int
}

type Template struct {
	ID      int64
	Name    string
	Content string
}

// Интерфейсы

type SupportClient interface {
	// Пользовательская часть
	CreateTicket(ctx context.Context, input CreateTicketInput, idempotencyKey string) (string, error)
	SendMessage(ctx context.Context, input SendMessageInput, idempotencyKey string) (int64, error)
	GetUserTickets(ctx context.Context, clientID *int64, guestID *string) ([]Ticket, error)
	GetTicketEvents(ctx context.Context, ticketPublicID string, clientID *int64, guestID *string) ([]Event, error)
	RateTicket(ctx context.Context, ticketPublicID string, rating int, clientID *int64, idempotencyKey string) error

	// Операторская часть
	GetAssignedTickets(ctx context.Context, agentID int64) ([]Ticket, error)
	ChangeTicketStatus(ctx context.Context, ticketPublicID string, status string, agentID int64, idempotencyKey string) error
	ReassignTicket(ctx context.Context, ticketPublicID string, agentID int64, line int, authorID int64, idempotencyKey string) error
	SetAgentStatus(ctx context.Context, agentID int64, status string) error

	// Справочники
	GetCategories(ctx context.Context) ([]Category, error)
	GetTemplates(ctx context.Context) ([]Template, error)
	GetStats(ctx context.Context) (SupportStats, error)
}

type supportClient struct {
	client pbSupport.SupportServiceClient
}

func NewSupportClient(cl pbSupport.SupportServiceClient) SupportClient {
	return &supportClient{
		client: cl,
	}
}

// Пользовательская часть

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
		TicketPublicId: input.TicketPublicID,
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
			ID:               pbT.Id,
			PublicID:         pbT.PublicId,
			CategoryID:       pbT.CategoryId,
			CurrentStatus:    pbT.CurrentStatus,
			SupportLine:      int(pbT.SupportLine),
			AssigneeID:       pbT.AssigneeId,
			ResolutionRating: int(pbT.ResolutionRating),
			CreatedAt:        pbT.CreatedAt.AsTime(),
		})
	}

	return tickets, nil
}

func (c *supportClient) GetTicketEvents(ctx context.Context, ticketPublicID string, clientID *int64, guestID *string) ([]Event, error) {
	req := &pbSupport.GetTicketEventsRequest{
		TicketPublicId: ticketPublicID,
	}

	if clientID != nil {
		req.ClientId = *clientID
	}
	if guestID != nil {
		req.GuestId = *guestID
	}

	resp, err := c.client.GetTicketEvents(ctx, req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.NotFound:
				return nil, ErrTicketNotFound
			case codes.PermissionDenied:
				return nil, ErrUnauthorized
			}
		}
		return nil, ErrInternal
	}

	events := make([]Event, 0, len(resp.Events))
	for _, event := range resp.Events {
		events = append(events, Event{
			ID:         event.Id,
			TicketID:   event.TicketId,
			AuthorID:   event.AuthorId,
			AuthorRole: event.AuthorRole,
			EventType:  event.EventType,
			Payload:    json.RawMessage([]byte(event.Payload)),
			CreatedAt:  event.CreatedAt.AsTime(),
		})
	}

	return events, nil
}

func (c *supportClient) RateTicket(ctx context.Context, ticketPublicID string, rating int, clientID *int64, idempotencyKey string) error {
	req := &pbSupport.RateTicketRequest{
		TicketPublicId: ticketPublicID,
		Rating:         int32(rating),
		IdempotencyKey: idempotencyKey,
	}

	if clientID != nil {
		req.ClientId = *clientID
	}

	_, err := c.client.RateTicket(ctx, req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			return ErrTicketNotFound
		}
		return ErrInternal
	}

	return nil
}

func (c *supportClient) GetAssignedTickets(ctx context.Context, agentID int64) ([]Ticket, error) {
	req := &pbSupport.GetAssignedTicketsRequest{
		AgentId: agentID,
	}

	resp, err := c.client.GetAssignedTickets(ctx, req)
	if err != nil {
		return nil, ErrInternal
	}

	tickets := make([]Ticket, 0, len(resp.Tickets))
	for _, pbT := range resp.Tickets {
		tickets = append(tickets, Ticket{
			ID:               pbT.Id,
			PublicID:         pbT.PublicId,
			CategoryID:       pbT.CategoryId,
			CurrentStatus:    pbT.CurrentStatus,
			SupportLine:      int(pbT.SupportLine),
			AssigneeID:       pbT.AssigneeId,
			ResolutionRating: int(pbT.ResolutionRating),
			CreatedAt:        pbT.CreatedAt.AsTime(),
		})
	}

	return tickets, nil
}

func (c *supportClient) ChangeTicketStatus(ctx context.Context, ticketPublicID string, newStatus string, agentID int64, idempotencyKey string) error {
	req := &pbSupport.ChangeTicketStatusRequest{
		TicketPublicId: ticketPublicID,
		Status:         newStatus,
		AgentId:        agentID,
		IdempotencyKey: idempotencyKey,
	}

	_, err := c.client.ChangeTicketStatus(ctx, req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			return ErrTicketNotFound
		}
		return ErrInternal
	}

	return nil
}

func (c *supportClient) ReassignTicket(ctx context.Context, ticketPublicID string, agentID int64, line int, authorID int64, idempotencyKey string) error {
	req := &pbSupport.ReassignTicketRequest{
		TicketPublicId: ticketPublicID,
		AgentId:        agentID,
		Line:           int32(line),
		AuthorId:       authorID,
		IdempotencyKey: idempotencyKey,
	}

	_, err := c.client.ReassignTicket(ctx, req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			return ErrTicketNotFound
		}
		return ErrInternal
	}

	return nil
}

func (c *supportClient) SetAgentStatus(ctx context.Context, agentID int64, agentStatus string) error {
	req := &pbSupport.SetAgentStatusRequest{
		AgentId: agentID,
		Status:  agentStatus,
	}

	_, err := c.client.SetAgentStatus(ctx, req)
	if err != nil {
		return ErrInternal
	}

	return nil
}

// === Справочники ===

func (c *supportClient) GetCategories(ctx context.Context) ([]Category, error) {
	resp, err := c.client.GetCategories(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, ErrInternal
	}

	categories := make([]Category, 0, len(resp.Categories))
	for _, cat := range resp.Categories {
		categories = append(categories, Category{
			ID:          cat.Id,
			Name:        cat.Name,
			Description: cat.Description,
			DefaultLine: int(cat.DefaultLine),
		})
	}

	return categories, nil
}

func (c *supportClient) GetTemplates(ctx context.Context) ([]Template, error) {
	resp, err := c.client.GetTemplates(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, ErrInternal
	}

	templates := make([]Template, 0, len(resp.Templates))
	for _, tpl := range resp.Templates {
		templates = append(templates, Template{
			ID:      tpl.Id,
			Name:    tpl.Name,
			Content: tpl.Content,
		})
	}

	return templates, nil
}

func (c *supportClient) GetStats(ctx context.Context) (SupportStats, error) {
	resp, err := c.client.GetStats(ctx, &emptypb.Empty{})
	if err != nil {
		return SupportStats{}, ErrInternal
	}

	return SupportStats{
		TotalTickets:         resp.TotalTickets,
		ByStatus:             resp.ByStatus,
		ByCategory:           resp.ByCategory,
		AverageRating:        resp.AverageRating,
		AvgResolutionTimeSec: resp.AvgResolutionTimeSec,
	}, nil
}
