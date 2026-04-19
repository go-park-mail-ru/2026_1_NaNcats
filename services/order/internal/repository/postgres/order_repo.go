package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/order/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/order/internal/repository"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/postgres"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type orderDB struct {
	ID                 int64     `db:"id"`
	PublicID           string    `db:"public_id"`
	ClientID           int64     `db:"client_account_id"`
	CourierID          *int64    `db:"courier_account_id"`
	RestaurantBranchID int64     `db:"restaurant_branch_id"`
	ClientAddressID    string    `db:"client_address_id"`
	TotalCost          int64     `db:"total_cost"`
	PromocodeID        *int64    `db:"promocode_id"`
	RestaurantName     string    `db:"restaurant_name"`
	PaymentMethodID    *string   `db:"payment_method_id"`
	YookassaPaymentID  *string   `db:"yookassa_payment_id"`
	Status             string    `db:"status"`
	CreatedAt          time.Time `db:"created_at"`
	UpdatedAt          time.Time `db:"updated_at"`
}

func (o orderDB) toDomain() domain.Order {
	order := domain.Order{
		ID:                 o.ID,
		PublicID:           o.PublicID,
		ClientID:           o.ClientID,
		RestaurantBranchID: o.RestaurantBranchID,
		RestaurantName:     o.RestaurantName,
		ClientAddressID:    o.ClientAddressID,
		TotalCost:          o.TotalCost,
		Status:             o.Status,
		CreatedAt:          o.CreatedAt,
		UpdatedAt:          o.UpdatedAt,
	}

	if o.CourierID != nil {
		order.CourierID = *o.CourierID
	}
	if o.PromocodeID != nil {
		order.PromocodeID = *o.PromocodeID
	}
	if o.PaymentMethodID != nil {
		order.PaymentMethodID = *o.PaymentMethodID
	}
	if o.YookassaPaymentID != nil {
		order.YookassaPaymentID = *o.YookassaPaymentID
	}

	return order
}

type orderRepo struct {
	pool postgres.PgxPool
}

func NewOrderRepo(pool postgres.PgxPool) repository.OrderRepository {
	return &orderRepo{
		pool: pool,
	}
}

func (r *orderRepo) CreateOrder(ctx context.Context, order domain.Order, idempotencyKey string) (string, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	orderQuery := `
			INSERT INTO "order" (
				client_account_id, restaurant_branch_id, restaurant_name,
				client_address_id, total_cost, payment_method_id,
				yookassa_payment_id, status, idempotency_key
			)
			VALUES ($1, $2, $3, $4, $5, NULLIF($6, ''), NULLIF($7, ''), $8, NULLIF($9, ''))
			RETURNING id, public_id;
		`

	var orderID int64
	var orderPublicID string

	err = tx.QueryRow(ctx, orderQuery,
		order.ClientID,
		order.RestaurantBranchID,
		order.RestaurantName,
		order.ClientAddressID,
		order.TotalCost,
		order.PaymentMethodID,
		"",
		order.Status,
		idempotencyKey,
	).Scan(&orderID, &orderPublicID)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return "", fmt.Errorf("order with this idempotency key already exists: %w", err)
		}
		return "", fmt.Errorf("insert order: %w", err)
	}

	if len(order.Items) > 0 {
		orderDishQuery := `
				INSERT INTO "order_dish" (order_id, dish_id, quantity, price)
				VALUES ($1, $2, $3, $4)
			`
		batch := &pgx.Batch{}
		for _, item := range order.Items {
			batch.Queue(orderDishQuery, orderID, item.DishID, item.Quantity, item.Price)
		}

		br := tx.SendBatch(ctx, batch)
		for i := 0; i < len(order.Items); i++ {
			if _, execErr := br.Exec(); execErr != nil {
				br.Close()
				return "", fmt.Errorf("insert order_dish at index %d: %w", i, execErr)
			}
		}
		br.Close()
	}

	if err = tx.Commit(ctx); err != nil {
		return "", fmt.Errorf("commit tx: %w", err)
	}

	return orderPublicID, nil
}

func (r *orderRepo) UpdateStatusByPaymentID(ctx context.Context, yookassaPaymentID, newStatus string) error {
	query := `
			UPDATE "order"
			SET status = $1, updated_at = NOW()
			WHERE yookassa_payment_id = $2;
		`
	tag, err := r.pool.Exec(ctx, query, newStatus, yookassaPaymentID)
	if err != nil {
		return fmt.Errorf("update status: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return errors.New("order not found by payment ID")
	}
	return nil
}

func (r *orderRepo) GetOrderByPublicID(ctx context.Context, publicID string, userID int64) (domain.Order, error) {
	query := `
			SELECT 
				id, public_id, client_account_id, courier_account_id, 
				restaurant_branch_id, client_address_id, total_cost, 
				promocode_id, restaurant_name, payment_method_id,
				yookassa_payment_id, status, created_at, updated_at
			FROM "order" 
			WHERE public_id = $1 AND client_account_id = $2;
		`

	rows, err := r.pool.Query(ctx, query, publicID, userID)
	if err != nil {
		return domain.Order{}, fmt.Errorf("query order by public id: %w", err)
	}

	dbOrder, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[orderDB])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Order{}, errors.New("order not found")
		}
		return domain.Order{}, fmt.Errorf("collect one order row: %w", err)
	}

	return dbOrder.toDomain(), nil
}

func (r *orderRepo) SetYookassaID(ctx context.Context, orderPublicID, yookassaID string) error {
	query := `
			UPDATE "order"
			SET yookassa_payment_id = $1, updated_at = NOW()
			WHERE public_id = $2;
		`
	tag, err := r.pool.Exec(ctx, query, yookassaID, orderPublicID)
	if err != nil {
		return fmt.Errorf("set yookassa id: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return errors.New("order not found")
	}
	return nil
}

func (r *orderRepo) GetOrdersByUserID(ctx context.Context, userID int64) ([]domain.Order, error) {
	query := `
    SELECT 
        id, public_id, client_account_id, courier_account_id, 
        restaurant_branch_id, restaurant_name, -- ДОБАВИЛИ
        client_address_id, total_cost, 
        promocode_id, payment_method_id, yookassa_payment_id,
        status, created_at, updated_at
    	FROM "order"
    	WHERE client_account_id = $1
    	ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("query user orders: %w", err)
	}

	dbOrders, err := pgx.CollectRows(rows, pgx.RowToStructByName[orderDB])
	if err != nil {
		return nil, fmt.Errorf("collect rows: %w", err)
	}

	orders := make([]domain.Order, 0, len(dbOrders))
	for _, dbO := range dbOrders {
		orders = append(orders, dbO.toDomain())
	}

	return orders, nil
}
