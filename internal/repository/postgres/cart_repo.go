package postgres

import (
	"context"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/repository"
)

type cartRepo struct {
	pool PgxPool
}

func NewCartRepo(pool PgxPool) repository.CartRepository {
	return &cartRepo{
		pool: pool,
	}
}

func (r *cartRepo) GetCartByUserID(ctx context.Context, userID int) (domain.Cart, error) {
	query := `
		SELECT 
			c.restaurant_brand_id,
			cd.dish_id,
			COALESCE(cd.quantity, 0),
			COALESCE(d.name, ''),
			COALESCE(d.price, 0),
			COALESCE(d.image_url, '')
		FROM cart c
		LEFT JOIN cart_dish cd ON c.client_account_id = cd.cart_id
		LEFT JOIN dish d ON cd.dish_id = d.id
		WHERE c.client_account_id = $1
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return domain.Cart{}, err
	}
	defer rows.Close()

	cart := domain.Cart{
		UserID: userID,
		Items:  []domain.CartItem{},
	}

	found := false

	for rows.Next() {
		found = true

		var item domain.CartItem
		var resID int

		err := rows.Scan(
			&resID,
			&item.DishID,
			&item.Quantity,
			&item.Name,
			&item.Price,
			&item.ImageURL,
		)
		if err != nil {
			return domain.Cart{}, err
		}

		cart.RestaurantBrandID = resID

		if item.DishID != 0 {
			cart.Items = append(cart.Items, item)
		}
	}

	if !found {
		return domain.Cart{Items: []domain.CartItem{}, UserID: 0, RestaurantBrandID: 0, UpdatedAt: time.Time{}}, nil
	}

	cart.UpdatedAt = time.Now()

	return cart, nil
}

func (r *cartRepo) UpdateCart(ctx context.Context, userID int, resID int, items []domain.CartItem) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		INSERT INTO cart (client_account_id, restaurant_brand_id, updated_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (client_account_id) 
		DO UPDATE SET restaurant_brand_id = $2, updated_at = NOW()`,
		userID, resID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `DELETE FROM cart_dish WHERE cart_id = $1`, userID)
	if err != nil {
		return err
	}

	if len(items) > 0 {
		dishIDs := make([]int, len(items))
		quantities := make([]int, len(items))
		for i, item := range items {
			dishIDs[i] = item.DishID
			quantities[i] = item.Quantity
		}

		query := `
			INSERT INTO cart_dish (cart_id, dish_id, quantity)
			SELECT $1, unnest($2::int[]), unnest($3::int[])
		`
		_, err = tx.Exec(ctx, query, userID, dishIDs, quantities)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *cartRepo) ClearCart(ctx context.Context, userId int) error {
	query := `
		DELETE FROM cart WHERE client_account_id = $1
	`
	_, err := r.pool.Exec(ctx, query, userId)
	return err
}
