package postgres

// import (
// 	"context"
// 	"errors"
// 	"testing"
// 	"time"

// 	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
// 	"github.com/pashagolub/pgxmock/v5"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/require"
// )

// func TestCartRepo_GetCartByUserID(t *testing.T) {
// 	ctx := context.Background()
// 	userID := 1
// 	now := time.Now().Truncate(time.Second) // Округляем для тестов

// 	// Вспомогательные переменные для передачи указателей в AddRow
// 	dID1, dID2 := 101, 102
// 	qty1, qty2 := 2, 1
// 	name1, name2 := "Burger", "Cola"
// 	price1, price2 := int64(500), int64(300)
// 	url1, url2 := "img1.png", "img2.png"

// 	// Список колонок должен СТРОГО совпадать с тегами в cartRowDB
// 	columns := []string{"restaurant_brand_id", "updated_at", "dish_id", "quantity", "name", "price", "image_url"}

// 	tests := []struct {
// 		name    string
// 		setup   func(mock pgxmock.PgxPoolIface)
// 		want    domain.Cart
// 		wantErr bool
// 	}{
// 		{
// 			name: "Успех: товары в корзине",
// 			setup: func(mock pgxmock.PgxPoolIface) {
// 				rows := pgxmock.NewRows(columns).
// 					AddRow(10, now, &dID1, &qty1, &name1, &price1, &url1).
// 					AddRow(10, now, &dID2, &qty2, &name2, &price2, &url2)

// 				mock.ExpectQuery(`SELECT (.+) FROM cart c`).
// 					WithArgs(userID).
// 					WillReturnRows(rows)
// 			},
// 			want: domain.Cart{
// 				UserID:            userID,
// 				RestaurantBrandID: 10,
// 				UpdatedAt:         now,
// 				Items: []domain.CartItem{
// 					{DishID: 101, Quantity: 2, Name: "Burger", Price: 500, ImageURL: "img1.png"},
// 					{DishID: 102, Quantity: 1, Name: "Cola", Price: 300, ImageURL: "img2.png"},
// 				},
// 			},
// 			wantErr: false,
// 		},
// 		{
// 			name: "Успех: корзина существует, но она пустая (LEFT JOIN вернул NULL)",
// 			setup: func(mock pgxmock.PgxPoolIface) {
// 				// В этом случае база возвращает 1 строку, где данные корзины есть, а товара - NULL
// 				rows := pgxmock.NewRows(columns).
// 					AddRow(10, now, nil, nil, nil, nil, nil)

// 				mock.ExpectQuery(`SELECT`).
// 					WithArgs(userID).
// 					WillReturnRows(rows)
// 			},
// 			want: domain.Cart{
// 				UserID:            userID,
// 				RestaurantBrandID: 10,
// 				UpdatedAt:         now,
// 				Items:             []domain.CartItem{},
// 			},
// 			wantErr: false,
// 		},
// 		{
// 			name: "Успех: запись о корзине вообще отсутствует в БД",
// 			setup: func(mock pgxmock.PgxPoolIface) {
// 				mock.ExpectQuery(`SELECT`).
// 					WithArgs(userID).
// 					WillReturnRows(pgxmock.NewRows(columns))
// 			},
// 			want: domain.Cart{
// 				UserID: userID,
// 				Items:  []domain.CartItem{},
// 			},
// 			wantErr: false,
// 		},
// 		{
// 			name: "Ошибка: сбой базы данных",
// 			setup: func(mock pgxmock.PgxPoolIface) {
// 				mock.ExpectQuery(`SELECT`).
// 					WithArgs(userID).
// 					WillReturnError(errors.New("db connection failure"))
// 			},
// 			want:    domain.Cart{},
// 			wantErr: true,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			mock, err := pgxmock.NewPool()
// 			require.NoError(t, err)
// 			defer mock.Close()

// 			repo := NewCartRepo(mock)
// 			tt.setup(mock)

// 			got, err := repo.GetCartByUserID(ctx, userID)

// 			if tt.wantErr {
// 				assert.Error(t, err)
// 			} else {
// 				assert.NoError(t, err)
// 				assert.Equal(t, tt.want.UserID, got.UserID)
// 				assert.Equal(t, tt.want.RestaurantBrandID, got.RestaurantBrandID)
// 				assert.Equal(t, tt.want.Items, got.Items)

// 				// Проверка времени с допуском в 1 секунду
// 				if !tt.want.UpdatedAt.IsZero() {
// 					assert.WithinDuration(t, tt.want.UpdatedAt, got.UpdatedAt, time.Second)
// 				}
// 			}
// 			assert.NoError(t, mock.ExpectationsWereMet())
// 		})
// 	}
// }

// func TestCartRepo_UpdateCart(t *testing.T) {
// 	ctx := context.Background()
// 	userID := 1
// 	resID := 10

// 	tests := []struct {
// 		name    string
// 		items   []domain.CartItem
// 		setup   func(mock pgxmock.PgxPoolIface)
// 		wantErr bool
// 	}{
// 		{
// 			name: "Успех",
// 			items: []domain.CartItem{
// 				{DishID: 101, Quantity: 2},
// 				{DishID: 102, Quantity: 1},
// 			},
// 			setup: func(mock pgxmock.PgxPoolIface) {
// 				batch := mock.ExpectBatch()

// 				batch.ExpectExec(`INSERT INTO cart`).
// 					WithArgs(userID, resID).
// 					WillReturnResult(pgxmock.NewResult("INSERT", 1))

// 				batch.ExpectExec(`DELETE FROM cart_dish`).
// 					WithArgs(userID).
// 					WillReturnResult(pgxmock.NewResult("DELETE", 2))

// 				batch.ExpectExec(`INSERT INTO cart_dish`).
// 					WithArgs(userID, []int{101, 102}, []int{2, 1}).
// 					WillReturnResult(pgxmock.NewResult("INSERT", 2))
// 			},
// 			wantErr: false,
// 		},
// 		{
// 			name:  "Успех: пустая корзина",
// 			items: []domain.CartItem{},
// 			setup: func(mock pgxmock.PgxPoolIface) {
// 				batch := mock.ExpectBatch()

// 				batch.ExpectExec(`INSERT INTO cart`).
// 					WithArgs(userID, resID).
// 					WillReturnResult(pgxmock.NewResult("INSERT", 1))

// 				batch.ExpectExec(`DELETE FROM cart_dish`).
// 					WithArgs(userID).
// 					WillReturnResult(pgxmock.NewResult("DELETE", 1))
// 			},
// 			wantErr: false,
// 		},
// 		{
// 			name:  "Ошибка бд: сбой на вставке товаров",
// 			items: []domain.CartItem{{DishID: 101, Quantity: 1}},
// 			setup: func(mock pgxmock.PgxPoolIface) {
// 				batch := mock.ExpectBatch()

// 				batch.ExpectExec(`INSERT INTO cart`).
// 					WithArgs(userID, resID).
// 					WillReturnResult(pgxmock.NewResult("INSERT", 1))

// 				batch.ExpectExec(`DELETE FROM cart_dish`).
// 					WithArgs(userID).
// 					WillReturnResult(pgxmock.NewResult("DELETE", 1))

// 				batch.ExpectExec(`INSERT INTO cart_dish`).
// 					WithArgs(userID, []int{101}, []int{1}).
// 					WillReturnError(errors.New("db write error"))
// 			},
// 			wantErr: true,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			mock, err := pgxmock.NewPool()
// 			require.NoError(t, err)
// 			defer mock.Close()

// 			repo := NewCartRepo(mock)
// 			tt.setup(mock)

// 			err = repo.UpdateCart(ctx, userID, resID, tt.items)
// 			if tt.wantErr {
// 				assert.Error(t, err)
// 			} else {
// 				assert.NoError(t, err)
// 			}
// 			assert.NoError(t, mock.ExpectationsWereMet())
// 		})
// 	}
// }

// func TestCartRepo_ClearCart(t *testing.T) {
// 	ctx := context.Background()
// 	userID := 1

// 	tests := []struct {
// 		name    string
// 		setup   func(mock pgxmock.PgxPoolIface)
// 		wantErr bool
// 	}{
// 		{
// 			name: "Успех",
// 			setup: func(mock pgxmock.PgxPoolIface) {
// 				mock.ExpectExec(`DELETE FROM cart WHERE client_account_id = \$1`).
// 					WithArgs(userID).
// 					WillReturnResult(pgxmock.NewResult("DELETE", 1))
// 			},
// 			wantErr: false,
// 		},
// 		{
// 			name: "Ошибка бд",
// 			setup: func(mock pgxmock.PgxPoolIface) {
// 				mock.ExpectExec(`DELETE FROM cart`).
// 					WithArgs(userID).
// 					WillReturnError(errors.New("fail"))
// 			},
// 			wantErr: true,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			mock, err := pgxmock.NewPool()
// 			require.NoError(t, err)
// 			defer mock.Close()

// 			repo := NewCartRepo(mock)
// 			tt.setup(mock)

// 			err = repo.ClearCart(ctx, userID)
// 			if tt.wantErr {
// 				assert.Error(t, err)
// 			} else {
// 				assert.NoError(t, err)
// 			}
// 			assert.NoError(t, mock.ExpectationsWereMet())
// 		})
// 	}
// }
