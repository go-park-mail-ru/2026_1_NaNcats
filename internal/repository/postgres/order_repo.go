package postgres

import (
	"context"
	"errors"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/repository"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type orderRepo struct {
	pool *pgxpool.Pool
}

func NewOrderRepo(pool *pgxpool.Pool) repository.OrderRepository {
	return &orderRepo{
		pool: pool,
	}
}

func (r *orderRepo) CreateOrder(ctx context.Context, order domain.Order) (string, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)

	orderQuery := `
		INSERT INTO "order" (
			client_account_id, restaurant_branch_id, client_address_id,
			total_cost, payment_method_id, yookassa_payment_id,
			status
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, public_id;
	`

	orderDishQuery := `
		INSERT INTO "order_dish" (order_id, dish_id, quantity, price)
		VALUES ($1, $2, $3, $4)
	`

	var orderID int
	var orderPublicID string
	err = tx.QueryRow(ctx, orderQuery,
		order.ClientID,
		order.RestaurantBranchID,
		order.ClientAddressID,
		order.TotalCost,
		order.PaymentMethodID,
		order.YookassaPaymentID,
		order.Status,
	).Scan(&orderID, &orderPublicID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation { // проверка на уникальность
			return "", domain.ErrPaymentMethodAlreadyExists
		}
		return "", err
	}

	batch := &pgx.Batch{}
	for _, item := range order.Items {
		batch.Queue(orderDishQuery, orderID, item.DishID, item.Quantity, item.Price)
	}

	br := tx.SendBatch(ctx, batch)
	defer br.Close()

	for i := 0; i < len(order.Items); i++ {
		_, err = br.Exec()
		if err != nil {
			return "", err
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return "", err
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
		return err
	}

	rowsAffected := tag.RowsAffected()
	if rowsAffected == 0 {
		return domain.ErrOrderNotFound
	}

	return nil
}

func (r *orderRepo) GetOrderByPublicID(ctx context.Context, publicID string, userID int) (domain.Order, error) {
	query := `
		SELECT id, client_account_id, courier_account_id, restaurant_branch_id, client_address_id, total_cost, payment_method_id, yookassa_payment_id, status
		FROM "order" WHERE public_id = $1 AND client_account_id = $2;
	`
	var order domain.Order
	err := r.pool.QueryRow(ctx, query, publicID, userID).Scan(
		&order.ID,
		&order.ClientID,
		&order.CourierID,
		&order.RestaurantBranchID,
		&order.ClientAddressID,
		&order.TotalCost,
		&order.PaymentMethodID,
		&order.YookassaPaymentID,
		&order.Status,
	)
	if err != nil {
		return domain.Order{}, err
	}

	return order, nil
}

func (r *orderRepo) SetYookassaID(ctx context.Context, orderPublicID, yookassaID string) error {
	query := `
		UPDATE "order"
		SET yookassa_payment_id = $1
		WHERE public_id = $2;
	`

	tag, err := r.pool.Exec(ctx, query, yookassaID, orderPublicID)
	if err != nil {
		return err
	}

	rowsAffected := tag.RowsAffected()
	if rowsAffected == 0 {
		return domain.ErrOrderNotFound
	}

	return nil
}
