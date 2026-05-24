package usecase

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/restaurant/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/restaurant/internal/repository"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/errutil"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

//go:generate mockgen -destination=mocks/recommendations_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/services/restaurant/internal/usecase RecommendationsUseCase,OrderHistoryClient

// OrderHistoryClient — порт во внешний order-service. Restaurant-service сам
// не владеет историей заказов; вся межсервисная зависимость скрыта за этим
// интерфейсом, что сохраняет чистую архитектуру слоя usecase.
type OrderHistoryClient interface {
	GetUserPaidBrands(ctx context.Context, userID int64) ([]int64, error)
	GetTrendingBrands(ctx context.Context, windowDays, limit int32) ([]int64, error)
}

type RecommendationsUseCase interface {
	GetRecommendations(ctx context.Context, userID int64, limit int) ([]domain.RestaurantBrand, error)
}

type recommendationsUseCase struct {
	repo                     repository.RestaurantBrandRepository
	orderClient              OrderHistoryClient
	defaultRestaurantLogoURL string
	logger                   logger.Logger
}

func NewRecommendationsUseCase(
	repo repository.RestaurantBrandRepository,
	orderClient OrderHistoryClient,
	defaultLogoURL string,
	l logger.Logger,
) RecommendationsUseCase {
	return &recommendationsUseCase{
		repo:                     repo,
		orderClient:              orderClient,
		defaultRestaurantLogoURL: defaultLogoURL,
		logger:                   l,
	}
}

const (
	trendingWindowDays = 7
	trendingPoolFactor = 3
)

// GetRecommendations подбирает бренды для пользователя:
//   - есть история paid-заказов → берёт бренды с пересекающимися категориями,
//     исключая уже посещённые. Если пересечений мало — добивает выдачу
//     трендом за 7 дней.
//   - истории нет (или userID == 0) → возвращает топ-N тренда за 7 дней;
//     при пустом тренде fallback на «холодный старт» по promotion_tier.
func (u *recommendationsUseCase) GetRecommendations(ctx context.Context, userID int64, limit int) ([]domain.RestaurantBrand, error) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.Int64("user.id", userID),
		attribute.Int("recommendations.limit", limit),
	)

	if limit <= 0 {
		limit = 4
	}

	var seedBrands []int64
	if userID > 0 {
		ub, err := u.orderClient.GetUserPaidBrands(ctx, userID)
		if err != nil {
			u.logger.WithContext(ctx).Warn("recommendations: order client failed; treat as cold start", logger.Err(err))
		} else {
			seedBrands = ub
		}
	}

	span.SetAttributes(attribute.Int("recommendations.seed_count", len(seedBrands)))

	results, err := u.repo.RecommendByCategorySimilarity(ctx, seedBrands, seedBrands, limit)
	if err != nil {
		return nil, errutil.Internal("failed to query recommendations from repo", err)
	}

	if len(results) >= limit {
		u.applyDefaultLogos(results)
		span.SetAttributes(attribute.Int("recommendations.returned", len(results)))
		return results[:limit], nil
	}

	// Добиваем трендом, если эвристика по категориям не насыщает limit.
	exclude := append([]int64{}, seedBrands...)
	for _, b := range results {
		exclude = append(exclude, b.ID)
	}

	trending, err := u.orderClient.GetTrendingBrands(ctx, trendingWindowDays, int32(limit*trendingPoolFactor))
	if err != nil {
		u.logger.WithContext(ctx).Warn("recommendations: trending fetch failed", logger.Err(err))
		u.applyDefaultLogos(results)
		return results, nil
	}

	if len(trending) > 0 {
		extras, err := u.repo.GetRestaurantBrandsByIDs(ctx, trending)
		if err == nil {
			byID := make(map[int64]domain.RestaurantBrand, len(extras))
			for _, b := range extras {
				byID[b.ID] = b
			}
			excludeSet := make(map[int64]struct{}, len(exclude))
			for _, id := range exclude {
				excludeSet[id] = struct{}{}
			}
			for _, id := range trending {
				if len(results) >= limit {
					break
				}
				if _, skip := excludeSet[id]; skip {
					continue
				}
				if b, ok := byID[id]; ok {
					results = append(results, b)
					excludeSet[id] = struct{}{}
				}
			}
		}
	}

	if len(results) < limit {
		// Финальный fallback: «холодный старт» по promotion_tier, исключая
		// всё, что уже выдали и историю пользователя.
		exclude = exclude[:0]
		exclude = append(exclude, seedBrands...)
		for _, b := range results {
			exclude = append(exclude, b.ID)
		}
		cold, err := u.repo.RecommendByCategorySimilarity(ctx, nil, exclude, limit-len(results))
		if err == nil {
			results = append(results, cold...)
		}
	}

	u.applyDefaultLogos(results)
	span.SetAttributes(attribute.Int("recommendations.returned", len(results)))
	return results, nil
}

func (u *recommendationsUseCase) applyDefaultLogos(brands []domain.RestaurantBrand) {
	if u.defaultRestaurantLogoURL == "" {
		return
	}
	for i := range brands {
		if brands[i].LogoURL == "" {
			brands[i].LogoURL = u.defaultRestaurantLogoURL
		}
	}
}
