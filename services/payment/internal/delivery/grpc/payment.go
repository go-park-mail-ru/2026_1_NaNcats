package grpc

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/payment/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/payment/internal/usecase"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/api_clients/yookassa"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/grpcutil"
	pb "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/payment"

	"google.golang.org/protobuf/types/known/emptypb"
)

func mapDomainToPBPaymentMethod(pm domain.PaymentMethod) *pb.PaymentMethod {
	return &pb.PaymentMethod{
		Id:         pm.ID,
		UserId:     pm.UserID,
		ExternalId: pm.ExternalID,
		CardType:   pm.CardType,
		Last4:      pm.Last4,
		IssuerName: pm.IssuerName,
		IsDefault:  pm.IsDefault,
	}
}

type PaymentHandler struct {
	pb.UnimplementedPaymentServiceServer
	usecase usecase.PaymentUseCase
}

func NewPaymentHandler(uc usecase.PaymentUseCase) *PaymentHandler {
	return &PaymentHandler{
		usecase: uc,
	}
}

func (h *PaymentHandler) CreatePayment(ctx context.Context, req *pb.CreatePaymentRequest) (*pb.CreatePaymentResponse, error) {
	paymentID, confirmationURL, err := h.usecase.CreatePayment(ctx, req.Amount, req.PaymentMethodId, req.IdempotencyKey)
	if err != nil {
		return nil, grpcutil.ToGRPCError(err)
	}

	return &pb.CreatePaymentResponse{
		PaymentId:       paymentID,
		ConfirmationUrl: confirmationURL,
	}, nil
}

func (h *PaymentHandler) InitiateCardBinding(ctx context.Context, req *pb.InitiateCardBindingRequest) (*pb.InitiateCardBindingResponse, error) {
	confirmationURL, err := h.usecase.InitiateCardBinding(ctx, req.UserId, req.IdempotencyKey)
	if err != nil {
		return nil, grpcutil.ToGRPCError(err)
	}

	return &pb.InitiateCardBindingResponse{
		ConfirmationUrl: confirmationURL,
	}, nil
}

func (h *PaymentHandler) GetUserCards(ctx context.Context, req *pb.GetUserCardsRequest) (*pb.GetUserCardsResponse, error) {
	cards, err := h.usecase.GetUserCards(ctx, req.UserId)
	if err != nil {
		return nil, grpcutil.ToGRPCError(err)
	}

	pbCards := make([]*pb.PaymentMethod, 0, len(cards))
	for _, c := range cards {
		pbCards = append(pbCards, mapDomainToPBPaymentMethod(c))
	}

	return &pb.GetUserCardsResponse{
		UserId: req.UserId,
		Cards:  pbCards,
	}, nil
}

func (h *PaymentHandler) SetDefaultCard(ctx context.Context, req *pb.ChangeCardRequest) (*emptypb.Empty, error) {
	err := h.usecase.SetDefaultCard(ctx, req.CardId, req.UserId, req.IdempotencyKey)
	if err != nil {
		return nil, grpcutil.ToGRPCError(err)
	}

	return &emptypb.Empty{}, nil
}

func (h *PaymentHandler) DeleteCard(ctx context.Context, req *pb.ChangeCardRequest) (*emptypb.Empty, error) {
	err := h.usecase.DeleteCard(ctx, req.CardId, req.UserId, req.IdempotencyKey)
	if err != nil {
		return nil, grpcutil.ToGRPCError(err)
	}

	return &emptypb.Empty{}, nil
}

func (h *PaymentHandler) ProcessPaymentMethodWebhook(ctx context.Context, req *pb.ProcessPaymentMethodWebhookRequest) (*emptypb.Empty, error) {
	var cardInfo *yookassa.PaymentMethodResponseCard
	if req.Card != nil {
		cardInfo = &yookassa.PaymentMethodResponseCard{
			First6:      req.Card.First6,
			Last4:       req.Card.Last4,
			ExpiryMonth: req.Card.ExpiryMonth,
			ExpiryYear:  req.Card.ExpiryYear,
			CardType:    req.Card.CardType,
			IssuerName:  req.Card.IssuerName,
		}
	}

	webhookObj := &yookassa.WebhookPaymentMethodObject{
		ID:     req.Id,
		Status: req.Status,
		Saved:  req.Saved,
		Type:   req.Type,
		Card:   cardInfo,
	}

	err := h.usecase.ProcessPaymentMethodWebhook(ctx, webhookObj)
	if err != nil {
		return nil, grpcutil.ToGRPCError(err)
	}

	return &emptypb.Empty{}, nil
}

func (h *PaymentHandler) ProcessPaymentWebhook(ctx context.Context, req *pb.ProcessPaymentWebhookRequest) (*emptypb.Empty, error) {
	webhookObj := &yookassa.WebhookPaymentObject{
		ID:     req.Id,
		Status: req.Status,
	}

	err := h.usecase.ProcessPaymentWebhook(ctx, webhookObj)
	if err != nil {
		return nil, grpcutil.ToGRPCError(err)
	}

	return &emptypb.Empty{}, nil
}

func (h *PaymentHandler) RefreshPaymentStatus(ctx context.Context, req *pb.RefreshPaymentStatusRequest) (*pb.RefreshPaymentStatusResponse, error) {
	statusStr, err := h.usecase.RefreshPaymentStatus(ctx, req.YookassaPaymentId)
	if err != nil {
		return nil, grpcutil.ToGRPCError(err)
	}
	return &pb.RefreshPaymentStatusResponse{Status: statusStr}, nil
}
