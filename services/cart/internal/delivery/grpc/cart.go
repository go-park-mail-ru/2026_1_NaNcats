package grpc

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/cart/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/cart/internal/usecase"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/grpcutil"
	pb "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/cart"
	"google.golang.org/protobuf/types/known/emptypb"
)

func mapPBToDomainCart(p *pb.Cart) domain.Cart {
	if p == nil {
		return domain.Cart{}
	}
	items := make([]domain.CartItem, 0, len(p.Items))
	for _, item := range p.Items {
		items = append(items, domain.CartItem{
			DishID:   item.DishId,
			Quantity: int(item.Quantity),
		})
	}
	return domain.Cart{
		RestaurantBrandID: p.RestaurantBrandId,
		Items:             items,
	}
}

func mapDomainToPBCart(d domain.Cart) *pb.Cart {
	items := make([]*pb.CartItem, 0, len(d.Items))
	for _, item := range d.Items {
		items = append(items, &pb.CartItem{
			DishId:   item.DishID,
			Quantity: int32(item.Quantity),
			Name:     item.Name,
			Price:    item.Price,
			ImageUrl: item.ImageURL,
		})
	}
	return &pb.Cart{
		RestaurantBrandId: d.RestaurantBrandID,
		Items:             items,
	}
}

type CartHandler struct {
	pb.UnimplementedCartServiceServer
	usecase usecase.CartUseCase
}

func NewCartHandler(uc usecase.CartUseCase) *CartHandler {
	return &CartHandler{
		usecase: uc,
	}
}

func (h *CartHandler) GetCart(ctx context.Context, req *pb.GetCartRequest) (*pb.GetCartResponse, error) {
	cart, totalCost, err := h.usecase.GetCart(ctx, req.UserId)
	if err != nil {
		return nil, grpcutil.ToGRPCError(err)
	}

	return &pb.GetCartResponse{
		Cart:      mapDomainToPBCart(cart),
		TotalCost: totalCost,
	}, nil
}

func (h *CartHandler) UpdateCart(ctx context.Context, req *pb.UpdateCartRequest) (*emptypb.Empty, error) {
	domainCart := mapPBToDomainCart(req.CartData)
	err := h.usecase.UpdateCart(ctx, req.UserId, domainCart, req.IdempotencyKey)
	if err != nil {
		return nil, grpcutil.ToGRPCError(err)
	}

	return &emptypb.Empty{}, nil
}

func (h *CartHandler) LockCart(ctx context.Context, req *pb.CartOperationRequest) (*emptypb.Empty, error) {
	err := h.usecase.LockCart(ctx, req.UserId, req.IdempotencyKey)
	if err != nil {
		return nil, grpcutil.ToGRPCError(err)
	}
	return &emptypb.Empty{}, nil
}

func (h *CartHandler) UnlockCart(ctx context.Context, req *pb.CartOperationRequest) (*emptypb.Empty, error) {
	err := h.usecase.UnlockCart(ctx, req.UserId, req.IdempotencyKey)
	if err != nil {
		return nil, grpcutil.ToGRPCError(err)
	}
	return &emptypb.Empty{}, nil
}

func (h *CartHandler) ClearCart(ctx context.Context, req *pb.CartOperationRequest) (*emptypb.Empty, error) {
	err := h.usecase.ClearCart(ctx, req.UserId, req.IdempotencyKey)
	if err != nil {
		return nil, grpcutil.ToGRPCError(err)
	}
	return &emptypb.Empty{}, nil
}
