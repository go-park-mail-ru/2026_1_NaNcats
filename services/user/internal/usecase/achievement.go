package usecase

import (
	"context"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/user/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/user/internal/repository"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/errutil"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

//go:generate mockgen -destination=mocks/achievement_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/services/user/internal/usecase AchievementUseCase

type AchievementUseCase interface {
	ListAll(ctx context.Context) ([]domain.Achievement, error)
	ListForUser(ctx context.Context, accountID int64) ([]domain.UserAchievement, error)
	OnOrderPaid(ctx context.Context, accountID, restaurantID int64, paidAt time.Time) error
	OnWheelSpin(ctx context.Context, accountID int64) error
}

type achievementUseCase struct {
	repo   repository.AchievementRepository
	logger logger.Logger
}

func NewAchievementUseCase(r repository.AchievementRepository, l logger.Logger) AchievementUseCase {
	return &achievementUseCase{repo: r, logger: l}
}

func (u *achievementUseCase) ListAll(ctx context.Context) ([]domain.Achievement, error) {
	items, err := u.repo.ListAll(ctx)
	if err != nil {
		return nil, errutil.Internal("failed to list achievements", err)
	}
	return items, nil
}

func (u *achievementUseCase) ListForUser(ctx context.Context, accountID int64) ([]domain.UserAchievement, error) {
	items, err := u.repo.ListForUser(ctx, accountID)
	if err != nil {
		return nil, errutil.Internal("failed to list user achievements", err)
	}
	return items, nil
}

func (u *achievementUseCase) OnOrderPaid(ctx context.Context, accountID, restaurantID int64, paidAt time.Time) error {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.Int64("user.id", accountID),
		attribute.Int64("restaurant.id", restaurantID),
	)

	bump, err := u.repo.IncrementPaidOrders(ctx, accountID, paidAt)
	if err != nil {
		return errutil.Internal("failed to bump paid orders / streak", err)
	}
	span.SetAttributes(
		attribute.Int("user.paid_orders_count", bump.NewPaidOrdersCount),
		attribute.Int("user.streak_count", bump.NewStreakCount),
	)

	distinct, err := u.repo.RegisterRestaurantForAccount(ctx, accountID, restaurantID, paidAt)
	if err != nil {
		return errutil.Internal("failed to register restaurant for account", err)
	}
	span.SetAttributes(attribute.Int("user.distinct_restaurants", distinct))

	u.maybeAward(ctx, accountID, domain.AchievementCodeFirstOrder, bump.NewPaidOrdersCount >= 1)
	u.maybeAward(ctx, accountID, domain.AchievementCodeFiveOrders, bump.NewPaidOrdersCount >= 5)
	u.maybeAward(ctx, accountID, domain.AchievementCodeGourmandThree, distinct >= 3)
	u.maybeAward(ctx, accountID, domain.AchievementCodeStreakSixth, bump.NewStreakCount >= 6)

	return nil
}

func (u *achievementUseCase) OnWheelSpin(ctx context.Context, accountID int64) error {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.Int64("user.id", accountID))

	u.maybeAward(ctx, accountID, "first_spin", true)
	return nil
}

func (u *achievementUseCase) maybeAward(ctx context.Context, accountID int64, code string, ok bool) {
	if !ok {
		return
	}
	ach, err := u.repo.GetByCode(ctx, code)
	if err != nil {
		u.logger.WithContext(ctx).Warn("achievement lookup failed",
			logger.String("code", code), logger.Err(err))
		return
	}
	if err := u.repo.Award(ctx, accountID, ach.ID); err != nil {
		u.logger.WithContext(ctx).Warn("achievement award failed",
			logger.String("code", code), logger.Err(err))
	}
}
