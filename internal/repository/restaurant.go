package repository

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
)

type RestaurantBrandRepository interface {
	GetRestaurantBrandsList(ctx context.Context, limit, offset int) ([]domain.RestaurantBrand, error)
	GetDishByID(ctx context.Context, dishID int) (domain.Dish, error)
}
