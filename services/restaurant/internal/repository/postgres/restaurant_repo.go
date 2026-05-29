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
	BannerURL      *string   `db:"banner_url"`
	CreatedAt      time.Time `db:"created_at"`
	UpdatedAt      time.Time `db:"updated_at"`
}

type categoryDB struct {
	ID        int64     `db:"id"`
	Name      string    `db:"name"`
	Emoji     string    `db:"emoji"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
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
	bannerURL := ""
	if d.BannerURL != nil {
		bannerURL = *d.BannerURL
	}
	return domain.RestaurantBrand{
		ID:             d.ID,
		OwnerProfileID: d.OwnerProfileID,
		Name:           d.Name,
		Description:    description,
		PromotionTier:  d.PromotionTier,
		LogoURL:        logoURL,
		BannerURL:      bannerURL,
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
		SELECT id, owner_profile_id, name, description, promotion_tier, logo_url, banner_url, created_at, updated_at
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
		SELECT id, owner_profile_id, name, description, promotion_tier, logo_url, banner_url, created_at, updated_at
		FROM "restaurant_brand"
		WHERE id = $1;
	`
	var rb restaurantBrandDB
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&rb.ID, &rb.OwnerProfileID, &rb.Name, &rb.Description,
		&rb.PromotionTier, &rb.LogoURL, &rb.BannerURL, &rb.CreatedAt, &rb.UpdatedAt,
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
		SELECT id, owner_profile_id, name, description, promotion_tier, logo_url, banner_url, created_at, updated_at
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
		return nil, fmt.Errorf("mapping restaurant brands: %w", err)
	}

	restaurantBrands := make([]domain.RestaurantBrand, 0, len(dbRestaurantBrands))
	for _, rb := range dbRestaurantBrands {
		restaurantBrands = append(restaurantBrands, rb.toDomain())
	}
	return restaurantBrands, nil
}

func (r *restaurantBrandRepo) Create(ctx context.Context, b domain.RestaurantBrand, idempotencyKey string) (domain.RestaurantBrand, error) {
	query := `
		INSERT INTO "restaurant_brand" (owner_profile_id, name, description, promotion_tier, logo_url, idempotency_key)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (idempotency_key) DO UPDATE SET updated_at = NOW()
		RETURNING id, owner_profile_id, name, description, promotion_tier, logo_url, banner_url, created_at, updated_at;
	`
	var rb restaurantBrandDB
	err := r.pool.QueryRow(ctx, query, b.OwnerProfileID, b.Name, b.Description, b.PromotionTier, b.LogoURL, idempotencyKey).Scan(
		&rb.ID, &rb.OwnerProfileID, &rb.Name, &rb.Description, &rb.PromotionTier, &rb.LogoURL, &rb.BannerURL, &rb.CreatedAt, &rb.UpdatedAt,
	)
	if err != nil {
		return domain.RestaurantBrand{}, fmt.Errorf("create restaurant brand: %w", err)
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
		RETURNING id, owner_profile_id, name, description, promotion_tier, logo_url, banner_url, created_at, updated_at;
	`
	var rb restaurantBrandDB
	err := r.pool.QueryRow(ctx, query, b.Name, b.Description, b.LogoURL, b.PromotionTier, b.ID).Scan(
		&rb.ID, &rb.OwnerProfileID, &rb.Name, &rb.Description, &rb.PromotionTier, &rb.LogoURL, &rb.BannerURL, &rb.CreatedAt, &rb.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.RestaurantBrand{}, domain.ErrRestaurantNotFound
		}
		return domain.RestaurantBrand{}, fmt.Errorf("update restaurant brand: %w", err)
	}

	return rb.toDomain(), nil
}

func (r *restaurantBrandRepo) GetRestaurantBrandsByCategory(ctx context.Context, categoryID int64, limit, offset int) ([]domain.RestaurantBrand, error) {
	query := `
		SELECT rb.id, rb.owner_profile_id, rb.name, rb.description, rb.promotion_tier, rb.logo_url, rb.banner_url, rb.created_at, rb.updated_at
		FROM "restaurant_brand" rb
		JOIN "restaurant_brand_category" rbc ON rbc.restaurant_brand_id = rb.id
		WHERE rbc.category_id = $1
		ORDER BY rb.promotion_tier DESC, rb.id ASC
		LIMIT $2 OFFSET $3;
	`
	rows, err := r.pool.Query(ctx, query, categoryID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("query brands by category: %w", err)
	}
	defer rows.Close()

	dbBrands, err := pgx.CollectRows(rows, pgx.RowToStructByName[restaurantBrandDB])
	if err != nil {
		return nil, fmt.Errorf("scan brands by category: %w", err)
	}

	brands := make([]domain.RestaurantBrand, 0, len(dbBrands))
	for _, b := range dbBrands {
		brands = append(brands, b.toDomain())
	}
	return brands, nil
}

func (r *restaurantBrandRepo) GetRestaurantBrandsByCategoryName(ctx context.Context, categoryName string, limit, offset int) ([]domain.RestaurantBrand, error) {
	query := `
		SELECT rb.id, rb.owner_profile_id, rb.name, rb.description, rb.promotion_tier, rb.logo_url, rb.banner_url, rb.created_at, rb.updated_at
		FROM "restaurant_brand" rb
		JOIN "restaurant_brand_category" rbc ON rbc.restaurant_brand_id = rb.id
		JOIN "category" c ON c.id = rbc.category_id
		WHERE LOWER(c.name) = LOWER($1)
		ORDER BY rb.promotion_tier DESC, rb.id ASC
		LIMIT $2 OFFSET $3;
	`
	rows, err := r.pool.Query(ctx, query, categoryName, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("query brands by category name: %w", err)
	}
	defer rows.Close()

	dbBrands, err := pgx.CollectRows(rows, pgx.RowToStructByName[restaurantBrandDB])
	if err != nil {
		return nil, fmt.Errorf("scan brands by category name: %w", err)
	}

	brands := make([]domain.RestaurantBrand, 0, len(dbBrands))
	for _, b := range dbBrands {
		brands = append(brands, b.toDomain())
	}
	return brands, nil
}

func (r *restaurantBrandRepo) SearchRestaurantBrands(ctx context.Context, query string, limit, offset int) ([]domain.RestaurantBrand, error) {
	q := `
		SELECT id, owner_profile_id, name, description, promotion_tier, logo_url, banner_url, created_at, updated_at
		FROM "restaurant_brand"
		WHERE name ILIKE $1 OR description ILIKE $1
		ORDER BY promotion_tier DESC, id ASC
		LIMIT $2 OFFSET $3;
	`
	pattern := "%" + query + "%"
	rows, err := r.pool.Query(ctx, q, pattern, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("search brands: %w", err)
	}
	defer rows.Close()

	dbBrands, err := pgx.CollectRows(rows, pgx.RowToStructByName[restaurantBrandDB])
	if err != nil {
		return nil, fmt.Errorf("scan search brands: %w", err)
	}

	brands := make([]domain.RestaurantBrand, 0, len(dbBrands))
	for _, b := range dbBrands {
		brands = append(brands, b.toDomain())
	}
	return brands, nil
}

func (r *restaurantBrandRepo) GetAllCategories(ctx context.Context) ([]domain.Category, error) {
	query := `SELECT id, name, COALESCE(emoji, '') as emoji, created_at, updated_at FROM "category" ORDER BY id ASC;`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("get categories: %w", err)
	}
	defer rows.Close()

	dbCats, err := pgx.CollectRows(rows, pgx.RowToStructByName[categoryDB])
	if err != nil {
		return nil, fmt.Errorf("scan categories: %w", err)
	}

	cats := make([]domain.Category, 0, len(dbCats))
	for _, c := range dbCats {
		cats = append(cats, domain.Category{
			ID:        c.ID,
			Name:      c.Name,
			Emoji:     c.Emoji,
			CreatedAt: c.CreatedAt,
			UpdatedAt: c.UpdatedAt,
		})
	}
	return cats, nil
}

// RecommendByCategorySimilarity ранжирует бренды по числу общих категорий
// с seedBrandIDs. При пустом seed выбирает топ по promotion_tier как
// «холодный старт». excludeBrandIDs всегда фильтруются.
func (r *restaurantBrandRepo) RecommendByCategorySimilarity(ctx context.Context, seedBrandIDs, excludeBrandIDs []int64, limit int) ([]domain.RestaurantBrand, error) {
	if limit <= 0 {
		limit = 8
	}
	if excludeBrandIDs == nil {
		excludeBrandIDs = []int64{}
	}

	if len(seedBrandIDs) == 0 {
		rows, err := r.pool.Query(ctx, `
			SELECT id, owner_profile_id, name, description, promotion_tier, logo_url, banner_url, created_at, updated_at
			FROM "restaurant_brand"
			WHERE NOT (id = ANY($1))
			ORDER BY promotion_tier DESC, id ASC
			LIMIT $2
		`, excludeBrandIDs, limit)
		if err != nil {
			return nil, fmt.Errorf("recommend cold-start: %w", err)
		}
		defer rows.Close()
		return scanRecommendedBrands(rows)
	}

	rows, err := r.pool.Query(ctx, `
		WITH seed_categories AS (
			SELECT DISTINCT category_id
			FROM "restaurant_brand_category"
			WHERE restaurant_brand_id = ANY($1)
		)
		SELECT b.id, b.owner_profile_id, b.name, b.description, b.promotion_tier, b.logo_url, b.banner_url, b.created_at, b.updated_at
		FROM "restaurant_brand" b
		JOIN "restaurant_brand_category" rbc ON rbc.restaurant_brand_id = b.id
		WHERE rbc.category_id IN (SELECT category_id FROM seed_categories)
		  AND NOT (b.id = ANY($2))
		GROUP BY b.id, b.owner_profile_id, b.name, b.description, b.promotion_tier, b.logo_url, b.banner_url, b.created_at, b.updated_at
		ORDER BY COUNT(rbc.category_id) DESC, b.promotion_tier DESC, b.id ASC
		LIMIT $3
	`, seedBrandIDs, excludeBrandIDs, limit)
	if err != nil {
		return nil, fmt.Errorf("recommend by category: %w", err)
	}
	defer rows.Close()
	return scanRecommendedBrands(rows)
}

func scanRecommendedBrands(rows pgx.Rows) ([]domain.RestaurantBrand, error) {
	out := make([]domain.RestaurantBrand, 0, 8)
	for rows.Next() {
		var d restaurantBrandDB
		if err := rows.Scan(&d.ID, &d.OwnerProfileID, &d.Name, &d.Description, &d.PromotionTier, &d.LogoURL, &d.BannerURL, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan brand: %w", err)
		}
		out = append(out, d.toDomain())
	}
	return out, rows.Err()
}
