package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/order/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/order/internal/repository"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/postgres"
	"github.com/jackc/pgx/v5"
)

type promoRepo struct {
	pool postgres.PgxPool
}

func NewPromoRepo(pool postgres.PgxPool) repository.PromoRepository {
	return &promoRepo{pool: pool}
}

// promoViewColumns - набор колонок для карточки промокода (без служебных полей).
const promoViewColumns = `id, code, title, discount_percent, discount_amount,
	min_order_amount, restaurant_brand_id, expires_at`

func scanPromoView(row pgx.Row) (domain.Promocode, error) {
	var p domain.Promocode
	err := row.Scan(&p.ID, &p.Code, &p.Title, &p.DiscountPercent, &p.DiscountAmount,
		&p.MinOrderAmount, &p.RestaurantBrandID, &p.ExpiresAt)
	return p, err
}

func collectPromoViews(rows pgx.Rows) ([]domain.Promocode, error) {
	defer rows.Close()

	promos := make([]domain.Promocode, 0)
	for rows.Next() {
		p, err := scanPromoView(rows)
		if err != nil {
			return nil, fmt.Errorf("scan promocode: %w", err)
		}
		promos = append(promos, p)
	}
	return promos, rows.Err()
}

func (r *promoRepo) GetUserPromocodes(ctx context.Context, userID int64) ([]domain.Promocode, error) {
	query := `
		SELECT ` + promoViewColumns + `
		FROM "user_promocode" up
		JOIN "promocode" p ON p.id = up.promocode_id
		WHERE up.user_id = $1
		  AND p.expires_at > NOW()
		  AND (p.max_uses IS NULL OR p.current_uses < p.max_uses)
		  AND NOT EXISTS (
		      SELECT 1 FROM "promocode_usage" pu
		      WHERE pu.promocode_id = p.id AND pu.user_id = $1
		  )
		ORDER BY up.added_at DESC
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("query user promocodes: %w", err)
	}
	return collectPromoViews(rows)
}

func (r *promoRepo) GetRestaurantPromocodes(ctx context.Context, brandID int64) ([]domain.Promocode, error) {
	query := `
		SELECT ` + promoViewColumns + `
		FROM "promocode" p
		WHERE p.restaurant_brand_id = $1
		  AND p.expires_at > NOW()
		ORDER BY p.created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, brandID)
	if err != nil {
		return nil, fmt.Errorf("query restaurant promocodes: %w", err)
	}
	return collectPromoViews(rows)
}

func (r *promoRepo) GetPromocodeByCode(ctx context.Context, code string) (domain.Promocode, error) {
	query := `
		SELECT id, code, title, discount_percent, discount_amount, max_uses,
		       current_uses, min_order_amount, user_id, restaurant_brand_id,
		       is_global, created_at, expires_at
		FROM "promocode"
		WHERE code = $1
	`
	var p domain.Promocode
	err := r.pool.QueryRow(ctx, query, code).Scan(
		&p.ID, &p.Code, &p.Title, &p.DiscountPercent, &p.DiscountAmount, &p.MaxUses,
		&p.CurrentUses, &p.MinOrderAmount, &p.UserID, &p.RestaurantBrandID,
		&p.IsGlobal, &p.CreatedAt, &p.ExpiresAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Promocode{}, repository.ErrPromocodeNotFound
		}
		return domain.Promocode{}, fmt.Errorf("query promocode by code: %w", err)
	}
	return p, nil
}

func (r *promoRepo) BindPromocodeToUser(ctx context.Context, userID, promoID int64) (bool, error) {
	tag, err := r.pool.Exec(ctx, `
		INSERT INTO "user_promocode" (user_id, promocode_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`, userID, promoID)
	if err != nil {
		return false, fmt.Errorf("bind promocode to user: %w", err)
	}
	return tag.RowsAffected() > 0, nil
}

func (r *promoRepo) CountUserPromocodeUsage(ctx context.Context, promoID, userID int64) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM "promocode_usage"
		WHERE promocode_id = $1 AND user_id = $2
	`, promoID, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count promocode usage: %w", err)
	}
	return count, nil
}

func (r *promoRepo) RecordPromocodeUsage(ctx context.Context, promoID, userID int64, orderID *int64) error {
	tag, err := r.pool.Exec(ctx, `
		INSERT INTO "promocode_usage" (promocode_id, user_id, order_id)
		VALUES ($1, $2, $3)
		ON CONFLICT (promocode_id, user_id) DO NOTHING
	`, promoID, userID, orderID)
	if err != nil {
		return fmt.Errorf("record promocode usage: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return nil
	}

	if _, err := r.pool.Exec(ctx,
		`UPDATE "promocode" SET current_uses = current_uses + 1 WHERE id = $1`,
		promoID); err != nil {
		return fmt.Errorf("increment promocode uses: %w", err)
	}
	return nil
}

func (r *promoRepo) ResolveOrderInternalID(ctx context.Context, orderPublicID string) (*int64, error) {
	var id int64
	err := r.pool.QueryRow(ctx,
		`SELECT id FROM "order" WHERE public_id = $1`, orderPublicID).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("resolve order internal id: %w", err)
	}
	return &id, nil
}

func (r *promoRepo) CreatePromocode(ctx context.Context, p domain.Promocode) (domain.Promocode, error) {
	query := `
		INSERT INTO "promocode" (
			code, title, discount_percent, discount_amount, max_uses, 
			min_order_amount, user_id, restaurant_brand_id, is_global, expires_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at;
	`
	err := r.pool.QueryRow(ctx, query,
		p.Code, p.Title, p.DiscountPercent, p.DiscountAmount, p.MaxUses,
		p.MinOrderAmount, p.UserID, p.RestaurantBrandID, p.IsGlobal, p.ExpiresAt,
	).Scan(&p.ID, &p.CreatedAt)
	if err != nil {
		return domain.Promocode{}, fmt.Errorf("failed to insert promocode: %w", err)
	}
	return p, nil
}
