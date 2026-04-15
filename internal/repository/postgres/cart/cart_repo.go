package postgres

import (
	"context"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/repository"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/repository/postgres"
	"github.com/jackc/pgx/v5"
)

type cartItemDB struct {
	RestaurantID int       `db:"restaurant_brand_id"`
	UpdatedAt    time.Time `db:"updated_at"`
	// Поля из left должны быть указателями чтобы обработать null (пустую корзину)
	DishID   *int    `db:"dish_id"`
	Quantity *int    `db:"quantity"`
	Name     *string `db:"name"`
	Price    *int64  `db:"price"`
	ImageURL *string `db:"image_url"`
}

func (r cartItemDB) toDomainItem() domain.CartItem {
	item := domain.CartItem{}
	if r.DishID != nil {
		item.DishID = *r.DishID
	}
	if r.Quantity != nil {
		item.Quantity = *r.Quantity
	}
	if r.Name != nil {
		item.Name = *r.Name
	}
	if r.Price != nil {
		item.Price = *r.Price
	}
	if r.ImageURL != nil {
		item.ImageURL = *r.ImageURL
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

func (r *cartRepo) GetCartByUserID(ctx context.Context, userID int) (domain.Cart, error) {
	query := `
		SELECT 
			c.restaurant_brand_id,
			c.updated_at,
			cd.dish_id,
			cd.quantity,
			d.name,
			d.price,
			d.image_url
		FROM cart c
		LEFT JOIN cart_dish cd ON c.client_account_id = cd.cart_id
		LEFT JOIN dish d ON cd.dish_id = d.id
		WHERE c.client_account_id = $1
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return domain.Cart{}, err
	}

	dbRows, err := pgx.CollectRows(rows, pgx.RowToStructByName[cartItemDB])
	if err != nil {
		return domain.Cart{}, err
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

func (r *cartRepo) UpdateCart(ctx context.Context, userID int, resID int, items []domain.CartItem) error {
	batch := &pgx.Batch{}
	batch.Queue(`
		INSERT INTO cart (client_account_id, restaurant_brand_id, updated_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (client_account_id) 
		DO UPDATE SET restaurant_brand_id = $2, updated_at = NOW()`,
		userID, resID)

	batch.Queue(`DELETE FROM cart_dish WHERE cart_id = $1`, userID)

	if len(items) > 0 {
		dishIDs := make([]int, len(items))
		quantities := make([]int, len(items))
		for i, item := range items {
			dishIDs[i] = item.DishID
			quantities[i] = item.Quantity
		}

		batch.Queue(`
			INSERT INTO cart_dish (cart_id, dish_id, quantity)
			SELECT $1, unnest($2::int[]), unnest($3::int[])`,
			userID, dishIDs, quantities)
	}

	// Отправляем весь пакет в рамках одной транзакции
	br := r.pool.SendBatch(ctx, batch)
	defer br.Close()

	// Нужно пройтись по результатам всех команд в батче, чтобы поймать ошибку
	for i := 0; i < batch.Len(); i++ {
		_, err := br.Exec()
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *cartRepo) ClearCart(ctx context.Context, userId int) error {
	query := `
		DELETE FROM cart WHERE client_account_id = $1
	`
	_, err := r.pool.Exec(ctx, query, userId)
	return err
}
