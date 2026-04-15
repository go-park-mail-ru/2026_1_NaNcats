package restaurant

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/repository"
)

//go:generate mockgen -destination=mocks/restaurant_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/internal/usecase/restaurant RestaurantBrandUseCase
type RestaurantBrandUseCase interface {
	GetRestaurantBrandsList(ctx context.Context, limit, offset int) ([]domain.RestaurantBrand, error)
	GetRestaurantBrandByID(ctx context.Context, id int) (domain.RestaurantBrand, error)
}

type restaurantBrandUseCase struct {
	restaurantBrandRepo      repository.RestaurantBrandRepository
	defaultRestaurantLogoURL string
}

func NewRestaurantBrandUseCase(rbr repository.RestaurantBrandRepository, drlurl string) RestaurantBrandUseCase {
	return &restaurantBrandUseCase{
		restaurantBrandRepo:      rbr,
		defaultRestaurantLogoURL: drlurl,
	}
}

func (rb *restaurantBrandUseCase) GetRestaurantBrandsList(ctx context.Context, limit, offset int) ([]domain.RestaurantBrand, error) {
	restaurantBrands, err := rb.restaurantBrandRepo.GetRestaurantBrandsList(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	for i, restaurantBrand := range restaurantBrands {
		if restaurantBrand.LogoURL == "" {
			restaurantBrands[i].LogoURL = rb.defaultRestaurantLogoURL
		}
	}

	return restaurantBrands, nil
}

func (rb *restaurantBrandUseCase) GetRestaurantBrandByID(ctx context.Context, id int) (domain.RestaurantBrand, error) {
	restaurantBrand, err := rb.restaurantBrandRepo.GetByID(ctx, id)
	if err != nil {
		return domain.RestaurantBrand{}, err
	}

	if restaurantBrand.LogoURL == "" {
		restaurantBrand.LogoURL = rb.defaultRestaurantLogoURL
	}

	return restaurantBrand, nil
}
