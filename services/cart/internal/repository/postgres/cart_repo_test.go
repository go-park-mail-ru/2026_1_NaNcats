package postgres

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/cart/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCartRepo_CreateCart(t *testing.T) {
	type mockBehavior func(mock pgxmock.PgxPoolIface, adminID int64, brandID int64)

	tests := []struct {
		name         string
		adminID      int64
		brandID      int64
		mockBehavior mockBehavior
		expectedID   string
		expectedErr  error
	}{
		{
			name:    "Успешное создание корзины",
			adminID: 1,
			brandID: 10,
			mockBehavior: func(mock pgxmock.PgxPoolIface, adminID int64, brandID int64) {
				mock.ExpectBegin()

				mock.ExpectQuery(`INSERT INTO "cart"`).
					WithArgs(adminID, brandID).
					WillReturnRows(pgxmock.NewRows([]string{"cart_id"}).AddRow("cart-uuid-123"))

				mock.ExpectExec(`INSERT INTO "cart_member"`).
					WithArgs("cart-uuid-123", adminID).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))

				mock.ExpectCommit()
			},
			expectedID:  "cart-uuid-123",
			expectedErr: nil,
		},
		{
			name:    "Ошибка при вставке корзины (откат транзакции)",
			adminID: 1,
			brandID: 10,
			mockBehavior: func(mock pgxmock.PgxPoolIface, adminID int64, brandID int64) {
				mock.ExpectBegin()

				mock.ExpectQuery(`INSERT INTO "cart"`).
					WithArgs(adminID, brandID).
					WillReturnError(errors.New("db error"))

				mock.ExpectRollback()
			},
			expectedID:  "",
			expectedErr: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			tt.mockBehavior(mock, tt.adminID, tt.brandID)

			repo := NewCartRepo(mock)
			cartID, err := repo.CreateCart(context.Background(), tt.adminID, tt.brandID)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
				assert.Empty(t, cartID)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedID, cartID)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestCartRepo_AddItem(t *testing.T) {
	type mockBehavior func(mock pgxmock.PgxPoolIface, cartID string, item domain.CartItem)

	ownerID := int64(1)
	item := domain.CartItem{DishID: 100, Quantity: 2, OwnerUserID: &ownerID}

	tests := []struct {
		name         string
		cartID       string
		item         domain.CartItem
		mockBehavior mockBehavior
		expectedErr  string
	}{
		{
			name:   "Успешное добавление блюда",
			cartID: "cart-123",
			item:   item,
			mockBehavior: func(mock pgxmock.PgxPoolIface, cartID string, item domain.CartItem) {
				mock.ExpectBegin()

				mock.ExpectQuery(regexp.QuoteMeta(`SELECT status FROM "cart" WHERE cart_id = $1 FOR UPDATE`)).
					WithArgs(cartID).
					WillReturnRows(pgxmock.NewRows([]string{"status"}).AddRow("active"))

				// owner_user_id входит в первичный ключ и передаётся как int64.
				mock.ExpectExec(`INSERT INTO "cart_dish"`).
					WithArgs(cartID, item.DishID, *item.OwnerUserID, item.Quantity).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))

				mock.ExpectExec(`INSERT INTO "outbox_events"`).
					WithArgs(cartID, "CartItemAdded", pgxmock.AnyArg()).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))

				mock.ExpectCommit()
			},
			expectedErr: "",
		},
		{
			name:   "Ошибка: корзина заблокирована (locked)",
			cartID: "cart-locked",
			item:   item,
			mockBehavior: func(mock pgxmock.PgxPoolIface, cartID string, item domain.CartItem) {
				mock.ExpectBegin()

				mock.ExpectQuery(regexp.QuoteMeta(`SELECT status FROM "cart" WHERE cart_id = $1 FOR UPDATE`)).
					WithArgs(cartID).
					WillReturnRows(pgxmock.NewRows([]string{"status"}).AddRow(domain.CartStatusLocked))

				mock.ExpectRollback()
			},
			expectedErr: "cannot add item: cart is locked",
		},
		{
			name:   "Ошибка: корзина не найдена",
			cartID: "cart-404",
			item:   item,
			mockBehavior: func(mock pgxmock.PgxPoolIface, cartID string, item domain.CartItem) {
				mock.ExpectBegin()

				mock.ExpectQuery(regexp.QuoteMeta(`SELECT status FROM "cart" WHERE cart_id = $1 FOR UPDATE`)).
					WithArgs(cartID).
					WillReturnError(pgx.ErrNoRows)

				mock.ExpectRollback()
			},
			expectedErr: "cart not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			tt.mockBehavior(mock, tt.cartID, tt.item)

			repo := NewCartRepo(mock)
			err = repo.AddItem(context.Background(), tt.cartID, tt.item)

			if tt.expectedErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestCartRepo_LockCart(t *testing.T) {
	t.Run("Успешная блокировка с outbox", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		cartID := "cart-123"

		mock.ExpectBegin()

		mock.ExpectExec(`UPDATE "cart" SET status`).
			WithArgs(domain.CartStatusLocked, cartID).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		mock.ExpectExec(`INSERT INTO "outbox_events"`).
			WithArgs(cartID, "CartLocked", pgxmock.AnyArg()).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		mock.ExpectCommit()

		repo := NewCartRepo(mock)
		err = repo.LockCart(context.Background(), cartID)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestCartRepo_GetCartByID(t *testing.T) {
	type mockBehavior func(mock pgxmock.PgxPoolIface, cartID string)

	cols := []string{
		"cart_id", "admin_id", "restaurant_brand_id", "mode", "status", "updated_at",
		"dish_id", "quantity", "owner_user_id", "member_user_id", "joined_at",
	}

	tests := []struct {
		name         string
		cartID       string
		mockBehavior mockBehavior
		expectedErr  error
	}{
		{
			name:   "Успешное получение корзины",
			cartID: "cart-123",
			mockBehavior: func(mock pgxmock.PgxPoolIface, cartID string) {
				now := time.Now()
				var dishID int64 = 100
				var quantity int32 = 2
				var ownerID int64 = 1
				var memberID int64 = 2

				rows := pgxmock.NewRows(cols).
					AddRow("cart-123", int64(1), int64(10), "shared", "active", now, &dishID, &quantity, &ownerID, &memberID, &now).
					AddRow("cart-123", int64(1), int64(10), "shared", "active", now, &dishID, &quantity, &ownerID, &ownerID, &now)

				mock.ExpectQuery(`SELECT c\.cart_id, c\.admin_id, c\.restaurant_brand_id`).
					WithArgs(cartID).
					WillReturnRows(rows)
			},
			expectedErr: nil,
		},
		{
			name:   "Корзина не найдена",
			cartID: "cart-404",
			mockBehavior: func(mock pgxmock.PgxPoolIface, cartID string) {
				mock.ExpectQuery(`SELECT c\.cart_id, c\.admin_id, c\.restaurant_brand_id`).
					WithArgs(cartID).
					WillReturnRows(pgxmock.NewRows(cols))
			},
			expectedErr: domain.ErrDishNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			tt.mockBehavior(mock, tt.cartID)

			repo := NewCartRepo(mock)
			cart, err := repo.GetCartByID(context.Background(), tt.cartID)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
				assert.Empty(t, cart.ID)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.cartID, cart.ID)
				assert.Len(t, cart.Items, 1)
				assert.Len(t, cart.Members, 2)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
