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

type dishDB struct {
	ID                int64     `db:"id"`
	RestaurantBrandID int64     `db:"restaurant_brand_id"`
	Name              string    `db:"name"`
	Description       *string   `db:"description"`
	ImageURL          *string   `db:"image_url"`
	Price             int64     `db:"price"`
	CreatedAt         time.Time `db:"created_at"`
	UpdatedAt         time.Time `db:"updated_at"`
}

func (d dishDB) toDomain() domain.Dish {
	desc := ""
	if d.Description != nil {
		desc = *d.Description
	}
	img := ""
	if d.ImageURL != nil {
		img = *d.ImageURL
	}

	return domain.Dish{
		ID:                d.ID,
		RestaurantBrandID: d.RestaurantBrandID,
		Name:              d.Name,
		Description:       desc,
		ImageURL:          img,
		Price:             d.Price,
		CreatedAt:         d.CreatedAt,
		UpdatedAt:         d.UpdatedAt,
	}
}

type dishRepo struct {
	pool postgres.PgxPool
}

func NewDishRepo(pool postgres.PgxPool) repository.DishRepository {
	return &dishRepo{pool: pool}
}

func (r *dishRepo) SearchDishes(ctx context.Context, query string, limit int) ([]domain.Dish, error) {
	q := `
		SELECT id, restaurant_brand_id, name, description, image_url, price, created_at, updated_at
		FROM "dish"
		WHERE name ILIKE $1 OR description ILIKE $1
		ORDER BY id ASC
		LIMIT $2;
	`
	pattern := "%" + query + "%"
	rows, err := r.pool.Query(ctx, q, pattern, limit)
	if err != nil {
		return nil, fmt.Errorf("search dishes: %w", err)
	}
	defer rows.Close()

	dbDishes, err := pgx.CollectRows(rows, pgx.RowToStructByName[dishDB])
	if err != nil {
		return nil, fmt.Errorf("scan search dishes: %w", err)
	}

	dishes := make([]domain.Dish, 0, len(dbDishes))
	for _, d := range dbDishes {
		dishes = append(dishes, d.toDomain())
	}
	return dishes, nil
}

func (r *dishRepo) SearchDishesByBrand(ctx context.Context, brandID int64, query string, limit int) ([]domain.Dish, error) {
	q := `
		SELECT id, restaurant_brand_id, name, description, image_url, price, created_at, updated_at
		FROM "dish"
		WHERE restaurant_brand_id = $1
		  AND (name ILIKE $2 OR description ILIKE $2)
		ORDER BY id ASC
		LIMIT $3;
	`
	pattern := "%" + query + "%"
	rows, err := r.pool.Query(ctx, q, brandID, pattern, limit)
	if err != nil {
		return nil, fmt.Errorf("search dishes by brand: %w", err)
	}
	defer rows.Close()

	dbDishes, err := pgx.CollectRows(rows, pgx.RowToStructByName[dishDB])
	if err != nil {
		return nil, fmt.Errorf("scan search dishes by brand: %w", err)
	}

	dishes := make([]domain.Dish, 0, len(dbDishes))
	for _, d := range dbDishes {
		dishes = append(dishes, d.toDomain())
	}
	return dishes, nil
}

func (r *dishRepo) GetDishesByRestaurantBrandID(ctx context.Context, restaurantBrandID int64, limit, offset int) ([]domain.Dish, error) {
	query := `
		SELECT
			id,
			restaurant_brand_id,
			name,
			description,
			image_url,
			price,
			created_at,
			updated_at
		FROM "dish"
		WHERE restaurant_brand_id = $1
		ORDER BY id ASC
		LIMIT $2 OFFSET $3;
	`

	rows, err := r.pool.Query(ctx, query, restaurantBrandID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("postgres query error: %w", err)
	}
	defer rows.Close()

	dbDishes, err := pgx.CollectRows(rows, pgx.RowToStructByName[dishDB])
	if err != nil {
		return nil, fmt.Errorf("mapping dishes: %w", err)
	}

	dishes := make([]domain.Dish, 0, len(dbDishes))
	for _, d := range dbDishes {
		dishes = append(dishes, d.toDomain())
	}
	return dishes, nil
}

func (r *dishRepo) GetDishByID(ctx context.Context, DishID int64) (domain.Dish, error) {
	query := `
		SELECT id,
			restaurant_brand_id,
			name,
			description,
			image_url,
			price,
			created_at,
			updated_at
		FROM "dish"
		WHERE id = $1
	`

	rows, err := r.pool.Query(ctx, query, DishID)
	if err != nil {
		return domain.Dish{}, fmt.Errorf("query dish by id: %w", err)
	}

	dbDish, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[dishDB])

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Dish{}, domain.ErrDishNotFound
		}
		return domain.Dish{}, fmt.Errorf("scan dish row: %w", err)
	}

	return dbDish.toDomain(), nil
}

func (r *dishRepo) GetDishesByIDs(ctx context.Context, ids []int64) ([]domain.Dish, error) {
	query := `
		SELECT id,
			restaurant_brand_id,
			name,
			description,
			image_url,
			price,
			created_at,
			updated_at
		FROM "dish"
		WHERE id=ANY($1);
	`

	rows, err := r.pool.Query(ctx, query, ids)
	if err != nil {
		return nil, fmt.Errorf("query dishes by ids: %w", err)
	}

	dbDishes, err := pgx.CollectRows(rows, pgx.RowToStructByName[dishDB])
	if err != nil {
		return nil, fmt.Errorf("scan dishes rows: %w", err)
	}

	dishes := make([]domain.Dish, 0, len(dbDishes))
	for _, dish := range dbDishes {
		dishes = append(dishes, dish.toDomain())
	}

	return dishes, nil
}

func (r *dishRepo) Create(ctx context.Context, d domain.Dish, idemKey string) (domain.Dish, error) {
	query := `
		INSERT INTO "dish" (restaurant_brand_id, name, description, price, image_url, idempotency_key)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (idempotency_key) DO UPDATE SET updated_at = NOW()
		RETURNING id, restaurant_brand_id, name, description, price, image_url, created_at, updated_at;
	`
	var res dishDB
	err := r.pool.QueryRow(ctx, query, d.RestaurantBrandID, d.Name, d.Description, d.Price, d.ImageURL, idemKey).Scan(
		&res.ID, &res.RestaurantBrandID, &res.Name, &res.Description, &res.Price, &res.ImageURL, &res.CreatedAt, &res.UpdatedAt,
	)
	if err != nil {
		return domain.Dish{}, fmt.Errorf("create dish: %w", err)
	}
	return res.toDomain(), nil
}

func (r *dishRepo) Update(ctx context.Context, d domain.Dish) (domain.Dish, error) {
	query := `
		UPDATE "dish" 
		SET name = $1, description = $2, price = $3, image_url = $4, updated_at = NOW()
		WHERE id = $5
		RETURNING id, restaurant_brand_id, name, description, price, image_url, created_at, updated_at;
	`
	var res dishDB
	err := r.pool.QueryRow(ctx, query, d.Name, d.Description, d.Price, d.ImageURL, d.ID).Scan(
		&res.ID, &res.RestaurantBrandID, &res.Name, &res.Description, &res.Price, &res.ImageURL, &res.CreatedAt, &res.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Dish{}, domain.ErrDishNotFound
		}
		return domain.Dish{}, err
	}
	return res.toDomain(), nil
}

func (r *dishRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM "dish" WHERE id = $1`, id)
	return err
}
