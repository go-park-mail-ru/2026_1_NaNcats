package grpc

import (
	"context"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/order/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/order/internal/usecase"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/grpcutil"
	pb "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/order"
	"google.golang.org/protobuf/types/known/emptypb"
)

func mapDomainToPBPromocode(p domain.Promocode) *pb.Promocode {
	out := &pb.Promocode{
		Id:                 p.ID,
		Code:               p.Code,
		Title:              p.Title,
		RestaurantBrandIds: p.RestaurantBrandIDs(),
	}
	if p.DiscountPercent != nil {
		v := int32(*p.DiscountPercent)
		out.DiscountPercent = &v
	}
	if p.DiscountAmount != nil {
		out.DiscountAmount = p.DiscountAmount
	}
	if p.MinOrderAmount != nil {
		out.MinOrderAmount = *p.MinOrderAmount
	}
	if !p.ExpiresAt.IsZero() {
		s := p.ExpiresAt.UTC().Format(time.RFC3339)
		out.ExpiresAt = &s
	}
	return out
}

func mapDomainToPBPromocodeList(promos []domain.Promocode) *pb.PromocodeList {
	out := make([]*pb.Promocode, 0, len(promos))
	for _, p := range promos {
		out = append(out, mapDomainToPBPromocode(p))
	}
	return &pb.PromocodeList{Promocodes: out}
}

type PromoHandler struct {
	pb.UnimplementedPromoServiceServer
	usecase usecase.PromoUseCase
}

func NewPromoHandler(uc usecase.PromoUseCase) *PromoHandler {
	return &PromoHandler{usecase: uc}
}

func (h *PromoHandler) GetUserPromos(ctx context.Context, req *pb.GetUserPromosRequest) (*pb.PromocodeList, error) {
	promos, err := h.usecase.GetUserPromos(ctx, req.UserId)
	if err != nil {
		return nil, grpcutil.ToGRPCError(err)
	}
	return mapDomainToPBPromocodeList(promos), nil
}

func (h *PromoHandler) GetRestaurantPromos(ctx context.Context, req *pb.GetRestaurantPromosRequest) (*pb.PromocodeList, error) {
	promos, err := h.usecase.GetRestaurantPromos(ctx, req.RestaurantBrandId)
	if err != nil {
		return nil, grpcutil.ToGRPCError(err)
	}
	return mapDomainToPBPromocodeList(promos), nil
}

func (h *PromoHandler) BindPromocode(ctx context.Context, req *pb.BindPromocodeRequest) (*pb.Promocode, error) {
	promo, err := h.usecase.BindPromo(ctx, req.UserId, req.Code)
	if err != nil {
		return nil, grpcutil.ToGRPCError(err)
	}
	return mapDomainToPBPromocode(promo), nil
}

func (h *PromoHandler) ValidatePromocode(ctx context.Context, req *pb.ValidatePromocodeRequest) (*pb.ValidatePromocodeResponse, error) {
	result, err := h.usecase.ValidatePromo(ctx, req.UserId, req.Code,
		req.RestaurantBrandId, req.OrderAmount, req.DeliveryCost, req.ServiceFee)
	if err != nil {
		return nil, grpcutil.ToGRPCError(err)
	}
	return &pb.ValidatePromocodeResponse{
		Valid:    result.Valid,
		Discount: result.Discount,
		Reason:   result.Reason,
	}, nil
}

func (h *PromoHandler) UsePromocode(ctx context.Context, req *pb.UsePromocodeRequest) (*emptypb.Empty, error) {
	if err := h.usecase.UsePromo(ctx, req.UserId, req.Code, req.OrderPublicId); err != nil {
		return nil, grpcutil.ToGRPCError(err)
	}
	return &emptypb.Empty{}, nil
}

func (h *PromoHandler) CreateAndBindWheelPromo(ctx context.Context, req *pb.CreateAndBindWheelPromoRequest) (*pb.Promocode, error) {
	var discountAmount *int64
	if req.DiscountAmount != nil {
		discountAmount = req.DiscountAmount
	}

	var discountPercent *int
	if req.DiscountPercent != nil {
		val := int(*req.DiscountPercent)
		discountPercent = &val
	}

	var brandID *int64
	if req.RestaurantBrandId != nil {
		brandID = req.RestaurantBrandId
	}

	var minOrderAmount *int64
	if req.MinOrderAmount != nil {
		minOrderAmount = req.MinOrderAmount
	}

	promo, err := h.usecase.CreateAndBindWheelPromo(
		ctx,
		req.UserId,
		req.Title,
		discountAmount,
		discountPercent,
		brandID,
		minOrderAmount,
	)
	if err != nil {
		return nil, grpcutil.ToGRPCError(err)
	}

	return mapDomainToPBPromocode(promo), nil
}
