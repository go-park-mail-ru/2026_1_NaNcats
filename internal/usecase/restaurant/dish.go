package restaurant

import (
	"context"
	"errors"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/repository"
)

//go:generate mockgen -destination=mocks/dish_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/internal/usecase/restaurant DishUseCase
type DishUseCase interface {
	GetDishesByRestaurantBrandID(ctx context.Context, restaurantBrandID, limit, offset int) ([]domain.Dish, error)
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

func (uc *dishUseCase) GetDishesByRestaurantBrandID(ctx context.Context, restaurantBrandID, limit, offset int) ([]domain.Dish, error) {
	// Валидация входа
	if restaurantBrandID <= 0 {
		return nil, errors.New("invalid restaurant_brand_id")
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
		return []domain.Dish{}, err
	}

	for i, dish := range dishes {
		if dish.ImageURL == "" {
			dishes[i].ImageURL = uc.defaultFoodLogoURL
		}
	}

	return dishes, nil
}
