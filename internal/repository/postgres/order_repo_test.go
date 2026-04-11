package postgres

import (
	"context"
	"errors"
	"testing"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v5"
	"github.com/stretchr/testify/assert"
)

func TestOrderRepo_CreateOrder(t *testing.T) {
	ctx := context.Background()
	order := domain.Order{
		ClientID:           1,
		RestaurantBranchID: 10,
		ClientAddressID:    100,
		TotalCost:          2000,
		PaymentMethodID:    "pay-uuid",
		YookassaPaymentID:  "yoo-uuid",
		Status:             "created",
		Items: []domain.OrderDish{
			{DishID: 1, Quantity: 2, Price: 500},
		},
	}

	tests := []struct {
		name    string
		setup   func(mock pgxmock.PgxPoolIface)
		want    string
		wantErr error
	}{
		{
			name: "Успех",
			setup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectBegin()
				mock.ExpectQuery(`INSERT INTO "order"`).
					WithArgs(order.ClientID, order.RestaurantBranchID, order.ClientAddressID, order.TotalCost, order.PaymentMethodID, order.YookassaPaymentID, order.Status).
					WillReturnRows(pgxmock.NewRows([]string{"id", "public_id"}).AddRow(1, "pub-uuid"))

				// Для SendBatch используем ExpectBatch
				batch := mock.ExpectBatch()
				batch.ExpectExec(`INSERT INTO "order_dish"`).
					WithArgs(1, order.Items[0].DishID, order.Items[0].Quantity, order.Items[0].Price).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))

				mock.ExpectCommit()
			},
			want:    "pub-uuid",
			wantErr: nil,
		},
		{
			name: "Ошибка в begin",
			setup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectBegin().WillReturnError(errors.New("tx error"))
			},
			want:    "",
			wantErr: errors.New("tx error"),
		},
		{
			name: "Ошибка insert с rollback",
			setup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectBegin()
				mock.ExpectQuery(`INSERT INTO "order"`).
					WithArgs(order.ClientID, order.RestaurantBranchID, order.ClientAddressID, order.TotalCost, order.PaymentMethodID, order.YookassaPaymentID, order.Status).
					WillReturnError(errors.New("insert error"))
				mock.ExpectRollback()
			},
			want:    "",
			wantErr: errors.New("insert error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, _ := pgxmock.NewPool()
			defer mock.Close()
			repo := NewOrderRepo(mock)

			tt.setup(mock)
			got, err := repo.CreateOrder(ctx, order)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestOrderRepo_UpdateStatusByPaymentID(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		paymentID string
		status    string
		setup     func(mock pgxmock.PgxPoolIface)
		wantErr   error
	}{
		{
			name:      "Успех",
			paymentID: "yoo-1",
			status:    "paid",
			setup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec(`UPDATE "order"`).
					WithArgs("paid", "yoo-1").
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
			},
			wantErr: nil,
		},
		{
			name:      "Ошибка: not found",
			paymentID: "yoo-none",
			status:    "paid",
			setup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec(`UPDATE "order"`).
					WithArgs("paid", "yoo-none").
					WillReturnResult(pgxmock.NewResult("UPDATE", 0))
			},
			wantErr: domain.ErrOrderNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, _ := pgxmock.NewPool()
			defer mock.Close()
			repo := NewOrderRepo(mock)

			tt.setup(mock)
			err := repo.UpdateStatusByPaymentID(ctx, tt.paymentID, tt.status)

			assert.ErrorIs(t, err, tt.wantErr)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestOrderRepo_GetOrderByPublicID(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		publicID string
		userID   int
		setup    func(mock pgxmock.PgxPoolIface)
		wantErr  error
	}{
		{
			name:     "Успех",
			publicID: "pub-1",
			userID:   1,
			setup: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "client_account_id", "courier_account_id", "restaurant_branch_id", "client_address_id", "total_cost", "payment_method_id", "yookassa_payment_id", "status"}).
					AddRow(100, 1, nil, 10, 20, int64(500), "pm-1", "yoo-1", "created")
				mock.ExpectQuery(`SELECT (.+) FROM "order"`).
					WithArgs("pub-1", 1).
					WillReturnRows(rows)
			},
			wantErr: nil,
		},
		{
			name:     "Ошибка: not found",
			publicID: "none",
			userID:   1,
			setup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery(`SELECT`).
					WithArgs("none", 1).
					WillReturnError(pgx.ErrNoRows)
			},
			wantErr: pgx.ErrNoRows,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, _ := pgxmock.NewPool()
			defer mock.Close()
			repo := NewOrderRepo(mock)

			tt.setup(mock)
			_, err := repo.GetOrderByPublicID(ctx, tt.publicID, tt.userID)

			assert.ErrorIs(t, err, tt.wantErr)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
