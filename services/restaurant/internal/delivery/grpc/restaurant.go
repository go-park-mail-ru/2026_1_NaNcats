package grpc

import (
	"context"
	"net/url"
	"strconv"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/restaurant/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/restaurant/internal/repository"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/restaurant/internal/usecase"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/grpcutil"
	pb "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/restaurant"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
)

// decodeMetadataValue safely URL-decodes a gRPC metadata value (used to carry
// non-ASCII strings like Cyrillic queries through the wire).
func decodeMetadataValue(v string) string {
	if decoded, err := url.QueryUnescape(v); err == nil {
		return decoded
	}
	return v
}

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
	brandUC     usecase.RestaurantBrandUseCase
	dishUC      usecase.DishUseCase
	extRepo     repository.ExtendedRestaurantRepository
	extDishRepo repository.ExtendedDishRepository
}

func NewRestaurantHandler(buc usecase.RestaurantBrandUseCase, duc usecase.DishUseCase, extRepo repository.ExtendedRestaurantRepository, extDishRepo repository.ExtendedDishRepository) *RestaurantHandler {
	return &RestaurantHandler{
		brandUC:     buc,
		dishUC:      duc,
		extRepo:     extRepo,
		extDishRepo: extDishRepo,
	}
}

// Получение списка ресторанов (поддерживает фильтрацию по категории и поиск через gRPC metadata)
func (h *RestaurantHandler) GetRestaurantBrandsList(ctx context.Context, req *pb.GetRestaurantBrandsListRequest) (*pb.GetRestaurantBrandsListResponse, error) {
	var brands []domain.RestaurantBrand
	var err error

	md, hasMD := metadata.FromIncomingContext(ctx)

	if hasMD {
		// Поиск по запросу (URL-encoded, чтобы пройти ASCII-only ограничение gRPC metadata)
		if vals := md.Get("x-search-query"); len(vals) > 0 && vals[0] != "" {
			query := decodeMetadataValue(vals[0])
			brands, err = h.extRepo.SearchRestaurantBrands(ctx, query, int(req.Limit), int(req.Offset))
			if err != nil {
				return nil, grpcutil.ToGRPCError(err)
			}
			pbBrands := make([]*pb.RestaurantBrand, 0, len(brands))
			for _, b := range brands {
				pbBrands = append(pbBrands, mapDomainToPBRestaurant(b))
			}
			return &pb.GetRestaurantBrandsListResponse{RestaurantBrands: pbBrands}, nil
		}

		// Фильтрация по имени категории (URL-encoded для не-ASCII символов)
		if vals := md.Get("x-category-name"); len(vals) > 0 && vals[0] != "" {
			catName := decodeMetadataValue(vals[0])
			brands, err = h.extRepo.GetRestaurantBrandsByCategoryName(ctx, catName, int(req.Limit), int(req.Offset))
			if err != nil {
				return nil, grpcutil.ToGRPCError(err)
			}
			pbBrands := make([]*pb.RestaurantBrand, 0, len(brands))
			for _, b := range brands {
				pbBrands = append(pbBrands, mapDomainToPBRestaurant(b))
			}
			return &pb.GetRestaurantBrandsListResponse{RestaurantBrands: pbBrands}, nil
		}

		// Фильтрация по числовому ID категории (legacy)
		if vals := md.Get("x-category-id"); len(vals) > 0 && vals[0] != "" {
			if catID, parseErr := strconv.ParseInt(vals[0], 10, 64); parseErr == nil && catID > 0 {
				brands, err = h.extRepo.GetRestaurantBrandsByCategory(ctx, catID, int(req.Limit), int(req.Offset))
				if err != nil {
					return nil, grpcutil.ToGRPCError(err)
				}
				pbBrands := make([]*pb.RestaurantBrand, 0, len(brands))
				for _, b := range brands {
					pbBrands = append(pbBrands, mapDomainToPBRestaurant(b))
				}
				return &pb.GetRestaurantBrandsListResponse{RestaurantBrands: pbBrands}, nil
			}
		}
	}

	brands, err = h.brandUC.GetRestaurantBrandsList(ctx, int(req.Limit), int(req.Offset))
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

// Меню конкретного ресторана.
// Поддерживает поиск по блюдам внутри бренда через gRPC metadata `x-dish-search`.
func (h *RestaurantHandler) GetDishesByRestaurantBrandID(ctx context.Context, req *pb.GetDishesByRestaurantBrandIDRequest) (*pb.GetDishesByRestaurantBrandIDResponse, error) {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if vals := md.Get("x-dish-search"); len(vals) > 0 && vals[0] != "" {
			query := decodeMetadataValue(vals[0])
			dishes, err := h.extDishRepo.SearchDishesByBrand(ctx, req.RestaurantBrandId, query, int(req.Limit))
			if err != nil {
				return nil, grpcutil.ToGRPCError(err)
			}
			pbDishes := make([]*pb.Dish, 0, len(dishes))
			for _, d := range dishes {
				pbDishes = append(pbDishes, mapDomainToPBDish(d))
			}
			return &pb.GetDishesByRestaurantBrandIDResponse{Dishes: pbDishes}, nil
		}
	}

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

// Получение конкретных блюд по айдишникам.
// Поддерживает глобальный поиск по всем блюдам через metadata `x-dish-search`
// (если задан, dish_ids игнорируются и возвращаются результаты поиска).
func (h *RestaurantHandler) GetDishesByIDs(ctx context.Context, req *pb.GetDishesByIDsRequest) (*pb.GetDishesByIDsResponse, error) {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if vals := md.Get("x-dish-search"); len(vals) > 0 && vals[0] != "" {
			query := decodeMetadataValue(vals[0])
			limit := 20
			if vlim := md.Get("x-dish-search-limit"); len(vlim) > 0 && vlim[0] != "" {
				if n, err := strconv.Atoi(vlim[0]); err == nil && n > 0 {
					limit = n
				}
			}
			dishes, err := h.extDishRepo.SearchDishes(ctx, query, limit)
			if err != nil {
				return nil, grpcutil.ToGRPCError(err)
			}
			pbDishes := make([]*pb.Dish, 0, len(dishes))
			for _, d := range dishes {
				pbDishes = append(pbDishes, mapDomainToPBDish(d))
			}
			return &pb.GetDishesByIDsResponse{Dishes: pbDishes}, nil
		}
	}

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
