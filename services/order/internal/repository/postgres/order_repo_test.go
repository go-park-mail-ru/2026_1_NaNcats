package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/order/internal/domain"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pashagolub/pgxmock/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrderRepo_CreateOrder(t *testing.T) {
	type mockBehavior func(mock pgxmock.PgxPoolIface, order domain.Order, idemKey string)

	ownerID := int64(1)
	pmID := "pm-123"

	order := domain.Order{
		AdminID:            1,
		RestaurantBranchID: 10,
		RestaurantBrandID:  20,
		RestaurantName:     "Test Rest",
		ClientAddressID:    "addr-uuid",
		TotalCost:          1500,
		Status:             "created",
		Items: []domain.OrderDish{
			{DishID: 100, Quantity: 2, Price: 500, OwnerUserID: &ownerID},
		},
		Splits: []domain.OrderSplit{
			{ID: "split-1", UserID: 1, Amount: 1500, PaymentMethodID: &pmID},
		},
	}

	tests := []struct {
		name           string
		order          domain.Order
		idemKey        string
		mockBehavior   mockBehavior
		expectedPubID  string
		expectedErrStr string
	}{
		{
			name:    "Успешное создание заказа (Транзакция + Батч)",
			order:   order,
			idemKey: "idem-key-1",
			mockBehavior: func(mock pgxmock.PgxPoolIface, order domain.Order, idemKey string) {
				mock.ExpectBegin()

				mock.ExpectQuery(`INSERT INTO "order"`).
					WithArgs(
						order.AdminID, order.RestaurantBranchID, order.RestaurantBrandID,
						order.RestaurantName, order.ClientAddressID, order.TotalCost,
						order.Status, idemKey,
					).
					WillReturnRows(pgxmock.NewRows([]string{"id", "public_id"}).AddRow(int64(42), "pub-uuid-123"))

				b := mock.ExpectBatch()

				ownerID := int64(1)
				pmID := "pm-123"

				b.ExpectExec(`INSERT INTO "order_dish"`).
					WithArgs(int64(42), int64(100), 2, int64(500), &ownerID).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))

				b.ExpectExec(`INSERT INTO "order_split"`).
					WithArgs("split-1", int64(42), int64(1), int64(1500), &pmID).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))

				mock.ExpectCommit()
			},
			expectedPubID:  "pub-uuid-123",
			expectedErrStr: "",
		},
		{
			name:    "Ошибка: дубликат по ключу идемпотентности",
			order:   order,
			idemKey: "idem-key-dup",
			mockBehavior: func(mock pgxmock.PgxPoolIface, order domain.Order, idemKey string) {
				mock.ExpectBegin()

				pgErr := &pgconn.PgError{Code: pgerrcode.UniqueViolation}
				mock.ExpectQuery(`INSERT INTO "order"`).
					WithArgs(
						order.AdminID, order.RestaurantBranchID, order.RestaurantBrandID,
						order.RestaurantName, order.ClientAddressID, order.TotalCost,
						order.Status, idemKey,
					).
					WillReturnError(pgErr)

				mock.ExpectRollback()
			},
			expectedPubID:  "",
			expectedErrStr: "order with this idempotency key already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			tt.mockBehavior(mock, tt.order, tt.idemKey)

			repo := NewOrderRepo(mock)
			pubID, err := repo.CreateOrder(context.Background(), tt.order, tt.idemKey)

			if tt.expectedErrStr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrStr)
				assert.Empty(t, pubID)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedPubID, pubID)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestOrderRepo_UpdateOrderStatus(t *testing.T) {
	type mockBehavior func(mock pgxmock.PgxPoolIface, publicID string, status string)

	tests := []struct {
		name         string
		publicID     string
		status       string
		mockBehavior mockBehavior
		expectedErr  error
	}{
		{
			name:     "Успешное обновление статуса",
			publicID: "pub-123",
			status:   "paid",
			mockBehavior: func(mock pgxmock.PgxPoolIface, publicID string, status string) {
				mock.ExpectExec(`UPDATE "order" SET status`).
					WithArgs(status, publicID).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
			},
			expectedErr: nil,
		},
		{
			name:     "Ошибка: заказ не найден",
			publicID: "pub-404",
			status:   "paid",
			mockBehavior: func(mock pgxmock.PgxPoolIface, publicID string, status string) {
				mock.ExpectExec(`UPDATE "order" SET status`).
					WithArgs(status, publicID).
					WillReturnResult(pgxmock.NewResult("UPDATE", 0))
			},
			expectedErr: errors.New("order not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			tt.mockBehavior(mock, tt.publicID, tt.status)

			repo := NewOrderRepo(mock)
			err = repo.UpdateOrderStatus(context.Background(), tt.publicID, tt.status)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestOrderRepo_AreAllSplitsPaid(t *testing.T) {
	type mockBehavior func(mock pgxmock.PgxPoolIface, publicID string)

	tests := []struct {
		name         string
		publicID     string
		mockBehavior mockBehavior
		expectedRes  bool
		expectedErr  bool
	}{
		{
			name:     "Все сплиты оплачены (count = 0)",
			publicID: "pub-1",
			mockBehavior: func(mock pgxmock.PgxPoolIface, publicID string) {
				mock.ExpectQuery(`SELECT COUNT\(\*\)`).
					WithArgs(publicID).
					WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(0))
			},
			expectedRes: true,
			expectedErr: false,
		},
		{
			name:     "Есть неоплаченные сплиты (count = 2)",
			publicID: "pub-2",
			mockBehavior: func(mock pgxmock.PgxPoolIface, publicID string) {
				mock.ExpectQuery(`SELECT COUNT\(\*\)`).
					WithArgs(publicID).
					WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(2))
			},
			expectedRes: false,
			expectedErr: false,
		},
		{
			name:     "Ошибка БД",
			publicID: "pub-err",
			mockBehavior: func(mock pgxmock.PgxPoolIface, publicID string) {
				mock.ExpectQuery(`SELECT COUNT\(\*\)`).
					WithArgs(publicID).
					WillReturnError(errors.New("db disconnect"))
			},
			expectedRes: false,
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			tt.mockBehavior(mock, tt.publicID)

			repo := NewOrderRepo(mock)
			res, err := repo.AreAllSplitsPaid(context.Background(), tt.publicID)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedRes, res)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestOrderRepo_GetOrderByPublicID(t *testing.T) {
	type mockBehavior func(mock pgxmock.PgxPoolIface, publicID string)

	colsMaster := []string{"id", "public_id", "admin_account_id", "courier_account_id", "restaurant_branch_id", "restaurant_brand_id", "client_address_id", "total_cost", "promocode_id", "restaurant_name", "status", "created_at", "updated_at"}
	colsDishes := []string{"dish_id", "quantity", "price", "owner_user_id"}
	colsSplits := []string{"id", "user_id", "amount", "status", "payment_method_id", "yookassa_payment_id"}

	tests := []struct {
		name         string
		publicID     string
		mockBehavior mockBehavior
		expectedErr  error
	}{
		{
			name:     "Успешное получение полного заказа",
			publicID: "pub-123",
			mockBehavior: func(mock pgxmock.PgxPoolIface, publicID string) {
				var courierID *int64 = nil
				var promoID *int64 = nil
				mock.ExpectQuery(`SELECT id, public_id`).
					WithArgs(publicID).
					WillReturnRows(pgxmock.NewRows(colsMaster).
						AddRow(int64(10), "pub-123", int64(1), courierID, int64(2), int64(3), "addr-1", int64(1500), promoID, "KFC", "paid", time.Now(), time.Now()))

				var ownerID int64 = 1
				mock.ExpectQuery(`SELECT dish_id, quantity, price`).
					WithArgs(int64(10)).
					WillReturnRows(pgxmock.NewRows(colsDishes).
						AddRow(int64(100), 2, int64(500), &ownerID))

				pmID := "pm-1"
				mock.ExpectQuery(`SELECT id, user_id, amount`).
					WithArgs(int64(10)).
					WillReturnRows(pgxmock.NewRows(colsSplits).
						AddRow("split-1", int64(1), int64(1500), "paid", &pmID, nil))
			},
			expectedErr: nil,
		},
		{
			name:     "Заказ не найден (pgx.ErrNoRows)",
			publicID: "pub-404",
			mockBehavior: func(mock pgxmock.PgxPoolIface, publicID string) {
				mock.ExpectQuery(`SELECT id, public_id`).
					WithArgs(publicID).
					WillReturnError(pgx.ErrNoRows)
			},
			expectedErr: errors.New("order not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			tt.mockBehavior(mock, tt.publicID)

			repo := NewOrderRepo(mock)
			order, err := repo.GetOrderByPublicID(context.Background(), tt.publicID)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
				assert.Empty(t, order.ID)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, int64(10), order.ID)
				assert.Len(t, order.Items, 1)
				assert.Len(t, order.Splits, 1)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
