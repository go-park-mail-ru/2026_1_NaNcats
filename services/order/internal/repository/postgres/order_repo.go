package postgres

//go:generate easyjson $GOFILE

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/order/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/order/internal/repository"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/postgres"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/mailru/easyjson"
)

//easyjson:json
type idempotencyResponse struct {
	PublicID string `json:"public_id,omitempty"`
}

type orderRepo struct {
	pool postgres.PgxPool
}

func NewOrderRepo(pool postgres.PgxPool) repository.OrderRepository {
	return &orderRepo{
		pool: pool,
	}
}

type txKey struct{}

type PgxQuerier interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults
}

func (r *orderRepo) getTxOrDB(ctx context.Context) PgxQuerier {
	if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return tx
	}
	return r.pool
}

func (r *orderRepo) WithTransaction(ctx context.Context, fn func(txCtx context.Context) error) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	txCtx := context.WithValue(ctx, txKey{}, tx)

	err = fn(txCtx)
	if err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil && !errors.Is(rbErr, pgx.ErrTxClosed) {
			return fmt.Errorf("rollback error: %v, original error: %w", rbErr, err)
		}
		return err
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}

func (r *orderRepo) CreateOrder(ctx context.Context, order domain.Order, idempotencyKey string) (int64, string, error) {
	q := r.getTxOrDB(ctx)

	var payloadBytes []byte
	err := q.QueryRow(ctx, `
		INSERT INTO "idempotency_records" (user_id, idempotency_key, grpc_method)
		VALUES ($1, $2, 'CreateOrder')
		ON CONFLICT (user_id, idempotency_key) DO NOTHING
		RETURNING response_payload;
	`, order.AdminID, idempotencyKey).Scan(&payloadBytes)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			err = q.QueryRow(ctx, `
				SELECT response_payload FROM "idempotency_records" 
				WHERE user_id = $1 AND idempotency_key = $2
			`, order.AdminID, idempotencyKey).Scan(&payloadBytes)
			if err != nil {
				return 0, "", fmt.Errorf("failed to fetch existing idempotency record: %w", err)
			}

			if payloadBytes == nil {
				return 0, "", fmt.Errorf("request is already in progress")
			}

			var savedResp idempotencyResponse
			if err := easyjson.Unmarshal(payloadBytes, &savedResp); err != nil {
				return 0, "", fmt.Errorf("failed to unmarshal idempotency payload: %w", err)
			}
			return 0, savedResp.PublicID, nil
		}
		return 0, "", fmt.Errorf("failed to insert idempotency record: %w", err)
	}

	orderQuery := `
		INSERT INTO "order" (
			admin_account_id, restaurant_branch_id, restaurant_brand_id,
			restaurant_name, client_address_id, total_cost, promocode_id, 
			discount_amount, promocode_code, status
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, public_id;
	`

	var orderID int64
	var orderPublicID string

	err = q.QueryRow(ctx, orderQuery,
		order.AdminID,
		order.RestaurantBranchID,
		order.RestaurantBrandID,
		order.RestaurantName,
		order.ClientAddressID,
		order.TotalCost,
		order.PromocodeID,
		order.DiscountAmount,
		order.PromocodeString,
		order.Status,
	).Scan(&orderID, &orderPublicID)

	if err != nil {
		return 0, "", fmt.Errorf("insert master order: %w", err)
	}

	batch := &pgx.Batch{}

	if len(order.Items) > 0 {
		dishQuery := `
			INSERT INTO "order_dish" (order_id, dish_id, dish_name, quantity, price, owner_user_id)
			VALUES ($1, $2, $3, $4, $5, $6)
		`
		for _, item := range order.Items {
			// owner_user_id входит в первичный ключ и не бывает NULL: ничьё сохраняем как 0.
			ownerID := int64(0)
			if item.OwnerUserID != nil {
				ownerID = *item.OwnerUserID
			}
			batch.Queue(dishQuery, orderID, item.DishID, item.Name, item.Quantity, item.Price, ownerID)
		}
	}

	if len(order.Splits) > 0 {
		splitQuery := `
			INSERT INTO "order_split" (id, order_id, user_id, base_amount, discount_amount, amount, status, payment_method_id)
			VALUES ($1, $2, $3, $4, $5, $6, 'pending', $7)
		`
		for _, split := range order.Splits {
			batch.Queue(splitQuery,
				split.ID, orderID, split.UserID,
				split.BaseAmount, split.DiscountAmount,
				split.Amount, split.PaymentMethodID,
			)
		}
	}

	br := q.SendBatch(ctx, batch)
	for i := 0; i < batch.Len(); i++ {
		if _, execErr := br.Exec(); execErr != nil {
			br.Close()
			return 0, "", fmt.Errorf("batch execution failed at step %d: %w", i, execErr)
		}
	}
	br.Close()

	respData, _ := easyjson.Marshal(idempotencyResponse{PublicID: orderPublicID})
	_, err = q.Exec(ctx, `
		UPDATE "idempotency_records" 
		SET response_payload = $1 
		WHERE user_id = $2 AND idempotency_key = $3
	`, respData, order.AdminID, idempotencyKey)
	if err != nil {
		return 0, "", fmt.Errorf("failed to update idempotency payload: %w", err)
	}

	return orderID, orderPublicID, nil
}

// UpdateSplitStatusByPaymentID помечает долю счёта по её платежу в YooKassa и
// возвращает идентификатор самой доли и публичный идентификатор заказа.
func (r *orderRepo) UpdateSplitStatusByPaymentID(ctx context.Context, yookassaPaymentID, newStatus string) (string, string, error) {
	query := `
		UPDATE "order_split"
		SET status = $1, updated_at = NOW()
		WHERE yookassa_payment_id = $2
		RETURNING id, (SELECT public_id FROM "order" WHERE id = "order_split".order_id);
	`
	var splitID, orderPublicID string
	err := r.pool.QueryRow(ctx, query, newStatus, yookassaPaymentID).Scan(&splitID, &orderPublicID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", "", errors.New("split not found by payment ID")
		}
		return "", "", fmt.Errorf("update split status: %w", err)
	}

	return splitID, orderPublicID, nil
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

func (r *orderRepo) UpdateOrderStatus(ctx context.Context, publicID string, newStatus string, expectedStatuses ...string) error {
	var query string
	var tag pgconn.CommandTag
	var err error

	if len(expectedStatuses) > 0 {
		query = `UPDATE "order" SET status = $1, updated_at = NOW() WHERE public_id = $2 AND status = ANY($3)`
		tag, err = r.pool.Exec(ctx, query, newStatus, publicID, expectedStatuses)
	} else {
		query = `UPDATE "order" SET status = $1, updated_at = NOW() WHERE public_id = $2`
		tag, err = r.pool.Exec(ctx, query, newStatus, publicID)
	}

	if err != nil {
		return fmt.Errorf("update status by public id: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return repository.ErrStateChanged
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
		SELECT s.id, s.order_id, s.user_id, s.amount, s.status, s.payment_method_id,
		       s.yookassa_payment_id, s.created_at, s.updated_at, o.public_id
		FROM "order_split" s
		JOIN "order" o ON o.id = s.order_id
		WHERE s.id = $1
	`
	var s domain.OrderSplit
	err := r.pool.QueryRow(ctx, query, splitID).Scan(
		&s.ID, &s.OrderID, &s.UserID, &s.Amount, &s.Status,
		&s.PaymentMethodID, &s.YookassaPaymentID, &s.CreatedAt, &s.UpdatedAt, &s.OrderPublicID,
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
		restaurant_brand_id, client_address_id, total_cost, promocode_id, discount_amount, 
		promocode_code, restaurant_name, status, created_at, updated_at 
		FROM "order" WHERE public_id = $1
	`
	var o domain.Order
	var courierID *int64

	err := r.pool.QueryRow(ctx, orderQuery, publicID).Scan(
		&o.ID, &o.PublicID, &o.AdminID, &courierID, &o.RestaurantBranchID,
		&o.RestaurantBrandID, &o.ClientAddressID, &o.TotalCost, &o.PromocodeID,
		&o.DiscountAmount, &o.PromocodeString,
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

	dishQuery := `SELECT dish_id, dish_name, quantity, price, owner_user_id FROM "order_dish" WHERE order_id = $1`
	dishRows, _ := r.pool.Query(ctx, dishQuery, o.ID)
	defer dishRows.Close()
	for dishRows.Next() {
		var d domain.OrderDish
		if err := dishRows.Scan(&d.DishID, &d.Name, &d.Quantity, &d.Price, &d.OwnerUserID); err == nil {
			o.Items = append(o.Items, d)
		}
	}

	splitQuery := `
		SELECT id, user_id, base_amount, discount_amount, amount, status, payment_method_id, yookassa_payment_id 
		FROM "order_split" WHERE order_id = $1
	`
	splitRows, _ := r.pool.Query(ctx, splitQuery, o.ID)
	defer splitRows.Close()
	for splitRows.Next() {
		var s domain.OrderSplit
		s.OrderID = o.ID
		if err := splitRows.Scan(&s.ID, &s.UserID, &s.BaseAmount, &s.DiscountAmount, &s.Amount, &s.Status, &s.PaymentMethodID, &s.YookassaPaymentID); err == nil {
			o.Splits = append(o.Splits, s)
		}
	}

	return o, nil
}

func (r *orderRepo) GetOrdersByUserID(ctx context.Context, userID int64, limit, offset int32) ([]domain.Order, error) {
	query := `
		SELECT DISTINCT o.id, o.public_id, o.admin_account_id, o.restaurant_branch_id, 
		o.restaurant_brand_id, o.restaurant_name, o.total_cost, o.discount_amount, 
		o.promocode_code, o.status, o.created_at
		FROM "order" o
		LEFT JOIN "order_split" os ON o.id = os.order_id
		WHERE o.admin_account_id = $1 OR os.user_id = $1
		ORDER BY o.created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []domain.Order
	var orderIDs []int64

	for rows.Next() {
		var o domain.Order
		if err := rows.Scan(&o.ID, &o.PublicID, &o.AdminID, &o.RestaurantBranchID, &o.RestaurantBrandID, &o.RestaurantName, &o.TotalCost, &o.DiscountAmount, &o.PromocodeString, &o.Status, &o.CreatedAt); err == nil {
			orders = append(orders, o)
			orderIDs = append(orderIDs, o.ID)
		}
	}

	if len(orderIDs) == 0 {
		return orders, nil
	}

	splitQuery := `SELECT order_id, id, user_id, base_amount, discount_amount, amount, status FROM "order_split" WHERE order_id = ANY($1)`
	splitRows, err := r.pool.Query(ctx, splitQuery, orderIDs)
	if err != nil {
		return nil, fmt.Errorf("batch fetch splits: %w", err)
	}
	defer splitRows.Close()

	splitsMap := make(map[int64][]domain.OrderSplit)
	for splitRows.Next() {
		var s domain.OrderSplit
		var orderID int64
		if err := splitRows.Scan(&orderID, &s.ID, &s.UserID, &s.BaseAmount, &s.DiscountAmount, &s.Amount, &s.Status); err == nil {
			splitsMap[orderID] = append(splitsMap[orderID], s)
		}
	}

	dishQuery := `SELECT order_id, dish_id, dish_name, quantity, price, owner_user_id FROM "order_dish" WHERE order_id = ANY($1)`
	dishRows, err := r.pool.Query(ctx, dishQuery, orderIDs)
	if err != nil {
		return nil, fmt.Errorf("batch fetch dishes: %w", err)
	}
	defer dishRows.Close()

	dishesMap := make(map[int64][]domain.OrderDish)
	for dishRows.Next() {
		var d domain.OrderDish
		var orderID int64
		if err := dishRows.Scan(&orderID, &d.DishID, &d.Name, &d.Quantity, &d.Price, &d.OwnerUserID); err == nil {
			dishesMap[orderID] = append(dishesMap[orderID], d)
		}
	}

	for i := range orders {
		orders[i].Splits = splitsMap[orders[i].ID]
		orders[i].Items = dishesMap[orders[i].ID]
	}

	return orders, nil
}

func (r *orderRepo) GetOrdersByStatuses(ctx context.Context, statuses []string) ([]domain.Order, error) {
	if len(statuses) == 0 {
		return nil, nil
	}
	query := `
		SELECT id, public_id, admin_account_id, status
		FROM "order"
		WHERE status = ANY($1)
		ORDER BY created_at ASC
	`
	rows, err := r.pool.Query(ctx, query, statuses)
	if err != nil {
		return nil, fmt.Errorf("query orders by statuses: %w", err)
	}
	defer rows.Close()
	var out []domain.Order
	for rows.Next() {
		var o domain.Order
		if err := rows.Scan(&o.ID, &o.PublicID, &o.AdminID, &o.Status); err == nil {
			out = append(out, o)
		}
	}
	return out, nil
}

func (r *orderRepo) GetPromocodeByCodeWithLock(ctx context.Context, code string) (domain.Promocode, error) {
	query := `
		SELECT id, code, discount_percent, discount_amount, max_uses, current_uses, 
		       min_order_amount, user_id, restaurant_brand_id, is_global, created_at, expires_at
		FROM "promocode"
		WHERE code = $1
		FOR UPDATE
	`

	var p domain.Promocode

	err := r.getTxOrDB(ctx).QueryRow(ctx, query, code).Scan(
		&p.ID, &p.Code, &p.DiscountPercent, &p.DiscountAmount, &p.MaxUses, &p.CurrentUses,
		&p.MinOrderAmount, &p.UserID, &p.RestaurantBrandID, &p.IsGlobal, &p.CreatedAt, &p.ExpiresAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return p, errors.New("promocode not found")
		}
		return p, fmt.Errorf("failed to lock promocode: %w", err)
	}

	return p, nil
}

func (r *orderRepo) CheckPromocodeUsage(ctx context.Context, promoID, userID int64) (bool, error) {
	query := `
		SELECT EXISTS
		(SELECT 1 FROM "promocode_usage"
		WHERE promocode_id = $1 AND user_id = $2)
	`

	var exists bool

	err := r.getTxOrDB(ctx).QueryRow(ctx, query, promoID, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check promo usage: %w", err)
	}

	return exists, nil
}

func (r *orderRepo) IncrementPromocodeUses(ctx context.Context, promoID int64) error {
	query := `
		UPDATE "promocode" SET current_uses = current_uses + 1 WHERE id = $1`

	tag, err := r.getTxOrDB(ctx).Exec(ctx, query, promoID)
	if err != nil {
		return fmt.Errorf("failed to increment promo uses: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return errors.New("promocode not found during increment")
	}

	return nil
}

func (r *orderRepo) CreatePromocodeUsage(ctx context.Context, promoID, orderID, userID int64) error {
	query := `
		INSERT INTO "promocode_usage"
		(promocode_id, order_id, user_id)
		VALUES ($1, $2, $3)
	`

	_, err := r.getTxOrDB(ctx).Exec(ctx, query, promoID, orderID, userID)
	if err != nil {
		return fmt.Errorf("failed to create promo usage record: %w", err)
	}

	return nil
}

func (r *orderRepo) RollbackPromocodeUsage(ctx context.Context, orderPublicID string) error {
	return r.WithTransaction(ctx, func(txCtx context.Context) error {
		q := r.getTxOrDB(txCtx)

		var orderID int64
		var promoID *int64

		err := q.QueryRow(txCtx, `
			SELECT id, promocode_id 
			FROM "order" 
			WHERE public_id = $1
		`, orderPublicID).Scan(&orderID, &promoID)

		if err != nil {
			return fmt.Errorf("failed to fetch order for rollback: %w", err)
		}

		if promoID == nil {
			return nil
		}

		tag, err := q.Exec(txCtx, `
			DELETE FROM "promocode_usage" 
			WHERE order_id = $1 AND promocode_id = $2
		`, orderID, *promoID)

		if err != nil {
			return fmt.Errorf("failed to delete promocode usage: %w", err)
		}

		if tag.RowsAffected() == 0 {
			return nil
		}

		_, err = q.Exec(txCtx, `
			UPDATE "promocode" 
			SET current_uses = current_uses - 1 
			WHERE id = $1 AND current_uses > 0
		`, *promoID)

		if err != nil {
			return fmt.Errorf("failed to decrement promocode uses: %w", err)
		}

		return nil
	})
}
