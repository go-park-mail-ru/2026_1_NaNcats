package usecase

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/restaurant/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/restaurant/internal/repository"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/errutil"
	"google.golang.org/grpc/codes"
)

//go:generate mockgen -destination=mocks/dish_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/internal/usecase/restaurant DishUseCase
type DishUseCase interface {
	GetDishesByRestaurantBrandID(ctx context.Context, restaurantBrandID int64, limit, offset int) ([]domain.Dish, error)
	GetDishesByIDs(ctx context.Context, ids []int64) ([]domain.Dish, error)
}

type dishUseCase struct {
	dishRepo           repository.DishRepository
	defaultFoodLogoURL string
}

func NewDishUseCase(dr repository.DishRepository, dflurl string) DishUseCase {
	return &dishUseCase{
		dishRepo:           dr,
		defaultFoodLogoURL: dflurl,
	}
}

func (uc *dishUseCase) GetDishesByRestaurantBrandID(ctx context.Context, restaurantBrandID int64, limit, offset int) ([]domain.Dish, error) {
	// Валидация входа
	if restaurantBrandID <= 0 {
		return nil, errutil.New("invalid restaurant_brand_id", codes.InvalidArgument)
	}

	// Пагинация
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	dishes, err := uc.dishRepo.GetDishesByRestaurantBrandID(ctx, restaurantBrandID, limit, offset)
	if err != nil {
		return nil, err
	}

	for i, dish := range dishes {
		if dish.ImageURL == "" {
			dishes[i].ImageURL = uc.defaultFoodLogoURL
		}
	}

	return dishes, nil
}

func (uc *dishUseCase) GetDishesByIDs(ctx context.Context, ids []int64) ([]domain.Dish, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	dishes, err := uc.dishRepo.GetDishesByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	for i := range dishes {
		if dishes[i].ImageURL == "" {
			dishes[i].ImageURL = uc.defaultFoodLogoURL
		}
	}

	return dishes, nil
}
