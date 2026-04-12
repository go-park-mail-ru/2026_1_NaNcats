package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/repository"
)

type restaurantBrandDB struct {
	ID             int       `db:"id"`
	OwnerProfileID int       `db:"owner_profile_id"`
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
	pool PgxPool
}

func NewRestaurantBrandRepo(pool PgxPool) repository.RestaurantBrandRepository {
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
		return nil, err
	}
	defer rows.Close()

	dbRestaurantBrands, err := pgx.CollectRows(rows, pgx.RowToStructByName[restaurantBrandDB])
	if err != nil {
		return nil, err
	}

	domainRestaurantBrands := make([]domain.RestaurantBrand, 0, len(dbRestaurantBrands))
	for _, dbRestaurantBrand := range dbRestaurantBrands {
		domainRestaurantBrands = append(domainRestaurantBrands, dbRestaurantBrand.toDomain())
	}

	return domainRestaurantBrands, nil
}

func (r *restaurantBrandRepo) GetByID(ctx context.Context, id int) (domain.RestaurantBrand, error) {
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
		return domain.RestaurantBrand{}, err
	}
	return rb.toDomain(), nil
}
