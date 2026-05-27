package postgres

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/cart/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pashagolub/pgxmock/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCartRepo_WithTransaction(t *testing.T) {
	tests := []struct {
		name        string
		mock        func(mock pgxmock.PgxPoolIface)
		fn          func(ctx context.Context) error
		expectedErr string
	}{
		{
			name: "Успешная транзакция",
			mock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectBegin()
				mock.ExpectCommit()
			},
			fn: func(ctx context.Context) error {
				return nil
			},
			expectedErr: "",
		},
		{
			name: "Ошибка при начале транзакции",
			mock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectBegin().WillReturnError(errors.New("begin error"))
			},
			fn: func(ctx context.Context) error {
				return nil
			},
			expectedErr: "begin tx: begin error",
		},
		{
			name: "Ошибка в функции, откат успешен",
			mock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectBegin()
				mock.ExpectRollback()
			},
			fn: func(ctx context.Context) error {
				return errors.New("func error")
			},
			expectedErr: "func error",
		},
		{
			name: "Ошибка в функции и ошибка при откате",
			mock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectBegin()
				mock.ExpectRollback().WillReturnError(errors.New("rollback failed"))
			},
			fn: func(ctx context.Context) error {
				return errors.New("func error")
			},
			expectedErr: "rollback error: rollback failed, original error: func error",
		},
		{
			name: "Ошибка при коммите",
			mock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectBegin()
				mock.ExpectCommit().WillReturnError(errors.New("commit error"))
			},
			fn: func(ctx context.Context) error {
				return nil
			},
			expectedErr: "commit tx: commit error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			tt.mock(mock)

			repo := NewCartRepo(mock)
			err = repo.WithTransaction(context.Background(), tt.fn)

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

func TestCartRepo_CheckAndSaveIdempotency(t *testing.T) {
	tests := []struct {
		name        string
		userID      int64
		key         string
		method      string
		mock        func(mock pgxmock.PgxPoolIface)
		expectedErr error
	}{
		{
			name:   "Успешное сохранение ключа",
			userID: 1,
			key:    "idemp-key-123",
			method: "CreateCart",
			mock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec(`INSERT INTO "idempotency_records"`).
					WithArgs(int64(1), "idemp-key-123", "CreateCart").
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
			},
			expectedErr: nil,
		},
		{
			name:   "Конфликт идемпотентности (23505)",
			userID: 1,
			key:    "idemp-key-123",
			method: "CreateCart",
			mock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec(`INSERT INTO "idempotency_records"`).
					WithArgs(int64(1), "idemp-key-123", "CreateCart").
					WillReturnError(&pgconn.PgError{Code: "23505"})
			},
			expectedErr: domain.ErrIdempotencyConflict,
		},
		{
			name:   "Неизвестная ошибка БД",
			userID: 1,
			key:    "idemp-key-123",
			method: "CreateCart",
			mock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec(`INSERT INTO "idempotency_records"`).
					WithArgs(int64(1), "idemp-key-123", "CreateCart").
					WillReturnError(errors.New("db error"))
			},
			expectedErr: errors.New("insert idempotency record: db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			tt.mock(mock)

			repo := NewCartRepo(mock)
			err = repo.CheckAndSaveIdempotency(context.Background(), tt.userID, tt.key, tt.method)

			if tt.expectedErr != nil {
				if errors.Is(tt.expectedErr, domain.ErrIdempotencyConflict) {
					assert.ErrorIs(t, err, domain.ErrIdempotencyConflict)
				} else {
					assert.Error(t, err)
					assert.Contains(t, err.Error(), tt.expectedErr.Error())
				}
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestCartRepo_CreateCart(t *testing.T) {
	tests := []struct {
		name        string
		adminID     int64
		brandID     int64
		mock        func(mock pgxmock.PgxPoolIface)
		expectedID  string
		expectedErr error
	}{
		{
			name:    "Успешное создание корзины",
			adminID: 1,
			brandID: 10,
			mock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectBegin()
				mock.ExpectQuery(`INSERT INTO "cart"`).
					WithArgs(int64(1), int64(10)).
					WillReturnRows(pgxmock.NewRows([]string{"cart_id"}).AddRow("cart-uuid-123"))
				mock.ExpectExec(`INSERT INTO "cart_member"`).
					WithArgs("cart-uuid-123", int64(1)).
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
			mock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectBegin()
				mock.ExpectQuery(`INSERT INTO "cart"`).
					WithArgs(int64(1), int64(10)).
					WillReturnError(errors.New("db insert error"))
				mock.ExpectRollback()
			},
			expectedID:  "",
			expectedErr: errors.New("db insert error"),
		},
		{
			name:    "Ошибка при вставке администратора (откат транзакции)",
			adminID: 1,
			brandID: 10,
			mock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectBegin()
				mock.ExpectQuery(`INSERT INTO "cart"`).
					WithArgs(int64(1), int64(10)).
					WillReturnRows(pgxmock.NewRows([]string{"cart_id"}).AddRow("cart-uuid-123"))
				mock.ExpectExec(`INSERT INTO "cart_member"`).
					WithArgs("cart-uuid-123", int64(1)).
					WillReturnError(errors.New("member insert error"))
				mock.ExpectRollback()
			},
			expectedID:  "",
			expectedErr: errors.New("member insert error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			tt.mock(mock)

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
	ownerID := int64(1)
	itemWithOwner := domain.CartItem{DishID: 100, Quantity: 2, OwnerUserID: &ownerID}
	itemWithoutOwner := domain.CartItem{DishID: 100, Quantity: 2, OwnerUserID: nil}

	tests := []struct {
		name        string
		cartID      string
		item        domain.CartItem
		mock        func(mock pgxmock.PgxPoolIface)
		expectedErr string
	}{
		{
			name:   "Успешное добавление блюда с владельцем",
			cartID: "cart-123",
			item:   itemWithOwner,
			mock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectBegin()
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT status FROM "cart" WHERE cart_id = $1 FOR UPDATE`)).
					WithArgs("cart-123").
					WillReturnRows(pgxmock.NewRows([]string{"status"}).AddRow("active"))
				mock.ExpectExec(`INSERT INTO "cart_dish"`).
					WithArgs("cart-123", itemWithOwner.DishID, ownerID, itemWithOwner.Quantity).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
				mock.ExpectExec(`INSERT INTO "outbox_events"`).
					WithArgs("cart-123", "CartItemAdded", pgxmock.AnyArg()).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
				mock.ExpectCommit()
			},
			expectedErr: "",
		},
		{
			name:   "Успешное добавление блюда без владельца (nil)",
			cartID: "cart-123",
			item:   itemWithoutOwner,
			mock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectBegin()
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT status FROM "cart" WHERE cart_id = $1 FOR UPDATE`)).
					WithArgs("cart-123").
					WillReturnRows(pgxmock.NewRows([]string{"status"}).AddRow("active"))
				mock.ExpectExec(`INSERT INTO "cart_dish"`).
					WithArgs("cart-123", itemWithoutOwner.DishID, int64(0), itemWithoutOwner.Quantity).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
				mock.ExpectExec(`INSERT INTO "outbox_events"`).
					WithArgs("cart-123", "CartItemAdded", pgxmock.AnyArg()).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
				mock.ExpectCommit()
			},
			expectedErr: "",
		},
		{
			name:   "Ошибка: корзина заблокирована",
			cartID: "cart-locked",
			item:   itemWithOwner,
			mock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectBegin()
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT status FROM "cart" WHERE cart_id = $1 FOR UPDATE`)).
					WithArgs("cart-locked").
					WillReturnRows(pgxmock.NewRows([]string{"status"}).AddRow(domain.CartStatusLocked))
				mock.ExpectRollback()
			},
			expectedErr: "cannot add item: cart is locked",
		},
		{
			name:   "Ошибка: корзина не найдена",
			cartID: "cart-404",
			item:   itemWithOwner,
			mock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectBegin()
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT status FROM "cart" WHERE cart_id = $1 FOR UPDATE`)).
					WithArgs("cart-404").
					WillReturnError(pgx.ErrNoRows)
				mock.ExpectRollback()
			},
			expectedErr: "cart not found",
		},
		{
			name:   "Ошибка при вставке в cart_dish",
			cartID: "cart-123",
			item:   itemWithOwner,
			mock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectBegin()
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT status FROM "cart" WHERE cart_id = $1 FOR UPDATE`)).
					WithArgs("cart-123").
					WillReturnRows(pgxmock.NewRows([]string{"status"}).AddRow("active"))
				mock.ExpectExec(`INSERT INTO "cart_dish"`).
					WithArgs("cart-123", itemWithOwner.DishID, ownerID, itemWithOwner.Quantity).
					WillReturnError(errors.New("insert error"))
				mock.ExpectRollback()
			},
			expectedErr: "insert cart item: insert error",
		},
		{
			name:   "Ошибка при вставке outbox события",
			cartID: "cart-123",
			item:   itemWithOwner,
			mock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectBegin()
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT status FROM "cart" WHERE cart_id = $1 FOR UPDATE`)).
					WithArgs("cart-123").
					WillReturnRows(pgxmock.NewRows([]string{"status"}).AddRow("active"))
				mock.ExpectExec(`INSERT INTO "cart_dish"`).
					WithArgs("cart-123", itemWithOwner.DishID, ownerID, itemWithOwner.Quantity).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
				mock.ExpectExec(`INSERT INTO "outbox_events"`).
					WithArgs("cart-123", "CartItemAdded", pgxmock.AnyArg()).
					WillReturnError(errors.New("outbox error"))
				mock.ExpectRollback()
			},
			expectedErr: "insert outbox event: outbox error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			tt.mock(mock)

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

func TestCartRepo_KickMemberAtomic(t *testing.T) {
	cartID := "cart-123"
	targetUserID := int64(2)

	tests := []struct {
		name        string
		mock        func(mock pgxmock.PgxPoolIface)
		expectedErr string
	}{
		{
			name: "Успешное атомарное исключение участника",
			mock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectBegin()

				mock.ExpectExec(`DELETE FROM "cart_member"`).WithArgs(cartID, targetUserID).WillReturnResult(pgxmock.NewResult("DELETE", 1))
				mock.ExpectExec(`INSERT INTO "outbox_events"`).WithArgs(cartID, "MemberKicked", pgxmock.AnyArg()).WillReturnResult(pgxmock.NewResult("INSERT", 1))

				mock.ExpectExec(`WITH victim AS`).WithArgs(cartID, targetUserID).WillReturnResult(pgxmock.NewResult("INSERT", 2))
				mock.ExpectExec(`INSERT INTO "outbox_events"`).WithArgs(cartID, "ItemsOrphaned", pgxmock.AnyArg()).WillReturnResult(pgxmock.NewResult("INSERT", 1))

				mock.ExpectCommit()
			},
			expectedErr: "",
		},
		{
			name: "Ошибка при исключении (RemoveMember)",
			mock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectBegin()

				mock.ExpectExec(`DELETE FROM "cart_member"`).WithArgs(cartID, targetUserID).WillReturnError(errors.New("remove member error"))
				mock.ExpectRollback()
			},
			expectedErr: "remove member error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			tt.mock(mock)
			repo := NewCartRepo(mock)

			err = repo.KickMemberAtomic(context.Background(), cartID, targetUserID)

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

func TestCartRepo_DowngradeToSolo(t *testing.T) {
	cartID := "cart-123"
	adminID := int64(1)

	tests := []struct {
		name        string
		mock        func(mock pgxmock.PgxPoolIface)
		expectedErr string
	}{
		{
			name: "Успешное понижение до соло",
			mock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectBegin()
				eb := mock.ExpectBatch()
				eb.ExpectExec(`DELETE FROM "cart_member"`).WithArgs(cartID, adminID).WillReturnResult(pgxmock.NewResult("DELETE", 1))
				eb.ExpectExec(`DELETE FROM "cart_dish"`).WithArgs(cartID, adminID).WillReturnResult(pgxmock.NewResult("DELETE", 2))
				eb.ExpectExec(`UPDATE "cart" SET mode`).WithArgs(domain.CartModeSolo, cartID).WillReturnResult(pgxmock.NewResult("UPDATE", 1))
				mock.ExpectExec(`INSERT INTO "outbox_events"`).WithArgs(cartID, "SharedCartClosed", pgxmock.AnyArg()).WillReturnResult(pgxmock.NewResult("INSERT", 1))
				mock.ExpectCommit()
			},
			expectedErr: "",
		},
		{
			name: "Ошибка при выполнении batch-запроса",
			mock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectBegin()
				eb := mock.ExpectBatch()
				// Все три ExpectExec обязаны быть прописаны, иначе SendBatch
				// зафейлится с расхождением размера батча; ошибку возвращаем на первой.
				eb.ExpectExec(`DELETE FROM "cart_member"`).WithArgs(cartID, adminID).WillReturnError(errors.New("batch error"))
				eb.ExpectExec(`DELETE FROM "cart_dish"`).WithArgs(cartID, adminID).WillReturnResult(pgxmock.NewResult("DELETE", 0))
				eb.ExpectExec(`UPDATE "cart" SET mode`).WithArgs(domain.CartModeSolo, cartID).WillReturnResult(pgxmock.NewResult("UPDATE", 0))
				mock.ExpectRollback()
			},
			expectedErr: "downgrade batch step 0 failed: batch error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			tt.mock(mock)
			repo := NewCartRepo(mock)

			err = repo.DowngradeToSolo(context.Background(), cartID, adminID)

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

func TestCartRepo_Getters(t *testing.T) {
	cols := []string{
		"cart_id", "admin_id", "restaurant_brand_id", "mode", "status", "updated_at",
		"dish_id", "quantity", "owner_user_id", "member_user_id", "joined_at",
	}

	now := time.Now()
	var dishID1 int64 = 100
	var quantity1 int32 = 2
	var ownerID1 int64 = 1

	var dishID2 int64 = 101
	var quantity2 int32 = 1
	var ownerIDZero int64 = 0

	var memberID1 int64 = 1
	var memberID2 int64 = 2

	tests := []struct {
		name        string
		method      string
		args        any
		mock        func(mock pgxmock.PgxPoolIface)
		expectedErr error
		validate    func(t *testing.T, cart domain.Cart)
	}{
		{
			name:   "Успешное получение корзины со сложным джойном",
			method: "GetCartByID",
			args:   "cart-123",
			mock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows(cols).
					AddRow("cart-123", int64(1), int64(10), "shared", "active", now, &dishID1, &quantity1, &ownerID1, &memberID1, &now).
					AddRow("cart-123", int64(1), int64(10), "shared", "active", now, &dishID1, &quantity1, &ownerID1, &memberID2, &now).
					AddRow("cart-123", int64(1), int64(10), "shared", "active", now, &dishID2, &quantity2, &ownerIDZero, &memberID1, &now).
					AddRow("cart-123", int64(1), int64(10), "shared", "active", now, &dishID2, &quantity2, &ownerIDZero, &memberID2, &now)

				mock.ExpectQuery(`SELECT c\.cart_id`).WithArgs("cart-123").WillReturnRows(rows)
			},
			expectedErr: nil,
			validate: func(t *testing.T, cart domain.Cart) {
				assert.Equal(t, "cart-123", cart.ID)
				assert.Len(t, cart.Items, 2)
				assert.Len(t, cart.Members, 2)

				hasNilOwner := false
				hasUserOwner := false
				for _, item := range cart.Items {
					if item.OwnerUserID == nil {
						hasNilOwner = true
						assert.Equal(t, dishID2, item.DishID)
					} else if *item.OwnerUserID == ownerID1 {
						hasUserOwner = true
						assert.Equal(t, dishID1, item.DishID)
					}
				}
				assert.True(t, hasNilOwner)
				assert.True(t, hasUserOwner)
			},
		},
		{
			name:   "Корзина не найдена",
			method: "GetCartByID",
			args:   "cart-404",
			mock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery(`SELECT c\.cart_id`).WithArgs("cart-404").WillReturnRows(pgxmock.NewRows(cols))
			},
			expectedErr: domain.ErrDishNotFound,
			validate:    func(t *testing.T, cart domain.Cart) { assert.Empty(t, cart.ID) },
		},
		{
			name:   "Получение корзины по ID пользователя",
			method: "GetCartByUserID",
			args:   int64(1),
			mock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery(`SELECT c\.cart_id FROM "cart" c`).
					WithArgs(int64(1)).
					WillReturnRows(pgxmock.NewRows([]string{"cart_id"}).AddRow("cart-123"))

				mock.ExpectQuery(`SELECT c\.cart_id`).
					WithArgs("cart-123").
					WillReturnRows(pgxmock.NewRows(cols).AddRow("cart-123", int64(1), int64(10), "solo", "active", now, nil, nil, nil, nil, nil))
			},
			expectedErr: nil,
			validate: func(t *testing.T, cart domain.Cart) {
				assert.Equal(t, "cart-123", cart.ID)
				assert.Empty(t, cart.Items)
			},
		},
		{
			name:   "Получение корзины по ID пользователя - нет строк",
			method: "GetCartByUserID",
			args:   int64(1),
			mock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery(`SELECT c\.cart_id FROM "cart" c`).
					WithArgs(int64(1)).
					WillReturnError(pgx.ErrNoRows)
			},
			expectedErr: nil,
			validate:    func(t *testing.T, cart domain.Cart) { assert.Empty(t, cart.ID) },
		},
		{
			name:   "Получение активной корзины пользователя",
			method: "GetActiveCartByUserID",
			args:   int64(1),
			mock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery(`SELECT c\.cart_id FROM "cart" c JOIN "cart_member" cm`).
					WithArgs(int64(1)).
					WillReturnRows(pgxmock.NewRows([]string{"cart_id"}).AddRow("cart-123"))

				mock.ExpectQuery(`SELECT c\.cart_id`).
					WithArgs("cart-123").
					WillReturnRows(pgxmock.NewRows(cols).AddRow("cart-123", int64(1), int64(10), "solo", "active", now, nil, nil, nil, nil, nil))
			},
			expectedErr: nil,
			validate: func(t *testing.T, cart domain.Cart) {
				assert.Equal(t, "cart-123", cart.ID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			tt.mock(mock)
			repo := NewCartRepo(mock)

			var cart domain.Cart
			var callErr error

			switch tt.method {
			case "GetCartByID":
				cart, callErr = repo.GetCartByID(context.Background(), tt.args.(string))
			case "GetCartByUserID":
				cart, callErr = repo.GetCartByUserID(context.Background(), tt.args.(int64))
			case "GetActiveCartByUserID":
				cart, callErr = repo.GetActiveCartByUserID(context.Background(), tt.args.(int64))
			}

			if tt.expectedErr != nil {
				assert.ErrorIs(t, callErr, tt.expectedErr)
			} else {
				assert.NoError(t, callErr)
			}

			if tt.validate != nil {
				tt.validate(t, cart)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestCartRepo_Invites(t *testing.T) {
	invite := domain.CartInvite{
		Token:     "test-token",
		CartID:    "cart-123",
		ExpiresAt: time.Now().Add(time.Hour),
	}

	t.Run("SaveInvite Успех", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		mock.ExpectExec(`INSERT INTO "cart_invite"`).
			WithArgs(invite.Token, invite.CartID, invite.ExpiresAt).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		repo := NewCartRepo(mock)
		err = repo.SaveInvite(context.Background(), invite)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("GetInviteByToken Успех", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		mock.ExpectQuery(`SELECT token, cart_id, expires_at FROM "cart_invite"`).
			WithArgs(invite.Token).
			WillReturnRows(pgxmock.NewRows([]string{"token", "cart_id", "expires_at"}).
				AddRow(invite.Token, invite.CartID, invite.ExpiresAt))

		repo := NewCartRepo(mock)
		res, err := repo.GetInviteByToken(context.Background(), invite.Token)
		assert.NoError(t, err)
		assert.Equal(t, invite.Token, res.Token)
		assert.Equal(t, invite.CartID, res.CartID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestCartRepo_LockCart(t *testing.T) {
	t.Run("Успешная блокировка с outbox", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		cartID := "cart-123"

		// pgxmock реализует и pgxpool.Pool, и pgx.Tx, поэтому execWithOutbox
		// принимает pool за уже открытую транзакцию и пропускает Begin/Commit.
		mock.ExpectExec(`UPDATE "cart" SET status`).
			WithArgs(domain.CartStatusLocked, cartID).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		mock.ExpectExec(`INSERT INTO "outbox_events"`).
			WithArgs(cartID, "CartLocked", pgxmock.AnyArg()).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

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
