package cart

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/cart/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/cart/internal/repository"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/postgres"
	"github.com/jackc/pgx/v5"
)

type cartItemDB struct {
	RestaurantID int64     `db:"restaurant_brand_id"`
	UpdatedAt    time.Time `db:"updated_at"`
	// Указатели для обработки пустой корзины
	DishID   *int64 `db:"dish_id"`
	Quantity *int   `db:"quantity"`
}

func (r cartItemDB) toDomainItem() domain.CartItem {
	item := domain.CartItem{}
	if r.DishID != nil {
		item.DishID = *r.DishID
	}
	if r.Quantity != nil {
		item.Quantity = *r.Quantity
	}
	return item
}

type cartRepo struct {
	pool postgres.PgxPool
}

func NewCartRepo(pool postgres.PgxPool) repository.CartRepository {
	return &cartRepo{
		pool: pool,
	}
}

func (r *cartRepo) GetCartByUserID(ctx context.Context, userID int64) (domain.Cart, error) {
	query := `
		SELECT 
			c.restaurant_brand_id,
			c.updated_at,
			c.status,
			cd.dish_id,
			cd.quantity
		FROM "cart" c
		LEFT JOIN "cart_dish" cd ON c.client_account_id = cd.cart_id
		WHERE c.client_account_id = $1
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return domain.Cart{}, fmt.Errorf("query cart: %w", err)
	}

	dbRows, err := pgx.CollectRows(rows, pgx.RowToStructByName[cartItemDB])
	if err != nil {
		return domain.Cart{}, fmt.Errorf("scan cart rows: %w", err)
	}

	if len(dbRows) == 0 {
		return domain.Cart{UserID: userID, Items: []domain.CartItem{}}, nil
	}

	cart := domain.Cart{
		UserID:            userID,
		RestaurantBrandID: dbRows[0].RestaurantID,
		UpdatedAt:         dbRows[0].UpdatedAt,
		Items:             make([]domain.CartItem, 0, len(dbRows)),
	}

	for _, row := range dbRows {
		if row.DishID != nil {
			cart.Items = append(cart.Items, row.toDomainItem())
		}
	}

	return cart, nil
}

func (r *cartRepo) UpdateCart(ctx context.Context, userID int64, resID int64, items []domain.CartItem) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var currentStatus string
	err = tx.QueryRow(ctx, `SELECT status FROM "cart" WHERE client_account_id = $1 FOR UPDATE`, userID).Scan(&currentStatus)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("get cart status for update: %w", err)
		}
	}
	if currentStatus == "locked" {
		return fmt.Errorf("cannot update cart: status is 'locked' during payment processing")
	}

	batch := &pgx.Batch{}

	batch.Queue(`
		INSERT INTO "cart" (client_account_id, restaurant_brand_id, status, updated_at)
		VALUES ($1, $2, 'active', NOW())
		ON CONFLICT (client_account_id) 
		DO UPDATE SET restaurant_brand_id = $2, updated_at = NOW();`,
		userID, resID)

	batch.Queue(`DELETE FROM "cart_dish" WHERE cart_id = $1`, userID)

	if len(items) > 0 {
		dishIDs := make([]int64, len(items))
		quantities := make([]int, len(items))
		for i, item := range items {
			dishIDs[i] = item.DishID
			quantities[i] = item.Quantity
		}

		batch.Queue(`
			INSERT INTO "cart_dish" (cart_id, dish_id, quantity, updated_at)
			SELECT $1, unnest($2::bigint[]), unnest($3::int[]), NOW()`,
			userID, dishIDs, quantities)
	}

	br := tx.SendBatch(ctx, batch)
	defer br.Close()

	for i := 0; i < batch.Len(); i++ {
		_, execErr := br.Exec()
		if execErr != nil {
			return fmt.Errorf("cart batch execution failed at step %d: %w", i, execErr)
		}
	}

	if err := br.Close(); err != nil {
		return fmt.Errorf("close batch results: %w", err)
	}

	return tx.Commit(ctx)
}

func (r *cartRepo) ClearCart(ctx context.Context, userID int64) error {
	query := `
		DELETE FROM cart WHERE client_account_id = $1
	`
	if _, err := r.pool.Exec(ctx, query, userID); err != nil {
		return fmt.Errorf("clear cart: %w", err)
	}
	return nil
}

func (r *cartRepo) LockCart(ctx context.Context, userID int64) error {
	query := `UPDATE "cart" SET status = 'locked', updated_at = NOW() WHERE client_account_id = $1`
	res, err := r.pool.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("lock cart: %w", err)
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("cart not found")
	}
	return nil
}

func (r *cartRepo) UnlockCart(ctx context.Context, userID int64) error {
	query := `UPDATE "cart" SET status = 'active', updated_at = NOW() WHERE client_account_id = $1`
	res, err := r.pool.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("unlock cart: %w", err)
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("cart not found")
	}
	return nil
}
