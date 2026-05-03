package restaurant

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/restaurant/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/restaurant/internal/repository"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/postgres"
	"github.com/jackc/pgx/v5"
)

type restaurantBrandDB struct {
	ID             int64     `db:"id"`
	OwnerProfileID int64     `db:"owner_profile_id"`
	Name           string    `db:"name"`
	Description    *string   `db:"description"`
	PromotionTier  int       `db:"promotion_tier"`
	LogoURL        *string   `db:"logo_url"`
	CreatedAt      time.Time `db:"created_at"`
	UpdatedAt      time.Time `db:"updated_at"`
}

func (d restaurantBrandDB) toDomain() domain.RestaurantBrand {
	description := ""
	if d.Description != nil {
		description = *d.Description
	}
	logoURL := ""
	if d.LogoURL != nil {
		logoURL = *d.LogoURL
	}
	return domain.RestaurantBrand{
		ID:             d.ID,
		OwnerProfileID: d.OwnerProfileID,
		Name:           d.Name,
		Description:    description,
		PromotionTier:  d.PromotionTier,
		LogoURL:        logoURL,
		CreatedAt:      d.CreatedAt,
		UpdatedAt:      d.UpdatedAt,
	}
}

type restaurantBrandRepo struct {
	pool postgres.PgxPool
}

func NewRestaurantBrandRepo(pool postgres.PgxPool) repository.RestaurantBrandRepository {
	return &restaurantBrandRepo{
		pool: pool,
	}
}

func (r *restaurantBrandRepo) GetRestaurantBrandsList(ctx context.Context, limit, offset int) ([]domain.RestaurantBrand, error) {
	query := `
		SELECT id, owner_profile_id, name, description, promotion_tier, logo_url, created_at, updated_at
		FROM "restaurant_brand"
		ORDER BY promotion_tier DESC, id ASC
		LIMIT $1 OFFSET $2;
	`
	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("query restaurant brands list: %w", err)
	}
	defer rows.Close()

	dbRestaurantBrands, err := pgx.CollectRows(rows, pgx.RowToStructByName[restaurantBrandDB])
	if err != nil {
		return nil, fmt.Errorf("scan restaurant brands rows: %w", err)
	}

	domainRestaurantBrands := make([]domain.RestaurantBrand, 0, len(dbRestaurantBrands))
	for _, dbRestaurantBrand := range dbRestaurantBrands {
		domainRestaurantBrands = append(domainRestaurantBrands, dbRestaurantBrand.toDomain())
	}

	return domainRestaurantBrands, nil
}

func (r *restaurantBrandRepo) GetByID(ctx context.Context, id int64) (domain.RestaurantBrand, error) {
	query := `
		SELECT id, owner_profile_id, name, description, promotion_tier, logo_url, created_at, updated_at
		FROM "restaurant_brand"
		WHERE id = $1;
	`
	var rb restaurantBrandDB
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&rb.ID, &rb.OwnerProfileID, &rb.Name, &rb.Description,
		&rb.PromotionTier, &rb.LogoURL, &rb.CreatedAt, &rb.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.RestaurantBrand{}, domain.ErrRestaurantNotFound
		}
		return domain.RestaurantBrand{}, fmt.Errorf("get restaurant by id [%d]: %w", id, err)

	}
	return rb.toDomain(), nil
}

func (r *restaurantBrandRepo) GetRestaurantBrandsByIDs(ctx context.Context, ids []int64) ([]domain.RestaurantBrand, error) {
	query := `
		SELECT id, owner_profile_id, name, description, promotion_tier, logo_url, created_at, updated_at
		FROM "restaurant_brand"
		WHERE id = ANY($1);
	`

	rows, err := r.pool.Query(ctx, query, ids)
	if err != nil {
		return nil, fmt.Errorf("postgres query error: %w", err)
	}
	defer rows.Close()

	dbRestaurantBrands, err := pgx.CollectRows(rows, pgx.RowToStructByName[restaurantBrandDB])
	if err != nil {
		return nil, fmt.Errorf("mapping dishes: %w", err)
	}

	restaurantBrands := make([]domain.RestaurantBrand, 0, len(ids))
	for _, rb := range dbRestaurantBrands {
		restaurantBrands = append(restaurantBrands, rb.toDomain())
	}
	return restaurantBrands, nil
}

func (r *restaurantBrandRepo) Create(ctx context.Context, b domain.RestaurantBrand, idempotencyKey string) (domain.RestaurantBrand, error) {
	query := `
		INSERT INTO "restaurant_brand" (owner_profile_id, name, description, logo_url, idempotency_key)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (idempotency_key) DO UPDATE SET 
            idempotency_key = EXCLUDED.idempotency_key
		RETURNING id, owner_profile_id, name, description, promotion_tier, logo_url, created_at, updated_at;
	`
	var rb restaurantBrandDB
	err := r.pool.QueryRow(ctx, query, b.OwnerProfileID, b.Name, b.Description, b.LogoURL, idempotencyKey).Scan(
		&rb.ID, &rb.OwnerProfileID, &rb.Name, &rb.Description, &rb.PromotionTier, &rb.LogoURL, &rb.CreatedAt, &rb.UpdatedAt,
	)
	if err != nil {
		return domain.RestaurantBrand{}, err
	}
	return rb.toDomain(), nil
}

func (r *restaurantBrandRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM "restaurant_brand" WHERE id = $1`, id)
	return err
}

func (r *restaurantBrandRepo) Update(ctx context.Context, b domain.RestaurantBrand) (domain.RestaurantBrand, error) {
	query := `
		UPDATE "restaurant_brand" 
		SET name = $1, description = $2, logo_url = $3, promotion_tier = $4, updated_at = NOW()
		WHERE id = $5
	`
	_, err := r.pool.Exec(ctx, query, b.Name, b.Description, b.LogoURL, b.PromotionTier, b.ID)
	if err != nil {
		return domain.RestaurantBrand{}, err
	}

	return b, err
}
