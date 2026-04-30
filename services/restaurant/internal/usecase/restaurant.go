package usecase

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/restaurant/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/restaurant/internal/repository"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/errutil"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
)

//go:generate mockgen -destination=mocks/restaurant_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/services/restaurant/internal/usecase RestaurantBrandUseCase
//go:generate gowrap gen -i RestaurantBrandUseCase -t ../../../../shared/templates/tracing.tmpl -o restaurant_brand_tracing_mw.go -v TracerName=restaurant-service
type RestaurantBrandUseCase interface {
	GetRestaurantBrandsList(ctx context.Context, limit, offset int) ([]domain.RestaurantBrand, error)
	GetRestaurantBrandByID(ctx context.Context, id int64) (domain.RestaurantBrand, error)
	GetRestaurantBrandsByIDs(ctx context.Context, brandIDs []int64) ([]domain.RestaurantBrand, error)
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

func (rb *restaurantBrandUseCase) GetRestaurantBrandsByIDs(ctx context.Context, brandIDs []int64) ([]domain.RestaurantBrand, error) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.Int64Slice("brand.ids", brandIDs))

	brands, err := rb.restaurantBrandRepo.GetRestaurantBrandsByIDs(ctx, brandIDs)
	if err != nil {
		return nil, err
	}

	span.SetAttributes(attribute.Int("brands.count", len(brands)))

	for i, brand := range brands {
		if brand.LogoURL == "" {
			brands[i].LogoURL = rb.defaultRestaurantLogoURL
		}
	}

	return brands, nil
}

func (rb *restaurantBrandUseCase) GetRestaurantBrandsList(ctx context.Context, limit, offset int) ([]domain.RestaurantBrand, error) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.Int("pagination.limit", limit),
		attribute.Int("pagination.offset", offset),
	)

	restaurantBrands, err := rb.restaurantBrandRepo.GetRestaurantBrandsList(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	span.SetAttributes(attribute.Int("brands.count", len(restaurantBrands)))

	for i, restaurantBrand := range restaurantBrands {
		if restaurantBrand.LogoURL == "" {
			restaurantBrands[i].LogoURL = rb.defaultRestaurantLogoURL
		}
	}

	return restaurantBrands, nil
}

func (rb *restaurantBrandUseCase) GetRestaurantBrandByID(ctx context.Context, id int64) (domain.RestaurantBrand, error) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.Int64("brand.id", id))

	if id <= 0 {
		return domain.RestaurantBrand{}, errutil.New("INVALID_RESTAURANT_BRAND_ID", "invalid restaurant brand id", codes.InvalidArgument)
	}

	restaurantBrand, err := rb.restaurantBrandRepo.GetByID(ctx, id)
	if err != nil {
		return domain.RestaurantBrand{}, err
	}

	if restaurantBrand.LogoURL == "" {
		restaurantBrand.LogoURL = rb.defaultRestaurantLogoURL
	}

	return restaurantBrand, nil
}
