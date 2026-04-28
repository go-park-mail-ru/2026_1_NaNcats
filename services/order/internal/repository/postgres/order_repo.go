package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/order/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/order/internal/repository"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/postgres"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

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
			admin_account_id, restaurant_branch_id, restaurant_brand_id,
			restaurant_name, client_address_id, total_cost, status, idempotency_key
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NULLIF($8, ''))
		RETURNING id, public_id;
	`

	var orderID int64
	var orderPublicID string

	err = tx.QueryRow(ctx, orderQuery,
		order.AdminID,
		order.RestaurantBranchID,
		order.RestaurantBrandID,
		order.RestaurantName,
		order.ClientAddressID,
		order.TotalCost,
		order.Status,
		idempotencyKey,
	).Scan(&orderID, &orderPublicID)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return "", fmt.Errorf("order with this idempotency key already exists: %w", err)
		}
		return "", fmt.Errorf("insert master order: %w", err)
	}

	batch := &pgx.Batch{}

	if len(order.Items) > 0 {
		dishQuery := `
			INSERT INTO "order_dish" (order_id, dish_id, quantity, price, owner_user_id)
			VALUES ($1, $2, $3, $4, $5)
		`
		for _, item := range order.Items {
			batch.Queue(dishQuery, orderID, item.DishID, item.Quantity, item.Price, item.OwnerUserID)
		}
	}

	if len(order.Splits) > 0 {
		splitQuery := `
			INSERT INTO "order_split" (id, order_id, user_id, amount, status)
			VALUES ($1, $2, $3, $4, 'pending')
		`
		for _, split := range order.Splits {
			batch.Queue(splitQuery, split.ID, orderID, split.UserID, split.Amount)
		}
	}

	br := tx.SendBatch(ctx, batch)

	for i := 0; i < batch.Len(); i++ {
		if _, execErr := br.Exec(); execErr != nil {
			br.Close()
			return "", fmt.Errorf("batch execution failed at step %d: %w", i, execErr)
		}
	}
	br.Close()

	if err = tx.Commit(ctx); err != nil {
		return "", fmt.Errorf("commit tx: %w", err)
	}

	return orderPublicID, nil
}

func (r *orderRepo) UpdateSplitStatusByPaymentID(ctx context.Context, yookassaPaymentID, newStatus string) (string, error) {
	query := `
		UPDATE "order_split"
		SET status = $1, updated_at = NOW()
		WHERE yookassa_payment_id = $2
		RETURNING id;
	`
	var splitID string
	err := r.pool.QueryRow(ctx, query, newStatus, yookassaPaymentID).Scan(&splitID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", errors.New("split not found by payment ID")
		}
		return "", fmt.Errorf("update split status: %w", err)
	}

	return splitID, nil
}

func (r *orderRepo) AreAllSplitsPaid(ctx context.Context, orderPublicID string) (bool, error) {
	query := `
		SELECT COUNT(*) 
		FROM "order_split" os
		JOIN "order" o ON os.order_id = o.id
		WHERE o.public_id = $1 AND os.status != 'paid'
	`

	var unpaidCount int
	err := r.pool.QueryRow(ctx, query, orderPublicID).Scan(&unpaidCount)
	if err != nil {
		return false, fmt.Errorf("check unpaid splits: %w", err)
	}

	return unpaidCount == 0, nil
}

func (r *orderRepo) UpdateOrderStatus(ctx context.Context, publicID string, newStatus string) error {
	query := `UPDATE "order" SET status = $1, updated_at = NOW() WHERE public_id = $2`
	tag, err := r.pool.Exec(ctx, query, newStatus, publicID)
	if err != nil {
		return fmt.Errorf("update status by public id: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return errors.New("order not found")
	}
	return nil
}

func (r *orderRepo) SetSplitYookassaID(ctx context.Context, splitID string, yookassaID string) error {
	query := `UPDATE "order_split" SET yookassa_payment_id = $1, updated_at = NOW() WHERE id = $2`
	tag, err := r.pool.Exec(ctx, query, yookassaID, splitID)
	if err != nil {
		return fmt.Errorf("set yookassa id to split: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return errors.New("split not found")
	}
	return nil
}

func (r *orderRepo) GetSplitByID(ctx context.Context, splitID string) (domain.OrderSplit, error) {
	query := `
		SELECT id, order_id, user_id, amount, status, payment_method_id, yookassa_payment_id, created_at, updated_at 
		FROM "order_split" WHERE id = $1
	`
	var s domain.OrderSplit
	err := r.pool.QueryRow(ctx, query, splitID).Scan(
		&s.ID, &s.OrderID, &s.UserID, &s.Amount, &s.Status,
		&s.PaymentMethodID, &s.YookassaPaymentID, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.OrderSplit{}, errors.New("split not found")
		}
		return domain.OrderSplit{}, fmt.Errorf("query split by id: %w", err)
	}
	return s, nil
}

func (r *orderRepo) UpdateSplitPayer(ctx context.Context, splitID string, newPayerID int64) error {
	query := `UPDATE "order_split" SET user_id = $1, updated_at = NOW() WHERE id = $2 AND status != 'paid'`
	tag, err := r.pool.Exec(ctx, query, newPayerID, splitID)
	if err != nil {
		return fmt.Errorf("update split payer: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return errors.New("split not found or already paid")
	}
	return nil
}

func (r *orderRepo) UpdateSplitStatus(ctx context.Context, splitID string, newStatus string) error {
	query := `UPDATE "order_split" SET status = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.pool.Exec(ctx, query, newStatus, splitID)
	return err
}

func (r *orderRepo) GetOrderByPublicID(ctx context.Context, publicID string) (domain.Order, error) {
	orderQuery := `
		SELECT id, public_id, admin_account_id, courier_account_id, restaurant_branch_id, 
		restaurant_brand_id, client_address_id, total_cost, promocode_id, restaurant_name, 
		status, created_at, updated_at 
		FROM "order" WHERE public_id = $1
	`
	var o domain.Order
	var courierID, promoID *int64

	err := r.pool.QueryRow(ctx, orderQuery, publicID).Scan(
		&o.ID, &o.PublicID, &o.AdminID, &courierID, &o.RestaurantBranchID,
		&o.RestaurantBrandID, &o.ClientAddressID, &o.TotalCost, &promoID,
		&o.RestaurantName, &o.Status, &o.CreatedAt, &o.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Order{}, errors.New("order not found")
		}
		return domain.Order{}, fmt.Errorf("query master order: %w", err)
	}
	if courierID != nil {
		o.CourierID = *courierID
	}
	if promoID != nil {
		o.PromocodeID = *promoID
	}

	dishQuery := `SELECT dish_id, quantity, price, owner_user_id FROM "order_dish" WHERE order_id = $1`
	dishRows, _ := r.pool.Query(ctx, dishQuery, o.ID)
	defer dishRows.Close()
	for dishRows.Next() {
		var d domain.OrderDish
		if err := dishRows.Scan(&d.DishID, &d.Quantity, &d.Price, &d.OwnerUserID); err == nil {
			o.Items = append(o.Items, d)
		}
	}

	splitQuery := `
		SELECT id, user_id, amount, status, payment_method_id, yookassa_payment_id 
		FROM "order_split" WHERE order_id = $1
	`
	splitRows, _ := r.pool.Query(ctx, splitQuery, o.ID)
	defer splitRows.Close()
	for splitRows.Next() {
		var s domain.OrderSplit
		s.OrderID = o.ID
		if err := splitRows.Scan(&s.ID, &s.UserID, &s.Amount, &s.Status, &s.PaymentMethodID, &s.YookassaPaymentID); err == nil {
			o.Splits = append(o.Splits, s)
		}
	}

	return o, nil
}

func (r *orderRepo) GetOrdersByUserID(ctx context.Context, userID int64) ([]domain.Order, error) {
	query := `
		SELECT DISTINCT o.id, o.public_id, o.admin_account_id, o.restaurant_branch_id, 
		o.restaurant_brand_id, o.restaurant_name, o.total_cost, o.status, o.created_at
		FROM "order" o
		LEFT JOIN "order_split" os ON o.id = os.order_id
		WHERE o.admin_account_id = $1 OR os.user_id = $1
		ORDER BY o.created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []domain.Order
	for rows.Next() {
		var o domain.Order
		if err := rows.Scan(&o.ID, &o.PublicID, &o.AdminID, &o.RestaurantBranchID, &o.RestaurantBrandID, &o.RestaurantName, &o.TotalCost, &o.Status, &o.CreatedAt); err == nil {
			orders = append(orders, o)
		}
	}

	for i := range orders {
		splitRows, _ := r.pool.Query(ctx, `SELECT id, user_id, amount, status FROM "order_split" WHERE order_id = $1`, orders[i].ID)
		for splitRows.Next() {
			var s domain.OrderSplit
			if err := splitRows.Scan(&s.ID, &s.UserID, &s.Amount, &s.Status); err == nil {
				orders[i].Splits = append(orders[i].Splits, s)
			}
		}
		splitRows.Close()
	}

	return orders, nil
}
