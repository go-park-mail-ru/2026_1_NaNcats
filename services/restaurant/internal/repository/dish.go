package repository

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
)

//go:generate mockgen -destination=mocks/dish_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/internal/repository DishRepository
type DishRepository interface {
	// GetDishesByRestaurantBrandID возвращает блюда конкретного бренда ресторана
	GetDishesByRestaurantBrandID(ctx context.Context, restaurantBrandID int64, limit, offset int) ([]domain.Dish, error)
	GetDishByID(ctx context.Context, DishID int64) (domain.Dish, error)
	GetDishesByIDs(ctx context.Context, ids []int64) ([]domain.Dish, error)
}
