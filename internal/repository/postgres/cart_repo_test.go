package postgres

import (
	"context"
	"errors"
	"testing"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	"github.com/pashagolub/pgxmock/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCartRepo_GetCartByUserID(t *testing.T) {
	ctx := context.Background()
	userID := 1

	tests := []struct {
		name    string
		setup   func(mock pgxmock.PgxPoolIface)
		want    domain.Cart
		wantErr bool
	}{
		{
			name: "Успех",
			setup: func(mock pgxmock.PgxPoolIface) {
				columns := []string{"res_id", "dish_id", "qty", "name", "price", "url"}
				rows := pgxmock.NewRows(columns).
					AddRow(10, 101, 2, "Burger", int64(500), "img1.png").
					AddRow(10, 102, 1, "Cola", int64(300), "img2.png")

				mock.ExpectQuery(`SELECT (.+) FROM cart c`).
					WithArgs(userID).
					WillReturnRows(rows)
			},
			want: domain.Cart{
				UserID:            userID,
				RestaurantBrandID: 10,
				Items: []domain.CartItem{
					{DishID: 101, Quantity: 2, Name: "Burger", Price: 500, ImageURL: "img1.png"},
					{DishID: 102, Quantity: 1, Name: "Cola", Price: 300, ImageURL: "img2.png"},
				},
			},
			wantErr: false,
		},
		{
			name: "Успех: корзина пустая",
			setup: func(mock pgxmock.PgxPoolIface) {
				columns := []string{"res_id", "dish_id", "qty", "name", "price", "url"}
				mock.ExpectQuery(`SELECT`).
					WithArgs(userID).
					WillReturnRows(pgxmock.NewRows(columns))
			},
			want: domain.Cart{
				UserID:            0,
				RestaurantBrandID: 0,
				Items:             []domain.CartItem{},
			},
			wantErr: false,
		},
		{
			name: "Ошибка бд",
			setup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery(`SELECT`).
					WithArgs(userID).
					WillReturnError(errors.New("db error"))
			},
			want:    domain.Cart{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			repo := NewCartRepo(mock)
			tt.setup(mock)

			got, err := repo.GetCartByUserID(ctx, userID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want.UserID, got.UserID)
				assert.Equal(t, tt.want.RestaurantBrandID, got.RestaurantBrandID)
				assert.Equal(t, tt.want.Items, got.Items)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestCartRepo_UpdateCart(t *testing.T) {
	ctx := context.Background()
	userID := 1
	resID := 10

	tests := []struct {
		name    string
		items   []domain.CartItem
		setup   func(mock pgxmock.PgxPoolIface)
		wantErr bool
	}{
		{
			name: "Успех",
			items: []domain.CartItem{
				{DishID: 101, Quantity: 2},
				{DishID: 102, Quantity: 1},
			},
			setup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec(`WITH up_cart AS`).
					WithArgs(userID, resID, []int{101, 102}, []int{2, 1}).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
			},
			wantErr: false,
		},
		{
			name:  "Успех: пустая корзина",
			items: []domain.CartItem{},
			setup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec(`DELETE FROM cart_dish WHERE cart_id = \$1`).
					WithArgs(userID).
					WillReturnResult(pgxmock.NewResult("DELETE", 1))
			},
			wantErr: false,
		},
		{
			name:  "Ошибка бд",
			items: []domain.CartItem{{DishID: 101, Quantity: 1}},
			setup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec(`WITH up_cart AS`).
					WithArgs(userID, resID, []int{101}, []int{1}).
					WillReturnError(errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			repo := NewCartRepo(mock)
			tt.setup(mock)

			err = repo.UpdateCart(ctx, userID, resID, tt.items)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestCartRepo_ClearCart(t *testing.T) {
	ctx := context.Background()
	userID := 1

	tests := []struct {
		name    string
		setup   func(mock pgxmock.PgxPoolIface)
		wantErr bool
	}{
		{
			name: "Успех",
			setup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec(`DELETE FROM cart WHERE client_account_id = \$1`).
					WithArgs(userID).
					WillReturnResult(pgxmock.NewResult("DELETE", 1))
			},
			wantErr: false,
		},
		{
			name: "Ошибка бд",
			setup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec(`DELETE FROM cart`).
					WithArgs(userID).
					WillReturnError(errors.New("fail"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			repo := NewCartRepo(mock)
			tt.setup(mock)

			err = repo.ClearCart(ctx, userID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
