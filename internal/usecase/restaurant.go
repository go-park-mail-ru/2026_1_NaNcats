package usecase

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/repository"
)

//go:generate mockgen -destination=mocks/restaurant_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/internal/usecase RestaurantBrandUseCase
type RestaurantBrandUseCase interface {
	GetRestaurantBrandsList(ctx context.Context, limit, offset int) ([]domain.RestaurantBrand, error)
	GetRestaurantBrandByID(ctx context.Context, id int) (domain.RestaurantBrand, error)
}

type restaurantBrandUseCase struct {
	restaurantBrandRepo repository.RestaurantBrandRepository
}

func NewRestaurantBrandUseCase(rbr repository.RestaurantBrandRepository) RestaurantBrandUseCase {
	return &restaurantBrandUseCase{
		restaurantBrandRepo: rbr,
	}
}

func (rb *restaurantBrandUseCase) GetRestaurantBrandsList(ctx context.Context, limit, offset int) ([]domain.RestaurantBrand, error) {
	restaurantBrands, err := rb.restaurantBrandRepo.GetRestaurantBrandsList(ctx, limit, offset)
	if err != nil {
		return nil, err
	}
	return restaurantBrands, nil
}

func (rb *restaurantBrandUseCase) GetRestaurantBrandByID(ctx context.Context, id int) (domain.RestaurantBrand, error) {
	return rb.restaurantBrandRepo.GetByID(ctx, id)
}
