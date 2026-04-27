package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/cart/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/cart/internal/repository"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/postgres"
	"github.com/jackc/pgx/v5"
)

type cartRowDB struct {
	CartID            string     `db:"cart_id"`
	AdminID           int64      `db:"admin_id"`
	RestaurantBrandID int64      `db:"restaurant_brand_id"`
	Mode              string     `db:"mode"`
	Status            string     `db:"status"`
	UpdatedAt         time.Time  `db:"updated_at"`
	DishID            *int64     `db:"dish_id"`
	Quantity          *int32     `db:"quantity"`
	OwnerUserID       *int64     `db:"owner_user_id"`
	MemberUserID      *int64     `db:"member_user_id"`
	JoinedAt          *time.Time `db:"joined_at"`
}

type cartRepo struct {
	pool postgres.PgxPool
}

func NewCartRepo(pool postgres.PgxPool) repository.CartRepository {
	return &cartRepo{
		pool: pool,
	}
}

func insertOutboxEvent(ctx context.Context, tx pgx.Tx, aggregateID string, eventType string, payload any) error {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal outbox payload: %w", err)
	}

	query := `
		INSERT INTO "outbox_events" (aggregate_id, event_type, payload)
		VALUES ($1, $2, $3)
	`
	_, err = tx.Exec(ctx, query, aggregateID, eventType, payloadBytes)
	return err
}

func (r *cartRepo) AddItem(ctx context.Context, cartID string, item domain.CartItem) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var currentStatus string
	err = tx.QueryRow(ctx, `SELECT status FROM "cart" WHERE cart_id = $1 FOR UPDATE`, cartID).Scan(&currentStatus)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("cart not found")
		}
		return fmt.Errorf("lock cart for update: %w", err)
	}
	if currentStatus == domain.CartStatusLocked {
		return fmt.Errorf("cannot add item: cart is locked")
	}

	query := `
		INSERT INTO "cart_dish" (cart_id, dish_id, owner_user_id, quantity, updated_at)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (cart_id, dish_id) 
		DO UPDATE SET quantity = cart_dish.quantity + EXCLUDED.quantity, updated_at = NOW()
	`
	_, err = tx.Exec(ctx, query, cartID, item.DishID, item.OwnerUserID, item.Quantity)
	if err != nil {
		return fmt.Errorf("insert cart item: %w", err)
	}

	eventPayload := map[string]any{
		"dish_id":       item.DishID,
		"quantity":      item.Quantity,
		"owner_user_id": item.OwnerUserID,
		"action":        "ITEM_ADDED",
	}
	if err := insertOutboxEvent(ctx, tx, cartID, "CartItemAdded", eventPayload); err != nil {
		return fmt.Errorf("insert outbox event: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}

func (r *cartRepo) LockCart(ctx context.Context, cartID string) error {
	payload := map[string]any{
		"status":  domain.CartStatusLocked,
		"cart_id": cartID,
	}

	return r.execWithOutbox(ctx, cartID, "CartLocked", payload, func(tx pgx.Tx) error {
		query := `UPDATE "cart" SET status = $1, updated_at = NOW() WHERE cart_id = $2`
		_, err := tx.Exec(ctx, query, domain.CartStatusLocked, cartID)
		return err
	})
}

func (r *cartRepo) UnlockCart(ctx context.Context, cartID string) error {
	payload := map[string]any{"status": domain.CartStatusActive}

	return r.execWithOutbox(ctx, cartID, "CartUnlocked", payload, func(tx pgx.Tx) error {
		query := `UPDATE "cart" SET status = $1, updated_at = NOW() WHERE cart_id = $2`
		res, err := tx.Exec(ctx, query, domain.CartStatusActive, cartID)
		if err != nil {
			return fmt.Errorf("unlock cart: %w", err)
		}
		if res.RowsAffected() == 0 {
			return fmt.Errorf("cart not found")
		}
		return nil
	})
}

func (r *cartRepo) ClearCart(ctx context.Context, cartID string) error {
	return r.execWithOutbox(ctx, cartID, "CartCleared", map[string]any{}, func(tx pgx.Tx) error {
		query := `DELETE FROM "cart_dish" WHERE cart_id = $1`
		_, err := tx.Exec(ctx, query, cartID)
		if err != nil {
			return fmt.Errorf("clear cart dishes: %w", err)
		}

		return nil
	})
}

func (r *cartRepo) UpdateCartMode(ctx context.Context, cartID string, mode string) error {
	payload := map[string]any{"mode": mode}

	return r.execWithOutbox(ctx, cartID, "CartModeUpdated", payload, func(tx pgx.Tx) error {
		query := `UPDATE "cart" SET mode = $1, updated_at = NOW() WHERE cart_id = $2`
		res, err := tx.Exec(ctx, query, mode, cartID)
		if err != nil {
			return fmt.Errorf("update cart mode: %w", err)
		}
		if res.RowsAffected() == 0 {
			return fmt.Errorf("cart not found")
		}
		return nil
	})
}

func (r *cartRepo) DowngradeToSolo(ctx context.Context, cartID string, adminID int64) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	batch := &pgx.Batch{}

	batch.Queue(`DELETE FROM "cart_member" WHERE cart_id = $1 AND user_id != $2`, cartID, adminID)

	batch.Queue(`
		DELETE FROM "cart_dish" 
		WHERE cart_id = $1 AND (owner_user_id != $2 OR owner_user_id IS NULL)
	`, cartID, adminID)

	batch.Queue(`UPDATE "cart" SET mode = $1, updated_at = NOW() WHERE cart_id = $2`, domain.CartModeSolo, cartID)

	br := tx.SendBatch(ctx, batch)
	for i := 0; i < batch.Len(); i++ {
		if _, err := br.Exec(); err != nil {
			br.Close()
			return fmt.Errorf("downgrade batch step %d failed: %w", i, err)
		}
	}
	br.Close()

	eventPayload := map[string]any{
		"action":   "CART_CLOSED",
		"admin_id": adminID,
	}
	if err := insertOutboxEvent(ctx, tx, cartID, "SharedCartClosed", eventPayload); err != nil {
		return fmt.Errorf("insert outbox event: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}

func (r *cartRepo) GetCartByUserID(ctx context.Context, userID int64) (domain.Cart, error) {
	query := `
		SELECT c.cart_id FROM "cart" c
		LEFT JOIN "cart_member" cm ON c.cart_id = cm.cart_id
		WHERE c.admin_id = $1 OR cm.user_id = $1
		LIMIT 1
	`
	var cartID string
	err := r.pool.QueryRow(ctx, query, userID).Scan(&cartID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Cart{}, nil
		}
		return domain.Cart{}, fmt.Errorf("find cart by user: %w", err)
	}

	return r.GetCartByID(ctx, cartID)
}

func (r *cartRepo) GetCartByID(ctx context.Context, cartID string) (domain.Cart, error) {
	query := `
		SELECT 
			c.cart_id, c.admin_id, c.restaurant_brand_id, c.mode, c.status, c.updated_at,
			cd.dish_id, cd.quantity, cd.owner_user_id,
			cm.user_id as member_user_id, cm.joined_at
		FROM "cart" c
		LEFT JOIN "cart_dish" cd ON c.cart_id = cd.cart_id
		LEFT JOIN "cart_member" cm ON c.cart_id = cm.cart_id
		WHERE c.cart_id = $1
	`
	rows, err := r.pool.Query(ctx, query, cartID)
	if err != nil {
		return domain.Cart{}, err
	}
	defer rows.Close()

	dbRows, err := pgx.CollectRows(rows, pgx.RowToStructByName[cartRowDB])
	if err != nil {
		return domain.Cart{}, err
	}

	if len(dbRows) == 0 {
		return domain.Cart{}, domain.ErrDishNotFound
	}

	cart := domain.Cart{
		ID:                dbRows[0].CartID,
		AdminID:           dbRows[0].AdminID,
		RestaurantBrandID: dbRows[0].RestaurantBrandID,
		Mode:              dbRows[0].Mode,
		Status:            dbRows[0].Status,
		UpdatedAt:         dbRows[0].UpdatedAt,
		Items:             []domain.CartItem{},
		Members:           []domain.CartMember{},
	}

	itemCheck := make(map[int64]bool)
	memberCheck := make(map[int64]bool)

	for _, row := range dbRows {
		if row.DishID != nil && !itemCheck[*row.DishID] {
			cart.Items = append(cart.Items, domain.CartItem{
				DishID:      *row.DishID,
				Quantity:    *row.Quantity,
				OwnerUserID: row.OwnerUserID,
			})
			itemCheck[*row.DishID] = true
		}
		if row.MemberUserID != nil && !memberCheck[*row.MemberUserID] {
			cart.Members = append(cart.Members, domain.CartMember{
				UserID:   *row.MemberUserID,
				JoinedAt: *row.JoinedAt,
			})
			memberCheck[*row.MemberUserID] = true
		}
	}

	return cart, nil
}

func (r *cartRepo) UpdateItemQuantity(ctx context.Context, cartID string, dishID int64, quantity int32) error {
	return r.execWithOutbox(ctx, cartID, "ItemUpdated", map[string]any{"dish_id": dishID, "quantity": quantity}, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `UPDATE "cart_dish" SET quantity = $1, updated_at = NOW() WHERE cart_id = $2 AND dish_id = $3`, quantity, cartID, dishID)
		return err
	})
}

func (r *cartRepo) RemoveItem(ctx context.Context, cartID string, dishID int64) error {
	return r.execWithOutbox(ctx, cartID, "ItemRemoved", map[string]any{"dish_id": dishID}, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `DELETE FROM "cart_dish" WHERE cart_id = $1 AND dish_id = $2`, cartID, dishID)
		return err
	})
}

func (r *cartRepo) ReassignItemOwner(ctx context.Context, cartID string, dishID int64, newOwnerID *int64) error {
	return r.execWithOutbox(ctx, cartID, "ItemReassigned", map[string]any{"dish_id": dishID, "new_owner_id": newOwnerID}, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `UPDATE "cart_dish" SET owner_user_id = $1, updated_at = NOW() WHERE cart_id = $2 AND dish_id = $3`, newOwnerID, cartID, dishID)
		return err
	})
}

func (r *cartRepo) OrphanUserItems(ctx context.Context, cartID string, targetUserID int64) error {
	return r.execWithOutbox(ctx, cartID, "ItemsOrphaned", map[string]any{"user_id": targetUserID}, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `UPDATE "cart_dish" SET owner_user_id = NULL WHERE cart_id = $1 AND owner_user_id = $2`, cartID, targetUserID)
		return err
	})
}

func (r *cartRepo) AddMember(ctx context.Context, cartID string, userID int64) error {
	return r.execWithOutbox(ctx, cartID, "MemberJoined", map[string]any{"user_id": userID}, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `INSERT INTO "cart_member" (cart_id, user_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, cartID, userID)
		return err
	})
}

func (r *cartRepo) RemoveMember(ctx context.Context, cartID string, userID int64) error {
	return r.execWithOutbox(ctx, cartID, "MemberKicked", map[string]any{"user_id": userID}, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `DELETE FROM "cart_member" WHERE cart_id = $1 AND user_id = $2`, cartID, userID)
		return err
	})
}

func (r *cartRepo) SaveInvite(ctx context.Context, invite domain.CartInvite) error {
	query := `INSERT INTO "cart_invite" (token, cart_id, expires_at) VALUES ($1, $2, $3)`
	_, err := r.pool.Exec(ctx, query, invite.Token, invite.CartID, invite.ExpiresAt)
	return err
}

func (r *cartRepo) GetInviteByToken(ctx context.Context, token string) (domain.CartInvite, error) {
	var invite domain.CartInvite
	query := `SELECT token, cart_id, expires_at FROM "cart_invite" WHERE token = $1`
	err := r.pool.QueryRow(ctx, query, token).Scan(&invite.Token, &invite.CartID, &invite.ExpiresAt)
	return invite, err
}

func (r *cartRepo) execWithOutbox(ctx context.Context, cartID string, eventType string, payload any, fn func(pgx.Tx) error) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := fn(tx); err != nil {
		return err
	}

	if err := insertOutboxEvent(ctx, tx, cartID, eventType, payload); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
