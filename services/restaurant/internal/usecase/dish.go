package usecase

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/restaurant/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/restaurant/internal/repository"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

//go:generate mockgen -destination=mocks/dish_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/services/restaurant/internal/usecase DishUseCase
//go:generate gowrap gen -i DishUseCase -t ../../../../shared/templates/tracing.tmpl -o dish_tracing_mw.go -v TracerName=restaurant-service
type DishUseCase interface {
	GetDishesByRestaurantBrandID(ctx context.Context, restaurantBrandID int64, limit, offset int) ([]domain.Dish, error)
	GetDishesByIDs(ctx context.Context, ids []int64) ([]domain.Dish, error)
}

type dishUseCase struct {
	dishRepo           repository.DishRepository
	defaultFoodLogoURL string
}

func NewDishUseCase(dr repository.DishRepository, dflurl string) DishUseCase {
	return &dishUseCase{
		dishRepo:           dr,
		defaultFoodLogoURL: dflurl,
	}
}

func (uc *dishUseCase) GetDishesByRestaurantBrandID(ctx context.Context, restaurantBrandID int64, limit, offset int) ([]domain.Dish, error) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.Int64("restaurant_brand.id", restaurantBrandID),
		attribute.Int("pagination.limit", limit),
		attribute.Int("pagination.offset", offset),
	)

	// Валидация входа
	if restaurantBrandID <= 0 {
		return nil, domain.ErrInvalidRestaurantBrandID
	}

	// Пагинация
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	dishes, err := uc.dishRepo.GetDishesByRestaurantBrandID(ctx, restaurantBrandID, limit, offset)
	if err != nil {
		return nil, err
	}

	span.SetAttributes(attribute.Int("dishes.returned_count", len(dishes)))

	for i, dish := range dishes {
		if dish.ImageURL == "" {
			dishes[i].ImageURL = uc.defaultFoodLogoURL
		}
	}

	return dishes, nil
}

func (uc *dishUseCase) GetDishesByIDs(ctx context.Context, ids []int64) ([]domain.Dish, error) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.Int("dishes.requested_count", len(ids)))

	if len(ids) == 0 {
		return nil, nil
	}

	dishes, err := uc.dishRepo.GetDishesByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	span.SetAttributes(attribute.Int("dishes.found_count", len(dishes)))

	for i := range dishes {
		if dishes[i].ImageURL == "" {
			dishes[i].ImageURL = uc.defaultFoodLogoURL
		}
	}

	return dishes, nil
}
