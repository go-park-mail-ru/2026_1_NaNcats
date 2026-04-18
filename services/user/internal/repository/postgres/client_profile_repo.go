package user

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/repository"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/repository/postgres"
)

type clientProfileRepo struct {
	pool postgres.PgxPool
}

func NewClientProfileRepo(pool postgres.PgxPool) repository.ClientProfileRepository {
	return &clientProfileRepo{pool: pool}
}

func (r *clientProfileRepo) Create(ctx context.Context, accountID int64) error {
	query := `INSERT INTO "client_profile" (account_id) VALUES ($1)`
	_, err := r.pool.Exec(ctx, query, accountID)
	return err
}

func (r *clientProfileRepo) GetByAccountID(ctx context.Context, accountID int64) (domain.ClientProfile, error) {
	query := `SELECT account_id, bonus_balance, streak_count FROM "client_profile" WHERE account_id = $1`
	var p domain.ClientProfile
	err := r.pool.QueryRow(ctx, query, accountID).Scan(&p.AccountID, &p.BonusBalance, &p.StreakCount)
	if err != nil {
		return domain.ClientProfile{}, err
	}
	return p, nil
}
