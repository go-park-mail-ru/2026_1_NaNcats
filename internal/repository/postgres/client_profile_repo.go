package postgres

import (
	"context"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

type clientProfileRepo struct {
	pool *pgxpool.Pool
}

func NewClientProfileRepo(pool *pgxpool.Pool) repository.ClientProfileRepository {
	return &clientProfileRepo{pool: pool}
}

func (r *clientProfileRepo) Create(ctx context.Context, accountID int) error {
	query := `INSERT INTO "client_profile" (account_id) VALUES ($1)`
	_, err := r.pool.Exec(ctx, query, accountID)
	return err
}

func (r *clientProfileRepo) GetByAccountID(ctx context.Context, accountID int) (domain.ClientProfile, error) {
	query := `SELECT account_id, bonus_balance, streak_count FROM "client_profile" WHERE account_id = $1`
	var p domain.ClientProfile
	err := r.pool.QueryRow(ctx, query, accountID).Scan(&p.AccountID, &p.BonusBalance, &p.StreakCount)
	if err != nil {
		return domain.ClientProfile{}, err
	}
	return p, nil
}
