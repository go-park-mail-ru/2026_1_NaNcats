package grpc

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/cart/internal/usecase"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/grpcutil"
	pb "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/cart"
	"google.golang.org/protobuf/types/known/emptypb"
)

type CartHandler struct {
	pb.UnimplementedCartServiceServer
	usecase usecase.CartUseCase
}

func NewCartHandler(u usecase.CartUseCase) *CartHandler {
	return &CartHandler{usecase: u}
}

func (h *CartHandler) GetCart(ctx context.Context, req *pb.GetCartRequest) (*pb.GetCartResponse, error) {
	cart, totalCost, err := h.usecase.GetCart(ctx, req.UserId)
	if err != nil {
		return nil, grpcutil.ToGRPCError(err)
	}

	pbCart := &pb.Cart{
		CartId:            cart.ID,
		AdminId:           cart.AdminID,
		RestaurantBrandId: cart.RestaurantBrandID,
		Mode:              mapModeToPb(cart.Mode),
		Status:            mapStatusToPb(cart.Status),
	}

	for _, item := range cart.Items {
		pbItem := &pb.CartItem{
			DishId:   item.DishID,
			Quantity: item.Quantity,
			Name:     item.Name,
			Price:    item.Price,
			ImageUrl: item.ImageURL,
		}
		if item.OwnerUserID != nil {
			pbItem.OwnerUserId = item.OwnerUserID
		}
		pbCart.Items = append(pbCart.Items, pbItem)
	}

	for _, member := range cart.Members {
		pbCart.Members = append(pbCart.Members, &pb.CartMember{
			UserId:   member.UserID,
			JoinedAt: member.JoinedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	return &pb.GetCartResponse{
		Cart:      pbCart,
		TotalCost: totalCost,
	}, nil
}

func (h *CartHandler) LockCart(ctx context.Context, req *pb.CartOperationRequest) (*emptypb.Empty, error) {
	err := h.usecase.LockCart(ctx, req.CartId, req.UserId, req.IdempotencyKey)
	return &emptypb.Empty{}, grpcutil.ToGRPCError(err)
}

func (h *CartHandler) UnlockCart(ctx context.Context, req *pb.CartOperationRequest) (*emptypb.Empty, error) {
	err := h.usecase.UnlockCart(ctx, req.CartId, req.UserId, req.IdempotencyKey)
	return &emptypb.Empty{}, grpcutil.ToGRPCError(err)
}

func (h *CartHandler) ClearCart(ctx context.Context, req *pb.CartOperationRequest) (*emptypb.Empty, error) {
	err := h.usecase.ClearCart(ctx, req.CartId, req.UserId, req.IdempotencyKey)
	return &emptypb.Empty{}, grpcutil.ToGRPCError(err)
}

func (h *CartHandler) GenerateInvite(ctx context.Context, req *pb.GenerateInviteRequest) (*pb.GenerateInviteResponse, error) {
	invite, err := h.usecase.GenerateInvite(ctx, req.CartId, req.UserId)
	if err != nil {
		return nil, grpcutil.ToGRPCError(err)
	}
	return &pb.GenerateInviteResponse{
		Token:     invite.Token,
		ExpiresAt: invite.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

func (h *CartHandler) JoinCart(ctx context.Context, req *pb.JoinCartRequest) (*pb.JoinCartResponse, error) {
	cartID, err := h.usecase.JoinCart(ctx, req.Token, req.UserId)
	if err != nil {
		return nil, grpcutil.ToGRPCError(err)
	}
	return &pb.JoinCartResponse{CartId: cartID}, nil
}

func (h *CartHandler) KickMember(ctx context.Context, req *pb.CartMemberOperationRequest) (*emptypb.Empty, error) {
	err := h.usecase.KickMember(ctx, req.CartId, req.AdminUserId, req.TargetUserId, req.IdempotencyKey)
	return &emptypb.Empty{}, grpcutil.ToGRPCError(err)
}

func (h *CartHandler) CloseSharedCart(ctx context.Context, req *pb.CartOperationRequest) (*emptypb.Empty, error) {
	err := h.usecase.CloseSharedCart(ctx, req.CartId, req.UserId, req.IdempotencyKey)
	return &emptypb.Empty{}, grpcutil.ToGRPCError(err)
}

func (h *CartHandler) AddItem(ctx context.Context, req *pb.AddItemRequest) (*emptypb.Empty, error) {
	err := h.usecase.AddItem(ctx, req.CartId, req.UserId, req.DishId, req.Quantity, req.IdempotencyKey)
	return &emptypb.Empty{}, grpcutil.ToGRPCError(err)
}

func (h *CartHandler) RemoveItem(ctx context.Context, req *pb.RemoveItemRequest) (*emptypb.Empty, error) {
	err := h.usecase.RemoveItem(ctx, req.CartId, req.UserId, req.DishId, req.IdempotencyKey)
	return &emptypb.Empty{}, grpcutil.ToGRPCError(err)
}

func (h *CartHandler) UpdateItemQuantity(ctx context.Context, req *pb.UpdateQuantityRequest) (*emptypb.Empty, error) {
	err := h.usecase.UpdateItemQuantity(ctx, req.CartId, req.UserId, req.DishId, req.NewQuantity, req.IdempotencyKey)
	return &emptypb.Empty{}, grpcutil.ToGRPCError(err)
}

func (h *CartHandler) ReassignItemOwner(ctx context.Context, req *pb.ReassignOwnerRequest) (*emptypb.Empty, error) {
	err := h.usecase.ReassignItemOwner(ctx, req.CartId, req.AdminUserId, req.DishId, req.NewOwnerUserId, req.IdempotencyKey)
	return &emptypb.Empty{}, grpcutil.ToGRPCError(err)
}

func mapModeToPb(mode string) pb.CartMode {
	if mode == "shared" {
		return pb.CartMode_CART_MODE_SHARED
	}
	return pb.CartMode_CART_MODE_SOLO
}

func mapStatusToPb(status string) pb.CartStatus {
	if status == "locked" {
		return pb.CartStatus_CART_STATUS_LOCKED
	}
	return pb.CartStatus_CART_STATUS_ACTIVE
}
