package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/user/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/user/internal/repository"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/postgres"
	"github.com/jackc/pgx/v5"
)

type clientProfileRepo struct {
	pool postgres.PgxPool
}

func NewClientProfileRepo(pool postgres.PgxPool) repository.ClientProfileRepository {
	return &clientProfileRepo{pool: pool}
}

func (r *clientProfileRepo) Create(ctx context.Context, accountID int64, idempotencyKey string) error {
	query := `
		INSERT INTO "client_profile" (account_id, idempotency_key)
		VALUES ($1, $2)
		ON CONFLICT (idempotency_key) DO UPDATE 
		SET idempotency_key = EXCLUDED.idempotency_key
	`
	_, err := r.pool.Exec(ctx, query, accountID, idempotencyKey)
	return err
}

func (r *clientProfileRepo) GetByAccountID(ctx context.Context, accountID int64) (domain.ClientProfile, error) {
	query := `
		SELECT account_id, bonus_balance, streak_count, last_order_date, streak_freeze_active
		FROM "client_profile"
		WHERE account_id = $1
	`

	var p domain.ClientProfile
	err := r.pool.QueryRow(ctx, query, accountID).Scan(
		&p.AccountID,
		&p.BonusBalance,
		&p.StreakCount,
		&p.LastOrderDate,
		&p.StreakFreezeActive,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ClientProfile{}, domain.ErrUserNotFound
		}
		return domain.ClientProfile{}, err
	}

	return p, nil
}

func (r *clientProfileRepo) UpdateStreakFreeze(ctx context.Context, accountID int64, active bool) error {
	var query string
	if active {
		query = `UPDATE "client_profile" SET streak_freeze_active = true, updated_at = NOW() WHERE account_id = $1`
	} else {
		query = `UPDATE "client_profile" SET streak_freeze_active = false, last_order_date = NOW(), updated_at = NOW() WHERE account_id = $1`
	}

	tag, err := r.pool.Exec(ctx, query, accountID)
	if err != nil {
		return fmt.Errorf("failed to update streak freeze: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

func (r *clientProfileRepo) IncrementStreak(ctx context.Context, accountID int64) error {
	query := `
		UPDATE "client_profile" 
		SET streak_count = streak_count + 1, last_order_date = NOW(), updated_at = NOW() 
		WHERE account_id = $1
	`
	tag, err := r.pool.Exec(ctx, query, accountID)
	if err != nil {
		return fmt.Errorf("failed to increment streak: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

func (r *clientProfileRepo) ResetStreak(ctx context.Context, accountID int64) error {
	query := `
		UPDATE "client_profile" 
		SET streak_count = 0, last_order_date = NULL, updated_at = NOW() 
		WHERE account_id = $1
	`
	tag, err := r.pool.Exec(ctx, query, accountID)
	if err != nil {
		return fmt.Errorf("failed to reset streak in db: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}
