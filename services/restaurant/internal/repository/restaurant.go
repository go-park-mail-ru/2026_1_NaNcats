package repository

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/restaurant/internal/domain"
)

//go:generate mockgen -destination=mocks/restaurant_brand_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/services/restaurant/internal/repository RestaurantBrandRepository
type RestaurantBrandRepository interface {
	GetRestaurantBrandsList(ctx context.Context, limit, offset int) ([]domain.RestaurantBrand, error)
	GetByID(ctx context.Context, id int64) (domain.RestaurantBrand, error)
	GetRestaurantBrandsByIDs(ctx context.Context, ids []int64) ([]domain.RestaurantBrand, error)
	Create(ctx context.Context, b domain.RestaurantBrand, idempotencyKey string) (domain.RestaurantBrand, error)
	Delete(ctx context.Context, id int64) error
	Update(ctx context.Context, b domain.RestaurantBrand) (domain.RestaurantBrand, error)
}

type Category struct {
	ID    int64
	Name  string
	Emoji string
}

//go:generate mockgen -destination=mocks/extended_restaurant_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/services/restaurant/internal/repository ExtendedRestaurantRepository
type ExtendedRestaurantRepository interface {
	GetRestaurantBrandsByCategory(ctx context.Context, categoryID int64, limit, offset int) ([]domain.RestaurantBrand, error)
	GetRestaurantBrandsByCategoryName(ctx context.Context, categoryName string, limit, offset int) ([]domain.RestaurantBrand, error)
	SearchRestaurantBrands(ctx context.Context, query string, limit, offset int) ([]domain.RestaurantBrand, error)
	GetAllCategories(ctx context.Context) ([]Category, error)
}

type RestaurantBrandFullRepository interface {
	RestaurantBrandRepository
	ExtendedRestaurantRepository
}
