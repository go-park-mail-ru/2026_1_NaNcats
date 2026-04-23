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

func mapDomainToPBOrder(d domain.Order) *pb.Order {
	items := make([]*pb.OrderDish, 0, len(d.Items))
	for _, item := range d.Items {
		items = append(items, &pb.OrderDish{
			DishId:   item.DishID,
			Quantity: int32(item.Quantity),
			Price:    item.Price,
		})
	}

	return &pb.Order{
		PublicId:          d.PublicID,
		RestaurantName:    d.RestaurantName,
		RestaurantLogoUrl: d.RestaurantLogoURL,
		TotalCost:         d.TotalCost,
		Status:            d.Status,
		CreatedAt:         timestamppb.New(d.CreatedAt),
		Items:             items,
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
	input := domain.CreateOrderInput{
		AddressPublicID:    req.AddressPublicId,
		RestaurantBranchID: req.RestaurantBranchId,
		RestaurantBrandID:  req.RestaurantBrandId,
		PaymentMethodID:    req.PaymentMethodId,
		DeliveryCost:       req.DeliveryCost,
		ServiceFee:         req.ServiceFee,
	}

	orderPublicID, confirmationURL, err := h.usecase.CreateOrder(ctx, req.UserId, input, req.IdempotencyKey)
	if err != nil {
		return nil, grpcutil.ToGRPCError(err)
	}

	return &pb.CreateOrderResponse{
		OrderPublicId:   orderPublicID,
		ConfirmationUrl: confirmationURL,
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
	if err != nil {
		return nil, grpcutil.ToGRPCError(err)
	}

	return &emptypb.Empty{}, nil
}
