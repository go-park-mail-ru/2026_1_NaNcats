package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/order/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/order/internal/repository"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrderRepo_CreateOrder(t *testing.T) {
	type mockBehavior func(mock pgxmock.PgxPoolIface, order domain.Order, idemKey string)

	ownerID := int64(1)
	pmID := "pm-123"
	var promoID int64 = 99
	var discountAmount int64 = 100
	promoString := "PROMO2026"

	order := domain.Order{
		AdminID:            1,
		RestaurantBranchID: 10,
		RestaurantBrandID:  20,
		RestaurantName:     "Test Rest",
		ClientAddressID:    "addr-uuid",
		TotalCost:          1500,
		PromocodeID:        &promoID,
		DiscountAmount:     discountAmount,
		PromocodeString:    &promoString,
		Status:             "created",
		Items: []domain.OrderDish{
			{DishID: 100, Name: "Burger", Quantity: 2, Price: 500, OwnerUserID: &ownerID},
		},
		Splits: []domain.OrderSplit{
			{ID: "split-1", UserID: 1, BaseAmount: 1500, DiscountAmount: 100, Amount: 1400, PaymentMethodID: &pmID},
		},
	}

	tests := []struct {
		name           string
		order          domain.Order
		idemKey        string
		mockBehavior   mockBehavior
		expectedID     int64
		expectedPubID  string
		expectedErrStr string
	}{
		{
			name:    "Успешное создание заказа (Новый ключ, Транзакция + Батч)",
			order:   order,
			idemKey: "idem-key-1",
			mockBehavior: func(mock pgxmock.PgxPoolIface, order domain.Order, idemKey string) {
				mock.ExpectQuery(`INSERT INTO "idempotency_records"`).
					WithArgs(order.AdminID, idemKey).
					WillReturnRows(pgxmock.NewRows([]string{"response_payload"}).AddRow([]byte(nil)))

				mock.ExpectQuery(`INSERT INTO "order"`).
					WithArgs(
						order.AdminID, order.RestaurantBranchID, order.RestaurantBrandID,
						order.RestaurantName, order.ClientAddressID, order.TotalCost,
						order.PromocodeID, order.DiscountAmount, order.PromocodeString,
						order.Status,
					).
					WillReturnRows(pgxmock.NewRows([]string{"id", "public_id"}).AddRow(int64(42), "pub-uuid-123"))

				b := mock.ExpectBatch()

				pmID := "pm-123"

				// 3. Вставка блюд (owner_user_id передаётся как int64)
				b.ExpectExec(`INSERT INTO "order_dish"`).
					WithArgs(int64(42), int64(100), "Burger", 2, int64(500), int64(1)).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))

				b.ExpectExec(`INSERT INTO "order_split"`).
					WithArgs("split-1", int64(42), int64(1), int64(1500), int64(100), int64(1400), &pmID).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))

				mock.ExpectExec(`UPDATE "idempotency_records"`).
					WithArgs(pgxmock.AnyArg(), order.AdminID, idemKey).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
			},
			expectedID:     42,
			expectedPubID:  "pub-uuid-123",
			expectedErrStr: "",
		},
		{
			name:    "Идемпотентность: запрос уже в процессе (конфликт + пустое тело)",
			order:   order,
			idemKey: "idem-key-progress",
			mockBehavior: func(mock pgxmock.PgxPoolIface, order domain.Order, idemKey string) {
				mock.ExpectQuery(`INSERT INTO "idempotency_records"`).
					WithArgs(order.AdminID, idemKey).
					WillReturnError(pgx.ErrNoRows)

				mock.ExpectQuery(`SELECT response_payload FROM "idempotency_records"`).
					WithArgs(order.AdminID, idemKey).
					WillReturnRows(pgxmock.NewRows([]string{"response_payload"}).AddRow([]byte(nil)))
			},
			expectedID:     0,
			expectedPubID:  "",
			expectedErrStr: "request is already in progress",
		},
		{
			name:    "Идемпотентность: успешный возврат сохраненного результата",
			order:   order,
			idemKey: "idem-key-done",
			mockBehavior: func(mock pgxmock.PgxPoolIface, order domain.Order, idemKey string) {
				mock.ExpectQuery(`INSERT INTO "idempotency_records"`).
					WithArgs(order.AdminID, idemKey).
					WillReturnError(pgx.ErrNoRows)

				mock.ExpectQuery(`SELECT response_payload FROM "idempotency_records"`).
					WithArgs(order.AdminID, idemKey).
					WillReturnRows(pgxmock.NewRows([]string{"response_payload"}).AddRow([]byte(`{"public_id":"pub-uuid-saved"}`)))
			},
			expectedID:     0,
			expectedPubID:  "pub-uuid-saved",
			expectedErrStr: "",
		},
		{
			name:    "Ошибка: падение при вставке заказа",
			order:   order,
			idemKey: "idem-key-err",
			mockBehavior: func(mock pgxmock.PgxPoolIface, order domain.Order, idemKey string) {
				mock.ExpectQuery(`INSERT INTO "idempotency_records"`).
					WithArgs(order.AdminID, idemKey).
					WillReturnRows(pgxmock.NewRows([]string{"response_payload"}).AddRow([]byte(nil)))

				mock.ExpectQuery(`INSERT INTO "order"`).
					WithArgs(
						order.AdminID, order.RestaurantBranchID, order.RestaurantBrandID,
						order.RestaurantName, order.ClientAddressID, order.TotalCost,
						order.PromocodeID, order.DiscountAmount, order.PromocodeString,
						order.Status,
					).
					WillReturnError(errors.New("db error"))
			},
			expectedID:     0,
			expectedPubID:  "",
			expectedErrStr: "insert master order",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			tt.mockBehavior(mock, tt.order, tt.idemKey)

			repo := NewOrderRepo(mock)

			internalID, pubID, err := repo.CreateOrder(context.Background(), tt.order, tt.idemKey)

			if tt.expectedErrStr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrStr)
				assert.Empty(t, pubID)
				assert.Empty(t, internalID)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedID, internalID)
				assert.Equal(t, tt.expectedPubID, pubID)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestOrderRepo_UpdateOrderStatus(t *testing.T) {
	type mockBehavior func(mock pgxmock.PgxPoolIface, publicID string, status string, expectedStatuses ...string)

	tests := []struct {
		name             string
		publicID         string
		status           string
		expectedStatuses []string
		mockBehavior     mockBehavior
		expectedErr      error
	}{
		{
			name:             "Успешное обновление статуса (без optimistic lock)",
			publicID:         "pub-123",
			status:           "paid",
			expectedStatuses: nil,
			mockBehavior: func(mock pgxmock.PgxPoolIface, publicID string, status string, expectedStatuses ...string) {
				mock.ExpectExec(`UPDATE "order" SET status`).
					WithArgs(status, publicID).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
			},
			expectedErr: nil,
		},
		{
			name:             "Успешное обновление статуса (с optimistic lock)",
			publicID:         "pub-123",
			status:           "paid",
			expectedStatuses: []string{"created", "pending"},
			mockBehavior: func(mock pgxmock.PgxPoolIface, publicID string, status string, expectedStatuses ...string) {
				mock.ExpectExec(`UPDATE "order" SET status`).
					WithArgs(status, publicID, expectedStatuses).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
			},
			expectedErr: nil,
		},
		{
			name:             "Ошибка: статус изменен или заказ не найден (ErrStateChanged)",
			publicID:         "pub-404",
			status:           "paid",
			expectedStatuses: []string{"created"},
			mockBehavior: func(mock pgxmock.PgxPoolIface, publicID string, status string, expectedStatuses ...string) {
				mock.ExpectExec(`UPDATE "order" SET status`).
					WithArgs(status, publicID, expectedStatuses).
					WillReturnResult(pgxmock.NewResult("UPDATE", 0))
			},
			expectedErr: repository.ErrStateChanged,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			tt.mockBehavior(mock, tt.publicID, tt.status, tt.expectedStatuses...)

			repo := NewOrderRepo(mock)
			err = repo.UpdateOrderStatus(context.Background(), tt.publicID, tt.status, tt.expectedStatuses...)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedErr)
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

	colsMaster := []string{"id", "public_id", "admin_account_id", "courier_account_id", "restaurant_branch_id", "restaurant_brand_id", "client_address_id", "total_cost", "promocode_id", "discount_amount", "promocode_code", "restaurant_name", "status", "created_at", "updated_at"}
	colsDishes := []string{"dish_id", "dish_name", "quantity", "price", "owner_user_id"}
	colsSplits := []string{"id", "user_id", "base_amount", "discount_amount", "amount", "status", "payment_method_id", "yookassa_payment_id"}

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
				var promoCode *string = nil

				mock.ExpectQuery(`SELECT id, public_id, admin_account_id`).
					WithArgs(publicID).
					WillReturnRows(pgxmock.NewRows(colsMaster).
						AddRow(int64(10), "pub-123", int64(1), courierID, int64(2), int64(3), "addr-1", int64(1500), promoID, int64(0), promoCode, "KFC", "paid", time.Now(), time.Now()))

				var ownerID int64 = 1
				mock.ExpectQuery(`SELECT dish_id, dish_name, quantity, price, owner_user_id FROM "order_dish"`).
					WithArgs(int64(10)).
					WillReturnRows(pgxmock.NewRows(colsDishes).
						AddRow(int64(100), "Burger", 2, int64(500), &ownerID))

				pmID := "pm-1"
				mock.ExpectQuery(`SELECT id, user_id, base_amount, discount_amount, amount, status, payment_method_id, yookassa_payment_id FROM "order_split"`).
					WithArgs(int64(10)).
					WillReturnRows(pgxmock.NewRows(colsSplits).
						AddRow("split-1", int64(1), int64(1500), int64(0), int64(1500), "paid", &pmID, nil))
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
				assert.EqualError(t, err, tt.expectedErr.Error())
				assert.Empty(t, order.ID)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, int64(10), order.ID)
				assert.Len(t, order.Items, 1)
				assert.Equal(t, "Burger", order.Items[0].Name)
				assert.Len(t, order.Splits, 1)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestOrderRepo_GetOrdersByUserID(t *testing.T) {
	type mockBehavior func(mock pgxmock.PgxPoolIface, userID int64, limit, offset int32)

	tests := []struct {
		name         string
		userID       int64
		limit        int32
		offset       int32
		mockBehavior mockBehavior
		expectedErr  error
	}{
		{
			name:   "Успешное получение списка заказов (с пагинацией)",
			userID: 1,
			limit:  10,
			offset: 0,
			mockBehavior: func(mock pgxmock.PgxPoolIface, userID int64, limit, offset int32) {
				var promoCode *string = nil
				mock.ExpectQuery(`SELECT DISTINCT o.id, o.public_id, o.admin_account_id, o.restaurant_branch_id`).
					WithArgs(userID, limit, offset).
					WillReturnRows(pgxmock.NewRows([]string{"id", "public_id", "admin_account_id", "restaurant_branch_id", "restaurant_brand_id", "restaurant_name", "total_cost", "discount_amount", "promocode_code", "status", "created_at"}).
						AddRow(int64(10), "pub-1", int64(1), int64(2), int64(3), "Rest", int64(1500), int64(0), promoCode, "paid", time.Now()))

				mock.ExpectQuery(`SELECT order_id, id, user_id, base_amount, discount_amount, amount, status FROM "order_split"`).
					WithArgs([]int64{10}).
					WillReturnRows(pgxmock.NewRows([]string{"order_id", "id", "user_id", "base_amount", "discount_amount", "amount", "status"}).
						AddRow(int64(10), "split-1", int64(1), int64(1500), int64(0), int64(1500), "paid"))

				var ownerID int64 = 1
				mock.ExpectQuery(`SELECT order_id, dish_id, dish_name, quantity, price, owner_user_id FROM "order_dish"`).
					WithArgs([]int64{10}).
					WillReturnRows(pgxmock.NewRows([]string{"order_id", "dish_id", "dish_name", "quantity", "price", "owner_user_id"}).
						AddRow(int64(10), int64(100), "Burger", 2, int64(500), &ownerID))
			},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			tt.mockBehavior(mock, tt.userID, tt.limit, tt.offset)

			repo := NewOrderRepo(mock)
			res, err := repo.GetOrdersByUserID(context.Background(), tt.userID, tt.limit, tt.offset)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
				assert.Len(t, res, 1)
				assert.Len(t, res[0].Splits, 1)
				assert.Len(t, res[0].Items, 1)
				assert.Equal(t, "Burger", res[0].Items[0].Name)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestOrderRepo_UpdateSplitStatusByPaymentID(t *testing.T) {
	type mockBehavior func(mock pgxmock.PgxPoolIface, yookassaPaymentID, newStatus string)

	tests := []struct {
		name              string
		yookassaPaymentID string
		newStatus         string
		mockBehavior      mockBehavior
		expectedRes       string
		expectedErrStr    string
	}{
		{
			name:              "Успешное обновление статуса",
			yookassaPaymentID: "yoo-123",
			newStatus:         "paid",
			mockBehavior: func(mock pgxmock.PgxPoolIface, yookassaPaymentID, newStatus string) {
				mock.ExpectQuery(`UPDATE "order_split" SET status`).
					WithArgs(newStatus, yookassaPaymentID).
					WillReturnRows(pgxmock.NewRows([]string{"public_id"}).AddRow("pub-order-1"))
			},
			expectedRes:    "pub-order-1",
			expectedErrStr: "",
		},
		{
			name:              "Сплит не найден",
			yookassaPaymentID: "yoo-404",
			newStatus:         "paid",
			mockBehavior: func(mock pgxmock.PgxPoolIface, yookassaPaymentID, newStatus string) {
				mock.ExpectQuery(`UPDATE "order_split" SET status`).
					WithArgs(newStatus, yookassaPaymentID).
					WillReturnError(pgx.ErrNoRows)
			},
			expectedRes:    "",
			expectedErrStr: "split not found by payment ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			tt.mockBehavior(mock, tt.yookassaPaymentID, tt.newStatus)

			repo := NewOrderRepo(mock)
			res, err := repo.UpdateSplitStatusByPaymentID(context.Background(), tt.yookassaPaymentID, tt.newStatus)

			if tt.expectedErrStr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrStr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedRes, res)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestOrderRepo_SetSplitYookassaID(t *testing.T) {
	type mockBehavior func(mock pgxmock.PgxPoolIface, splitID, yookassaID string)

	tests := []struct {
		name           string
		splitID        string
		yookassaID     string
		mockBehavior   mockBehavior
		expectedErrStr string
	}{
		{
			name:       "Успешная установка yookassa id",
			splitID:    "split-1",
			yookassaID: "yoo-123",
			mockBehavior: func(mock pgxmock.PgxPoolIface, splitID, yookassaID string) {
				mock.ExpectExec(`UPDATE "order_split" SET yookassa_payment_id`).
					WithArgs(yookassaID, splitID).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
			},
			expectedErrStr: "",
		},
		{
			name:       "Сплит не найден",
			splitID:    "split-404",
			yookassaID: "yoo-123",
			mockBehavior: func(mock pgxmock.PgxPoolIface, splitID, yookassaID string) {
				mock.ExpectExec(`UPDATE "order_split" SET yookassa_payment_id`).
					WithArgs(yookassaID, splitID).
					WillReturnResult(pgxmock.NewResult("UPDATE", 0))
			},
			expectedErrStr: "split not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			tt.mockBehavior(mock, tt.splitID, tt.yookassaID)

			repo := NewOrderRepo(mock)
			err = repo.SetSplitYookassaID(context.Background(), tt.splitID, tt.yookassaID)

			if tt.expectedErrStr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrStr)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestOrderRepo_GetSplitByID(t *testing.T) {
	type mockBehavior func(mock pgxmock.PgxPoolIface, splitID string)

	tests := []struct {
		name           string
		splitID        string
		mockBehavior   mockBehavior
		expectedErrStr string
	}{
		{
			name:    "Успешное получение",
			splitID: "split-1",
			mockBehavior: func(mock pgxmock.PgxPoolIface, splitID string) {
				pmID := "pm-1"
				yooID := "yoo-1"
				mock.ExpectQuery(`SELECT id, order_id, user_id, amount, status, payment_method_id, yookassa_payment_id`).
					WithArgs(splitID).
					WillReturnRows(pgxmock.NewRows([]string{"id", "order_id", "user_id", "amount", "status", "payment_method_id", "yookassa_payment_id", "created_at", "updated_at"}).
						AddRow("split-1", int64(42), int64(1), int64(100), "pending", &pmID, &yooID, time.Now(), time.Now()))
			},
			expectedErrStr: "",
		},
		{
			name:    "Не найдено",
			splitID: "split-404",
			mockBehavior: func(mock pgxmock.PgxPoolIface, splitID string) {
				mock.ExpectQuery(`SELECT id, order_id`).
					WithArgs(splitID).
					WillReturnError(pgx.ErrNoRows)
			},
			expectedErrStr: "split not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			tt.mockBehavior(mock, tt.splitID)

			repo := NewOrderRepo(mock)
			split, err := repo.GetSplitByID(context.Background(), tt.splitID)

			if tt.expectedErrStr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrStr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.splitID, split.ID)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestOrderRepo_UpdateSplitPayer(t *testing.T) {
	type mockBehavior func(mock pgxmock.PgxPoolIface, splitID string, newPayerID int64)

	tests := []struct {
		name           string
		splitID        string
		newPayerID     int64
		mockBehavior   mockBehavior
		expectedErrStr string
	}{
		{
			name:       "Успешное обновление плательщика",
			splitID:    "split-1",
			newPayerID: 2,
			mockBehavior: func(mock pgxmock.PgxPoolIface, splitID string, newPayerID int64) {
				mock.ExpectExec(`UPDATE "order_split" SET user_id`).
					WithArgs(newPayerID, splitID).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
			},
			expectedErrStr: "",
		},
		{
			name:       "Сплит не найден или уже оплачен",
			splitID:    "split-404",
			newPayerID: 2,
			mockBehavior: func(mock pgxmock.PgxPoolIface, splitID string, newPayerID int64) {
				mock.ExpectExec(`UPDATE "order_split" SET user_id`).
					WithArgs(newPayerID, splitID).
					WillReturnResult(pgxmock.NewResult("UPDATE", 0))
			},
			expectedErrStr: "split not found or already paid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			tt.mockBehavior(mock, tt.splitID, tt.newPayerID)

			repo := NewOrderRepo(mock)
			err = repo.UpdateSplitPayer(context.Background(), tt.splitID, tt.newPayerID)

			if tt.expectedErrStr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrStr)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestOrderRepo_UpdateSplitStatus(t *testing.T) {
	type mockBehavior func(mock pgxmock.PgxPoolIface, splitID, newStatus string)

	tests := []struct {
		name         string
		splitID      string
		newStatus    string
		mockBehavior mockBehavior
		expectedErr  bool
	}{
		{
			name:      "Успешное обновление статуса",
			splitID:   "split-1",
			newStatus: "paid",
			mockBehavior: func(mock pgxmock.PgxPoolIface, splitID, newStatus string) {
				mock.ExpectExec(`UPDATE "order_split" SET status`).
					WithArgs(newStatus, splitID).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
			},
			expectedErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			tt.mockBehavior(mock, tt.splitID, tt.newStatus)

			repo := NewOrderRepo(mock)
			err = repo.UpdateSplitStatus(context.Background(), tt.splitID, tt.newStatus)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestOrderRepo_GetOrdersByStatuses(t *testing.T) {
	type mockBehavior func(mock pgxmock.PgxPoolIface, statuses []string)

	tests := []struct {
		name         string
		statuses     []string
		mockBehavior mockBehavior
		expectedLen  int
		expectedErr  bool
	}{
		{
			name:     "Успешное получение заказов",
			statuses: []string{"created", "pending"},
			mockBehavior: func(mock pgxmock.PgxPoolIface, statuses []string) {
				mock.ExpectQuery(`SELECT id, public_id, admin_account_id, status FROM "order"`).
					WithArgs(statuses).
					WillReturnRows(pgxmock.NewRows([]string{"id", "public_id", "admin_account_id", "status"}).
						AddRow(int64(1), "pub-1", int64(10), "created").
						AddRow(int64(2), "pub-2", int64(10), "pending"))
			},
			expectedLen: 2,
			expectedErr: false,
		},
		{
			name:         "Пустой массив статусов",
			statuses:     []string{},
			mockBehavior: func(mock pgxmock.PgxPoolIface, statuses []string) {},
			expectedLen:  0,
			expectedErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			tt.mockBehavior(mock, tt.statuses)

			repo := NewOrderRepo(mock)
			res, err := repo.GetOrdersByStatuses(context.Background(), tt.statuses)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, res, tt.expectedLen)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestOrderRepo_CheckPromocodeUsage(t *testing.T) {
	type mockBehavior func(mock pgxmock.PgxPoolIface, promoID, userID int64)

	tests := []struct {
		name         string
		promoID      int64
		userID       int64
		mockBehavior mockBehavior
		expectedRes  bool
		expectedErr  bool
	}{
		{
			name:    "Промокод использовался",
			promoID: 1,
			userID:  2,
			mockBehavior: func(mock pgxmock.PgxPoolIface, promoID, userID int64) {
				mock.ExpectQuery(`SELECT EXISTS`).
					WithArgs(promoID, userID).
					WillReturnRows(pgxmock.NewRows([]string{"exists"}).AddRow(true))
			},
			expectedRes: true,
			expectedErr: false,
		},
		{
			name:    "Промокод не использовался",
			promoID: 1,
			userID:  2,
			mockBehavior: func(mock pgxmock.PgxPoolIface, promoID, userID int64) {
				mock.ExpectQuery(`SELECT EXISTS`).
					WithArgs(promoID, userID).
					WillReturnRows(pgxmock.NewRows([]string{"exists"}).AddRow(false))
			},
			expectedRes: false,
			expectedErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			tt.mockBehavior(mock, tt.promoID, tt.userID)

			repo := NewOrderRepo(mock)
			res, err := repo.CheckPromocodeUsage(context.Background(), tt.promoID, tt.userID)

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

func TestOrderRepo_IncrementPromocodeUses(t *testing.T) {
	type mockBehavior func(mock pgxmock.PgxPoolIface, promoID int64)

	tests := []struct {
		name           string
		promoID        int64
		mockBehavior   mockBehavior
		expectedErrStr string
	}{
		{
			name:    "Успешный инкремент",
			promoID: 1,
			mockBehavior: func(mock pgxmock.PgxPoolIface, promoID int64) {
				mock.ExpectExec(`UPDATE "promocode" SET current_uses`).
					WithArgs(promoID).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
			},
			expectedErrStr: "",
		},
		{
			name:    "Промокод не найден",
			promoID: 404,
			mockBehavior: func(mock pgxmock.PgxPoolIface, promoID int64) {
				mock.ExpectExec(`UPDATE "promocode" SET current_uses`).
					WithArgs(promoID).
					WillReturnResult(pgxmock.NewResult("UPDATE", 0))
			},
			expectedErrStr: "promocode not found during increment",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			tt.mockBehavior(mock, tt.promoID)

			repo := NewOrderRepo(mock)
			err = repo.IncrementPromocodeUses(context.Background(), tt.promoID)

			if tt.expectedErrStr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrStr)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestOrderRepo_CreatePromocodeUsage(t *testing.T) {
	type mockBehavior func(mock pgxmock.PgxPoolIface, promoID, orderID, userID int64)

	tests := []struct {
		name         string
		promoID      int64
		orderID      int64
		userID       int64
		mockBehavior mockBehavior
		expectedErr  bool
	}{
		{
			name:    "Успешное сохранение использования",
			promoID: 1,
			orderID: 10,
			userID:  2,
			mockBehavior: func(mock pgxmock.PgxPoolIface, promoID, orderID, userID int64) {
				mock.ExpectExec(`INSERT INTO "promocode_usage"`).
					WithArgs(promoID, orderID, userID).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
			},
			expectedErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			tt.mockBehavior(mock, tt.promoID, tt.orderID, tt.userID)

			repo := NewOrderRepo(mock)
			err = repo.CreatePromocodeUsage(context.Background(), tt.promoID, tt.orderID, tt.userID)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestOrderRepo_RollbackPromocodeUsage(t *testing.T) {
	type mockBehavior func(mock pgxmock.PgxPoolIface, publicID string)

	tests := []struct {
		name           string
		publicID       string
		mockBehavior   mockBehavior
		expectedErrStr string
	}{
		{
			name:     "Успешный откат с декрементом",
			publicID: "pub-1",
			mockBehavior: func(mock pgxmock.PgxPoolIface, publicID string) {
				mock.ExpectBegin()

				var promoID int64 = 99
				mock.ExpectQuery(`SELECT id, promocode_id FROM "order"`).
					WithArgs(publicID).
					WillReturnRows(pgxmock.NewRows([]string{"id", "promocode_id"}).AddRow(int64(10), &promoID))

				mock.ExpectExec(`DELETE FROM "promocode_usage"`).
					WithArgs(int64(10), int64(99)).
					WillReturnResult(pgxmock.NewResult("DELETE", 1))

				mock.ExpectExec(`UPDATE "promocode" SET current_uses`).
					WithArgs(int64(99)).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))

				mock.ExpectCommit()
			},
			expectedErrStr: "",
		},
		{
			name:     "Промокод не применялся (nil)",
			publicID: "pub-2",
			mockBehavior: func(mock pgxmock.PgxPoolIface, publicID string) {
				mock.ExpectBegin()

				var promoID *int64 = nil
				mock.ExpectQuery(`SELECT id, promocode_id FROM "order"`).
					WithArgs(publicID).
					WillReturnRows(pgxmock.NewRows([]string{"id", "promocode_id"}).AddRow(int64(10), promoID))

				mock.ExpectCommit()
			},
			expectedErrStr: "",
		},
		{
			name:     "Заказ не найден",
			publicID: "pub-404",
			mockBehavior: func(mock pgxmock.PgxPoolIface, publicID string) {
				mock.ExpectBegin()

				mock.ExpectQuery(`SELECT id, promocode_id FROM "order"`).
					WithArgs(publicID).
					WillReturnError(pgx.ErrNoRows)

				mock.ExpectRollback()
			},
			expectedErrStr: "failed to fetch order for rollback",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			tt.mockBehavior(mock, tt.publicID)

			repo := NewOrderRepo(mock)
			err = repo.RollbackPromocodeUsage(context.Background(), tt.publicID)

			if tt.expectedErrStr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrStr)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
