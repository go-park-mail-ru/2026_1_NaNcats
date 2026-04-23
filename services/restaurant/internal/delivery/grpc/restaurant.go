package grpc

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/restaurant/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/restaurant/internal/usecase"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/grpcutil"
	pb "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/restaurant"
)

func mapDomainToPBRestaurant(b domain.RestaurantBrand) *pb.RestaurantBrand {
	return &pb.RestaurantBrand{
		Id:             b.ID,
		OwnerProfileId: b.OwnerProfileID,
		Name:           b.Name,
		Description:    b.Description,
		PromotionTier:  int32(b.PromotionTier),
		LogoUrl:        b.LogoURL,
	}
}

func mapDomainToPBDish(d domain.Dish) *pb.Dish {
	return &pb.Dish{
		Id:                d.ID,
		RestaurantBrandId: d.RestaurantBrandID,
		Name:              d.Name,
		Description:       d.Description,
		Price:             d.Price,
		ImageUrl:          d.ImageURL,
	}
}

type RestaurantHandler struct {
	pb.UnimplementedRestaurantServiceServer
	brandUC usecase.RestaurantBrandUseCase
	dishUC  usecase.DishUseCase
}

func NewRestaurantHandler(buc usecase.RestaurantBrandUseCase, duc usecase.DishUseCase) *RestaurantHandler {
	return &RestaurantHandler{
		brandUC: buc,
		dishUC:  duc,
	}
}

// Получение списка ресторанов
func (h *RestaurantHandler) GetRestaurantBrandsList(ctx context.Context, req *pb.GetRestaurantBrandsListRequest) (*pb.GetRestaurantBrandsListResponse, error) {
	brands, err := h.brandUC.GetRestaurantBrandsList(ctx, int(req.Limit), int(req.Offset))
	if err != nil {
		return nil, grpcutil.ToGRPCError(err)
	}

	pbBrands := make([]*pb.RestaurantBrand, 0, len(brands))
	for _, b := range brands {
		pbBrands = append(pbBrands, mapDomainToPBRestaurant(b))
	}

	return &pb.GetRestaurantBrandsListResponse{
		RestaurantBrands: pbBrands,
	}, nil
}

// Получение одного ресторана по ID
func (h *RestaurantHandler) GetRestaurantBrandByID(ctx context.Context, req *pb.GetRestaurantBrandByIDRequest) (*pb.GetRestaurantBrandByIDResponse, error) {
	brand, err := h.brandUC.GetRestaurantBrandByID(ctx, req.Id)
	if err != nil {
		return nil, grpcutil.ToGRPCError(err)
	}

	return &pb.GetRestaurantBrandByIDResponse{
		RestaurantBrand: mapDomainToPBRestaurant(brand),
	}, nil
}

// Пакетное получения данных о ресторанах
func (h *RestaurantHandler) GetRestaurantBrandsByIDs(ctx context.Context, req *pb.GetRestaurantBrandsByIDsRequest) (*pb.GetRestaurantBrandsByIDsResponse, error) {
	brands, err := h.brandUC.GetRestaurantBrandsByIDs(ctx, req.BrandIds)
	if err != nil {
		return nil, grpcutil.ToGRPCError(err)
	}

	pbBrands := make([]*pb.RestaurantBrand, 0, len(brands))
	for _, b := range brands {
		pbBrands = append(pbBrands, mapDomainToPBRestaurant(b))
	}

	return &pb.GetRestaurantBrandsByIDsResponse{
		RestaurantBrands: pbBrands,
	}, nil
}

// Меню конкретного ресторана
func (h *RestaurantHandler) GetDishesByRestaurantBrandID(ctx context.Context, req *pb.GetDishesByRestaurantBrandIDRequest) (*pb.GetDishesByRestaurantBrandIDResponse, error) {
	dishes, err := h.dishUC.GetDishesByRestaurantBrandID(ctx, req.RestaurantBrandId, int(req.Limit), int(req.Offset))
	if err != nil {
		return nil, grpcutil.ToGRPCError(err)
	}

	pbDishes := make([]*pb.Dish, 0, len(dishes))
	for _, d := range dishes {
		pbDishes = append(pbDishes, mapDomainToPBDish(d))
	}

	return &pb.GetDishesByRestaurantBrandIDResponse{
		Dishes: pbDishes,
	}, nil
}

// Получение конкретных блюд по айдишникам
func (h *RestaurantHandler) GetDishesByIDs(ctx context.Context, req *pb.GetDishesByIDsRequest) (*pb.GetDishesByIDsResponse, error) {
	dishes, err := h.dishUC.GetDishesByIDs(ctx, req.DishIds)
	if err != nil {
		return nil, grpcutil.ToGRPCError(err)
	}

	pbDishes := make([]*pb.Dish, 0, len(dishes))
	for _, d := range dishes {
		pbDishes = append(pbDishes, mapDomainToPBDish(d))
	}

	return &pb.GetDishesByIDsResponse{
		Dishes: pbDishes,
	}, nil
}
