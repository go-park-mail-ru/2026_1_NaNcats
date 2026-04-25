package grpc

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/support/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/support/internal/usecase"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/grpcutil"
	pb "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/support"

	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func mapDomainToPBTicket(t domain.Ticket) *pb.TicketResponse {
	var assigneeID int64
	if t.AssigneeID != nil {
		assigneeID = *t.AssigneeID
	}

	var rating int32
	if t.ResolutionRating != nil {
		rating = int32(*t.ResolutionRating)
	}

	return &pb.TicketResponse{
		Id:               t.ID,
		PublicId:         t.PublicID,
		CategoryId:       t.CategoryID,
		CurrentStatus:    t.CurrentStatus,
		SupportLine:      int64(t.SupportLine),
		AssigneeId:       assigneeID,
		ResolutionRating: rating,
		CreatedAt:        timestamppb.New(t.CreatedAt),
	}
}

func mapDomainToPBEvent(e domain.Event) *pb.Event {
	var authorID int64
	if e.AuthorID != nil {
		authorID = *e.AuthorID
	}

	return &pb.Event{
		Id:         e.ID,
		TicketId:   e.TicketID,
		AuthorId:   authorID,
		AuthorRole: e.AuthorRole,
		EventType:  e.EventType,
		Payload:    string(e.Payload), // json.RawMessage -> string
		CreatedAt:  timestamppb.New(e.CreatedAt),
	}
}

type SupportHandler struct {
	pb.UnimplementedSupportServiceServer
	usecase usecase.SupportUseCase
}

func NewSupportHandler(uc usecase.SupportUseCase) *SupportHandler {
	return &SupportHandler{
		usecase: uc,
	}
}

func (h *SupportHandler) CreateTicket(ctx context.Context, req *pb.CreateTicketRequest) (*pb.CreateTicketResponse, error) {
	input := domain.CreateTicketInput{
		ContactEmail:   req.ContactEmail,
		CategoryID:     req.CategoryId,
		FirstMessage:   req.InitialMessage,
		ClientMeta:     []byte(req.ClientMeta),
		CreatorRole:    "user",
		IdempotencyKey: req.IdempotencyKey,
	}

	if req.ClientId != 0 {
		cid := req.ClientId
		input.ClientID = &cid
	}
	if req.GuestId != "" {
		gid := req.GuestId
		input.GuestID = &gid
	}

	publicID, err := h.usecase.CreateTicket(ctx, input)
	if err != nil {
		return nil, grpcutil.ToGRPCError(err)
	}

	return &pb.CreateTicketResponse{
		PublicId: publicID,
	}, nil
}

func (h *SupportHandler) SendMessage(ctx context.Context, req *pb.SendMessageRequest) (*pb.EventResponse, error) {
	var authorID *int64
	if req.AuthorId != 0 {
		aid := req.AuthorId
		authorID = &aid
	}

	err := h.usecase.AddMessage(ctx, req.TicketPublicId, authorID, req.AuthorRole, req.Message, req.IdempotencyKey)
	if err != nil {
		return nil, grpcutil.ToGRPCError(err)
	}

	// Так как usecase.AddMessage пока возвращает только ошибку,
	// отдаем успешный респонс (id и дата сгенерируются в БД)
	// Для идеальной работы можно доработать usecase, чтобы он возвращал ID созданного эвента
	return &pb.EventResponse{
		Id:        0,
		CreatedAt: timestamppb.Now(),
	}, nil
}

func (h *SupportHandler) GetUserTickets(ctx context.Context, req *pb.GetUserTicketsRequest) (*pb.TicketListResponse, error) {
	var clientID *int64
	if req.ClientId != 0 {
		cid := req.ClientId
		clientID = &cid
	}

	var guestID *string
	if req.GuestId != "" {
		gid := req.GuestId
		guestID = &gid
	}

	tickets, err := h.usecase.GetMyTickets(ctx, clientID, guestID)
	if err != nil {
		return nil, grpcutil.ToGRPCError(err)
	}

	pbTickets := make([]*pb.TicketResponse, 0, len(tickets))
	for _, t := range tickets {
		pbTickets = append(pbTickets, mapDomainToPBTicket(t))
	}

	return &pb.TicketListResponse{
		Tickets: pbTickets,
	}, nil
}

func (h *SupportHandler) GetTicketEvents(ctx context.Context, req *pb.GetTicketEventsRequest) (*pb.EventListResponse, error) {
	var clientID *int64
	if req.ClientId != 0 {
		cid := req.ClientId
		clientID = &cid
	}

	var guestID *string
	if req.GuestId != "" {
		gid := req.GuestId
		guestID = &gid
	}

	events, err := h.usecase.GetTicketEvents(ctx, req.TicketPublicId, clientID, guestID)
	if err != nil {
		return nil, grpcutil.ToGRPCError(err)
	}

	pbEvents := make([]*pb.Event, 0, len(events))
	for _, e := range events {
		pbEvents = append(pbEvents, mapDomainToPBEvent(e))
	}

	return &pb.EventListResponse{
		Events: pbEvents,
	}, nil
}

func (h *SupportHandler) RateTicket(ctx context.Context, req *pb.RateTicketRequest) (*pb.SuccessResponse, error) {
	var clientID *int64
	if req.ClientId != 0 {
		cid := req.ClientId
		clientID = &cid
	}

	err := h.usecase.RateTicket(ctx, req.TicketPublicId, int(req.Rating), clientID, req.IdempotencyKey)
	if err != nil {
		return nil, grpcutil.ToGRPCError(err)
	}

	return &pb.SuccessResponse{Success: true}, nil
}

func (h *SupportHandler) GetAssignedTickets(ctx context.Context, req *pb.GetAssignedTicketsRequest) (*pb.TicketListResponse, error) {
	tickets, err := h.usecase.GetAssignedTickets(ctx, req.AgentId)
	if err != nil {
		return nil, grpcutil.ToGRPCError(err)
	}

	pbTickets := make([]*pb.TicketResponse, 0, len(tickets))
	for _, t := range tickets {
		pbTickets = append(pbTickets, mapDomainToPBTicket(t))
	}

	return &pb.TicketListResponse{
		Tickets: pbTickets,
	}, nil
}

func (h *SupportHandler) ChangeTicketStatus(ctx context.Context, req *pb.ChangeTicketStatusRequest) (*pb.SuccessResponse, error) {
	err := h.usecase.ChangeTicketStatus(ctx, req.TicketPublicId, req.Status, req.AgentId, req.IdempotencyKey)
	if err != nil {
		return nil, grpcutil.ToGRPCError(err)
	}

	return &pb.SuccessResponse{Success: true}, nil
}

func (h *SupportHandler) ReassignTicket(ctx context.Context, req *pb.ReassignTicketRequest) (*pb.SuccessResponse, error) {
	err := h.usecase.ReassignTicket(ctx, req.TicketPublicId, req.AgentId, int(req.Line), req.AuthorId, req.IdempotencyKey)
	if err != nil {
		return nil, grpcutil.ToGRPCError(err)
	}

	return &pb.SuccessResponse{Success: true}, nil
}

func (h *SupportHandler) SetAgentStatus(ctx context.Context, req *pb.SetAgentStatusRequest) (*pb.SuccessResponse, error) {
	err := h.usecase.SetAgentStatus(ctx, req.AgentId, req.Status)
	if err != nil {
		return nil, grpcutil.ToGRPCError(err)
	}

	return &pb.SuccessResponse{Success: true}, nil
}

func (h *SupportHandler) GetCategories(ctx context.Context, req *emptypb.Empty) (*pb.CategoryListResponse, error) {
	categories, err := h.usecase.GetCategories(ctx)
	if err != nil {
		return nil, grpcutil.ToGRPCError(err)
	}

	pbCategories := make([]*pb.Category, 0, len(categories))
	for _, c := range categories {
		pbCategories = append(pbCategories, &pb.Category{
			Id:          c.ID,
			Name:        c.Name,
			Description: c.Description,
			DefaultLine: int32(c.DefaultLine),
		})
	}

	return &pb.CategoryListResponse{
		Categories: pbCategories,
	}, nil
}

func (h *SupportHandler) GetTemplates(ctx context.Context, req *emptypb.Empty) (*pb.TemplateListResponse, error) {
	templates, err := h.usecase.GetTemplates(ctx)
	if err != nil {
		return nil, grpcutil.ToGRPCError(err)
	}

	pbTemplates := make([]*pb.Template, 0, len(templates))
	for _, t := range templates {
		pbTemplates = append(pbTemplates, &pb.Template{
			Id:      t.ID,
			Name:    t.Name,
			Content: t.Content,
		})
	}

	return &pb.TemplateListResponse{
		Templates: pbTemplates,
	}, nil
}
