package usecase

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/repository"
)

type RestaurantBrandUseCase interface {
	GetRestaurantBrandsList(ctx context.Context, limit, offset int) []domain.RestaurantBrand
}

type restaurantBrandUseCase struct {
	restaurantBrandRepo repository.RestaurantBrandRepository
}

func NewRestaurantBrandUseCase(rbr repository.RestaurantBrandRepository) RestaurantBrandUseCase {
	return &restaurantBrandUseCase{
		restaurantBrandRepo: rbr,
	}
}

func (rb *restaurantBrandUseCase) GetRestaurantBrandsList(ctx context.Context, limit, offset int) []domain.RestaurantBrand {
	return rb.restaurantBrandRepo.GetRestaurantBrandsList(ctx, limit, offset)
}
