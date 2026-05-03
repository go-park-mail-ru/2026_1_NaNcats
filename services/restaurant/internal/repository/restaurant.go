package repository

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/restaurant/internal/domain"
)

type RestaurantBrandRepository interface {
	GetRestaurantBrandsList(ctx context.Context, limit, offset int) ([]domain.RestaurantBrand, error)
	GetByID(ctx context.Context, id int64) (domain.RestaurantBrand, error)
	GetRestaurantBrandsByIDs(ctx context.Context, ids []int64) ([]domain.RestaurantBrand, error)
	Create(ctx context.Context, b domain.RestaurantBrand, idempotencyKey string) (domain.RestaurantBrand, error)
	Delete(ctx context.Context, id int64) error
	Update(ctx context.Context, b domain.RestaurantBrand) (domain.RestaurantBrand, error)
}
