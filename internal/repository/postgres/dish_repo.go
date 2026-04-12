package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/repository"
)

type dishDB struct {
	ID                int       `db:"id"`
	RestaurantBrandID int       `db:"restaurant_brand_id"`
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
	pool PgxPool
}

func NewDishRepo(pool PgxPool) repository.DishRepository {
	return &dishRepo{pool: pool}
}

func (r *dishRepo) GetDishesByRestaurantBrandID(ctx context.Context, restaurantBrandID, limit, offset int) ([]domain.Dish, error) {
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
		return nil, err
	}
	defer rows.Close()

	dbDishes, err := pgx.CollectRows(rows, pgx.RowToStructByName[dishDB])
	if err != nil {
		return nil, err
	}

	dishes := make([]domain.Dish, 0, len(dbDishes))
	for _, d := range dbDishes {
		dishes = append(dishes, d.toDomain())
	}
	return dishes, nil
}

func (r *dishRepo) GetDishByID(ctx context.Context, DishID int) (domain.Dish, error) {
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
		return domain.Dish{}, err
	}

	dbDish, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[dishDB])

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Dish{}, domain.ErrDishNotFound
		}
		return domain.Dish{}, err
	}

	return dbDish.toDomain(), nil
}
