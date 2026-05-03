package grpc

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/restaurant/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/restaurant/internal/usecase"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/grpcutil"
	pb "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/restaurant"
	"google.golang.org/protobuf/types/known/emptypb"
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

func (h *RestaurantHandler) CreateRestaurantBrand(ctx context.Context, req *pb.CreateBrandRequest) (*pb.RestaurantBrand, error) {
	domainBrand := domain.RestaurantBrand{
		OwnerProfileID: req.OwnerId,
		Name:           req.Name,
		Description:    req.Description,
	}

	brand, err := h.brandUC.CreateRestaurantBrand(ctx, domainBrand, req.LogoData, req.IdempotencyKey)
	if err != nil {
		return nil, grpcutil.ToGRPCError(err)
	}

	return mapDomainToPBRestaurant(brand), nil
}

func (h *RestaurantHandler) UpdateRestaurantBrand(ctx context.Context, req *pb.UpdateBrandRequest) (*pb.RestaurantBrand, error) {
	domainBrand := domain.RestaurantBrand{
		ID: req.Id,
	}

	// Обработка optional полей из proto3
	if req.Name != nil {
		domainBrand.Name = *req.Name
	}
	if req.Description != nil {
		domainBrand.Description = *req.Description
	}
	if req.PromotionTier != nil {
		domainBrand.PromotionTier = int(*req.PromotionTier)
	}

	brand, err := h.brandUC.UpdateRestaurantBrand(ctx, domainBrand, req.LogoData, req.IdempotencyKey)
	if err != nil {
		return nil, grpcutil.ToGRPCError(err)
	}

	return mapDomainToPBRestaurant(brand), nil
}

func (h *RestaurantHandler) DeleteRestaurantBrand(ctx context.Context, req *pb.DeleteBrandRequest) (*emptypb.Empty, error) {
	err := h.brandUC.DeleteRestaurantBrand(ctx, req.Id)
	if err != nil {
		return nil, grpcutil.ToGRPCError(err)
	}
	return &emptypb.Empty{}, nil
}

func (h *RestaurantHandler) CreateDish(ctx context.Context, req *pb.CreateDishRequest) (*pb.Dish, error) {
	domainDish := domain.Dish{
		RestaurantBrandID: req.RestaurantBrandId,
		Name:              req.Name,
		Description:       req.Description,
		Price:             req.Price,
	}

	dish, err := h.dishUC.CreateDish(ctx, domainDish, req.ImageData, req.IdempotencyKey)
	if err != nil {
		return nil, grpcutil.ToGRPCError(err)
	}

	return mapDomainToPBDish(dish), nil
}

func (h *RestaurantHandler) UpdateDish(ctx context.Context, req *pb.UpdateDishRequest) (*pb.Dish, error) {
	domainDish := domain.Dish{
		ID: req.Id,
	}

	if req.Name != nil {
		domainDish.Name = *req.Name
	}
	if req.Description != nil {
		domainDish.Description = *req.Description
	}
	if req.Price != nil {
		domainDish.Price = *req.Price
	}

	dish, err := h.dishUC.UpdateDish(ctx, domainDish, req.ImageData, req.IdempotencyKey)
	if err != nil {
		return nil, grpcutil.ToGRPCError(err)
	}

	return mapDomainToPBDish(dish), nil
}

func (h *RestaurantHandler) DeleteDish(ctx context.Context, req *pb.DeleteDishRequest) (*emptypb.Empty, error) {
	err := h.dishUC.DeleteDish(ctx, req.Id)
	if err != nil {
		return nil, grpcutil.ToGRPCError(err)
	}
	return &emptypb.Empty{}, nil
}
