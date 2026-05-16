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

func (r *orderRepo) CreateOrder(ctx context.Context, order domain.Order, idempotencyKey string) (string, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var payloadBytes []byte
	err = tx.QueryRow(ctx, `
		INSERT INTO "idempotency_records" (user_id, idempotency_key, grpc_method)
		VALUES ($1, $2, 'CreateOrder')
		ON CONFLICT (user_id, idempotency_key) DO NOTHING
		RETURNING response_payload;
	`, order.AdminID, idempotencyKey).Scan(&payloadBytes)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			err = tx.QueryRow(ctx, `
				SELECT response_payload FROM "idempotency_records" 
				WHERE user_id = $1 AND idempotency_key = $2
			`, order.AdminID, idempotencyKey).Scan(&payloadBytes)
			if err != nil {
				return "", fmt.Errorf("failed to fetch existing idempotency record: %w", err)
			}

			if payloadBytes == nil {
				return "", fmt.Errorf("request is already in progress")
			}

			var savedResp idempotencyResponse
			if err := easyjson.Unmarshal(payloadBytes, &savedResp); err != nil {
				return "", fmt.Errorf("failed to unmarshal idempotency payload: %w", err)
			}
			return savedResp.PublicID, nil
		}
		return "", fmt.Errorf("failed to insert idempotency record: %w", err)
	}

	orderQuery := `
		INSERT INTO "order" (
			admin_account_id, restaurant_branch_id, restaurant_brand_id,
			restaurant_name, client_address_id, total_cost, status
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
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
	).Scan(&orderID, &orderPublicID)

	if err != nil {
		return "", fmt.Errorf("insert master order: %w", err)
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
			INSERT INTO "order_split" (id, order_id, user_id, amount, status, payment_method_id)
			VALUES ($1, $2, $3, $4, 'pending', $5)
		`
		for _, split := range order.Splits {
			batch.Queue(splitQuery, split.ID, orderID, split.UserID, split.Amount, split.PaymentMethodID)
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

	respData, _ := easyjson.Marshal(idempotencyResponse{PublicID: orderPublicID})
	_, err = tx.Exec(ctx, `
		UPDATE "idempotency_records" 
		SET response_payload = $1 
		WHERE user_id = $2 AND idempotency_key = $3
	`, respData, order.AdminID, idempotencyKey)
	if err != nil {
		return "", fmt.Errorf("failed to update idempotency payload: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return "", fmt.Errorf("commit tx: %w", err)
	}

	return orderPublicID, nil
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
	// PromocodeID в домене это *int64 (необязательный FK): прокидываем
	// указатель как есть, без разыменования.
	o.PromocodeID = promoID

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

func (r *orderRepo) GetOrdersByUserID(ctx context.Context, userID int64, limit, offset int32) ([]domain.Order, error) {
	query := `
		SELECT DISTINCT o.id, o.public_id, o.admin_account_id, o.restaurant_branch_id, 
		o.restaurant_brand_id, o.restaurant_name, o.total_cost, o.status, o.created_at
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
		if err := rows.Scan(&o.ID, &o.PublicID, &o.AdminID, &o.RestaurantBranchID, &o.RestaurantBrandID, &o.RestaurantName, &o.TotalCost, &o.Status, &o.CreatedAt); err == nil {
			orders = append(orders, o)
			orderIDs = append(orderIDs, o.ID)
		}
	}

	if len(orderIDs) == 0 {
		return orders, nil
	}

	splitQuery := `SELECT order_id, id, user_id, amount, status FROM "order_split" WHERE order_id = ANY($1)`
	splitRows, err := r.pool.Query(ctx, splitQuery, orderIDs)
	if err != nil {
		return nil, fmt.Errorf("batch fetch splits: %w", err)
	}
	defer splitRows.Close()

	splitsMap := make(map[int64][]domain.OrderSplit)
	for splitRows.Next() {
		var s domain.OrderSplit
		var orderID int64
		if err := splitRows.Scan(&orderID, &s.ID, &s.UserID, &s.Amount, &s.Status); err == nil {
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
