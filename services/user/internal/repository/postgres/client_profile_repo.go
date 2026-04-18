package user

import (
	"context"
	"errors"

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
		SELECT account_id, bonus_balance, streak_count
		FROM "client_profile"
		WHERE account_id = $1
	`

	var p domain.ClientProfile
	err := r.pool.QueryRow(ctx, query, accountID).Scan(&p.AccountID, &p.BonusBalance, &p.StreakCount)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ClientProfile{}, domain.ErrUserNotFound
		}
		return domain.ClientProfile{}, err
	}

	return p, nil
}
