package repository

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
)

type DishRepository interface {
	// GetDishesByRestaurantBrandID возвращает блюда конкретного бренда ресторана
	GetDishesByRestaurantBrandID(ctx context.Context, restaurantBrandID, limit, offset int) ([]domain.Dish, error)

	GetDishByID(ctx context.Context, DishID int) (domain.Dish, error)
}
