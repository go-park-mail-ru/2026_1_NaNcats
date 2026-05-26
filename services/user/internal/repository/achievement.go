package repository

import (
	"context"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/user/internal/domain"
)

//go:generate mockgen -destination=mocks/achievement_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/services/user/internal/repository AchievementRepository

type IncrementPaidOrderResult struct {
	NewPaidOrdersCount int
	NewStreakCount     int
}

type AchievementRepository interface {
	ListAll(ctx context.Context) ([]domain.Achievement, error)
	GetByCode(ctx context.Context, code string) (domain.Achievement, error)
	ListForUser(ctx context.Context, accountID int64) ([]domain.UserAchievement, error)
	Award(ctx context.Context, accountID, achievementID int64) error

	IncrementPaidOrders(ctx context.Context, accountID int64, paidAt time.Time) (IncrementPaidOrderResult, error)
	RegisterRestaurantForAccount(ctx context.Context, accountID, restaurantID int64, at time.Time) (distinctRestaurants int, err error)
}
