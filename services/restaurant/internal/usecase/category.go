package usecase

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/restaurant/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/restaurant/internal/repository"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

//go:generate mockgen -destination=mocks/category_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/services/restaurant/internal/usecase CategoryUseCase
type CategoryUseCase interface {
	GetAllCategories(ctx context.Context) ([]domain.Category, error)
}

type categoryUseCase struct {
	repo repository.RestaurantBrandRepository
}

func NewCategoryUseCase(repo repository.RestaurantBrandRepository) CategoryUseCase {
	return &categoryUseCase{
		repo: repo,
	}
}

func (uc *categoryUseCase) GetAllCategories(ctx context.Context) ([]domain.Category, error) {
	span := trace.SpanFromContext(ctx)

	categories, err := uc.repo.GetAllCategories(ctx)
	if err != nil {
		return nil, err
	}

	span.SetAttributes(attribute.Int("categories.count", len(categories)))
	return categories, nil
}
