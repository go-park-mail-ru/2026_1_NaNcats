package grpc

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/order/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/order/internal/usecase"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/grpcutil"
	pb "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/order"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func mapDomainToPBOrder(o domain.Order) *pb.Order {
	items := make([]*pb.OrderDish, 0, len(o.Items))
	for _, item := range o.Items {
		pbItem := &pb.OrderDish{
			DishId:   item.DishID,
			Quantity: int32(item.Quantity),
			Price:    item.Price,
		}
		if item.OwnerUserID != nil {
			pbItem.OwnerUserId = item.OwnerUserID
		}
		items = append(items, pbItem)
	}

	splits := make([]*pb.OrderSplit, 0, len(o.Splits))
	for _, split := range o.Splits {
		splits = append(splits, &pb.OrderSplit{
			SplitId: split.ID,
			UserId:  split.UserID,
			Amount:  split.Amount,
			Status:  split.Status,
		})
	}

	return &pb.Order{
		PublicId:          o.PublicID,
		RestaurantBrandId: o.RestaurantBrandID,
		RestaurantName:    o.RestaurantName,
		RestaurantLogoUrl: o.RestaurantLogoURL,
		TotalCost:         o.TotalCost,
		Status:            o.Status,
		CreatedAt:         timestamppb.New(o.CreatedAt),
		Items:             items,
		Splits:            splits,
	}
}

func mapCreateOrderInputFromPB(req *pb.CreateOrderRequest) domain.CreateOrderInput {
	return domain.CreateOrderInput{
		UserID:             req.UserId,
		AddressPublicID:    req.AddressPublicId,
		RestaurantBranchID: req.RestaurantBranchId,
		RestaurantBrandID:  req.RestaurantBrandId,
		DeliveryCost:       req.DeliveryCost,
		ServiceFee:         req.ServiceFee,
		PaymentMethodID:    req.PaymentMethodId,
		PayForAll:          req.PayForAll,
		PayerMapping:       req.PayerMapping,
	}
}

type OrderHandler struct {
	pb.UnimplementedOrderServiceServer
	usecase usecase.OrderUseCase
}

func NewOrderHandler(uc usecase.OrderUseCase) *OrderHandler {
	return &OrderHandler{
		usecase: uc,
	}
}

func (h *OrderHandler) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*pb.CreateOrderResponse, error) {
	input := mapCreateOrderInputFromPB(req)

	orderPublicID, err := h.usecase.CreateOrder(ctx, input, req.IdempotencyKey)
	if err != nil {
		return nil, grpcutil.ToGRPCError(err)
	}

	return &pb.CreateOrderResponse{
		OrderPublicId: orderPublicID,
	}, nil

}

func (h *OrderHandler) GetOrders(ctx context.Context, req *pb.GetOrdersRequest) (*pb.GetOrdersResponse, error) {
	orders, err := h.usecase.GetOrders(ctx, req.UserId)
	if err != nil {
		return nil, grpcutil.ToGRPCError(err)
	}

	pbOrders := make([]*pb.Order, 0, len(orders))
	for _, o := range orders {
		pbOrders = append(pbOrders, mapDomainToPBOrder(o))
	}

	return &pb.GetOrdersResponse{
		Orders: pbOrders,
	}, nil

}

func (h *OrderHandler) UpdateOrderStatusByPaymentID(ctx context.Context, req *pb.UpdateStatusRequest) (*emptypb.Empty, error) {
	err := h.usecase.UpdateOrderStatusByPaymentID(ctx, req.YookassaPaymentId, req.Status, req.IdempotencyKey)
	return &emptypb.Empty{}, grpcutil.ToGRPCError(err)
}

func (h *OrderHandler) PayForFriend(ctx context.Context, req *pb.PayForFriendRequest) (*pb.PayForFriendResponse, error) {
	err := h.usecase.PayForFriend(ctx, req.SplitId, req.PayerUserId, req.PaymentMethodId, req.IdempotencyKey)
	if err != nil {
		return nil, grpcutil.ToGRPCError(err)
	}

	return &pb.PayForFriendResponse{}, nil
}

func (h *OrderHandler) GetOrderPaymentID(ctx context.Context, req *pb.GetOrderPaymentIDRequest) (*pb.GetOrderPaymentIDResponse, error) {
	paymentID, err := h.usecase.GetOrderPaymentID(ctx, req.OrderPublicId, req.UserId)
	if err != nil {
		return nil, grpcutil.ToGRPCError(err)
	}
	return &pb.GetOrderPaymentIDResponse{YookassaPaymentId: paymentID}, nil
}

func (h *OrderHandler) CancelOrder(ctx context.Context, req *pb.CancelOrderRequest) (*emptypb.Empty, error) {
	if err := h.usecase.CancelOrder(ctx, req.OrderPublicId, req.UserId); err != nil {
		return nil, grpcutil.ToGRPCError(err)
	}
	return &emptypb.Empty{}, nil
}
