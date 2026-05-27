package repository

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/restaurant/internal/domain"
)

//go:generate mockgen -destination=mocks/dish_repository_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/services/restaurant/internal/repository DishRepository
type DishRepository interface {
	GetDishesByRestaurantBrandID(ctx context.Context, restaurantBrandID int64, limit, offset int) ([]domain.Dish, error)
	GetDishByID(ctx context.Context, DishID int64) (domain.Dish, error)
	GetDishesByIDs(ctx context.Context, ids []int64) ([]domain.Dish, error)
	Create(ctx context.Context, d domain.Dish, idemKey string) (domain.Dish, error)
	Update(ctx context.Context, d domain.Dish) (domain.Dish, error)
	Delete(ctx context.Context, id int64) error

	SearchDishes(ctx context.Context, query string, limit int) ([]domain.Dish, error)
	SearchDishesByBrand(ctx context.Context, brandID int64, query string, limit int) ([]domain.Dish, error)
}
