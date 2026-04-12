package usecase

import (
	"context"
	"errors"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/repository"
)

type DishUseCase interface {
	GetDishesByRestaurantBrandID(ctx context.Context, restaurantBrandID, limit, offset int) ([]domain.Dish, error)
}

type dishUseCase struct {
	dishRepo repository.DishRepository
}

func NewDishUseCase(dr repository.DishRepository) DishUseCase {
	return &dishUseCase{
		dishRepo: dr,
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

	return uc.dishRepo.GetDishesByRestaurantBrandID(ctx, restaurantBrandID, limit, offset)
}
