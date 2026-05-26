package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/user/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/user/internal/repository"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/postgres"
	"github.com/jackc/pgx/v5"
)

type achievementRepo struct {
	pool postgres.PgxPool
}

func NewAchievementRepo(pool postgres.PgxPool) repository.AchievementRepository {
	return &achievementRepo{pool: pool}
}

func (r *achievementRepo) ListAll(ctx context.Context) ([]domain.Achievement, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, code, title, description, icon, sort_order
		FROM "achievement"
		ORDER BY sort_order, id
	`)
	if err != nil {
		return nil, fmt.Errorf("list achievements: %w", err)
	}
	defer rows.Close()

	out := make([]domain.Achievement, 0, 8)
	for rows.Next() {
		var a domain.Achievement
		if err := rows.Scan(&a.ID, &a.Code, &a.Title, &a.Description, &a.Icon, &a.SortOrder); err != nil {
			return nil, fmt.Errorf("scan achievement: %w", err)
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (r *achievementRepo) GetByCode(ctx context.Context, code string) (domain.Achievement, error) {
	var a domain.Achievement
	err := r.pool.QueryRow(ctx, `
		SELECT id, code, title, description, icon, sort_order
		FROM "achievement"
		WHERE code = $1
	`, code).Scan(&a.ID, &a.Code, &a.Title, &a.Description, &a.Icon, &a.SortOrder)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Achievement{}, domain.ErrAchievementNotFound
		}
		return domain.Achievement{}, fmt.Errorf("get achievement by code: %w", err)
	}
	return a, nil
}

func (r *achievementRepo) ListForUser(ctx context.Context, accountID int64) ([]domain.UserAchievement, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT achievement_id, awarded_at
		FROM "account_achievement"
		WHERE account_id = $1
		ORDER BY awarded_at
	`, accountID)
	if err != nil {
		return nil, fmt.Errorf("list user achievements: %w", err)
	}
	defer rows.Close()

	out := make([]domain.UserAchievement, 0, 8)
	for rows.Next() {
		var ua domain.UserAchievement
		if err := rows.Scan(&ua.AchievementID, &ua.AwardedAt); err != nil {
			return nil, fmt.Errorf("scan user achievement: %w", err)
		}
		out = append(out, ua)
	}
	return out, rows.Err()
}

func (r *achievementRepo) Award(ctx context.Context, accountID, achievementID int64) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO "account_achievement" (account_id, achievement_id)
		VALUES ($1, $2)
		ON CONFLICT (account_id, achievement_id) DO NOTHING
	`, accountID, achievementID)
	if err != nil {
		return fmt.Errorf("award achievement: %w", err)
	}
	return nil
}

// IncrementPaidOrders инкрементит счётчик paid-заказов и обновляет стрик по
// правилу «календарная неделя пн–вс»: если paidAt в той же ISO-неделе, что и
// last_order_date — стрик не меняется. Если ровно на одну ISO-неделю позже —
// +1. Иначе — сбрасывается в 1.
func (r *achievementRepo) IncrementPaidOrders(ctx context.Context, accountID int64, paidAt time.Time) (repository.IncrementPaidOrderResult, error) {
	query := `
		WITH cur AS (
			SELECT streak_count, last_order_date, streak_freeze_active
			FROM "client_profile"
			WHERE account_id = $1
			FOR UPDATE
		),
		calc AS (
			SELECT
				CASE
					WHEN last_order_date IS NULL THEN 1
					WHEN date_trunc('week', $2::timestamptz) = date_trunc('week', last_order_date) THEN streak_count
					WHEN date_trunc('week', $2::timestamptz) = date_trunc('week', last_order_date) + INTERVAL '1 week' THEN streak_count + 1
					WHEN streak_freeze_active = true THEN streak_count + 1
					ELSE 1
				END AS new_streak
			FROM cur
		)
		UPDATE "client_profile" AS cp
		SET paid_orders_count = cp.paid_orders_count + 1,
			streak_count = calc.new_streak,
			last_order_date = $2,
			streak_freeze_active = CASE 
				WHEN cp.last_order_date IS NOT NULL AND date_trunc('week', $2::timestamptz) > date_trunc('week', cp.last_order_date) + INTERVAL '1 week' THEN false 
				ELSE cp.streak_freeze_active 
			END
		FROM calc
		WHERE cp.account_id = $1
		RETURNING cp.paid_orders_count, cp.streak_count
	`

	var res repository.IncrementPaidOrderResult
	err := r.pool.QueryRow(ctx, query, accountID, paidAt).Scan(&res.NewPaidOrdersCount, &res.NewStreakCount)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return repository.IncrementPaidOrderResult{}, domain.ErrUserNotFound
		}
		return repository.IncrementPaidOrderResult{}, fmt.Errorf("increment paid orders: %w", err)
	}
	return res, nil
}

func (r *achievementRepo) RegisterRestaurantForAccount(ctx context.Context, accountID, restaurantID int64, at time.Time) (int, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	_, err = tx.Exec(ctx, `
		INSERT INTO "account_restaurant" (account_id, restaurant_id, first_order_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (account_id, restaurant_id) DO NOTHING
	`, accountID, restaurantID, at)
	if err != nil {
		return 0, fmt.Errorf("register restaurant: %w", err)
	}

	var count int
	err = tx.QueryRow(ctx, `
		SELECT COUNT(*) FROM "account_restaurant" WHERE account_id = $1
	`, accountID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count restaurants: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit tx: %w", err)
	}
	return count, nil
}
