package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	"github.com/pashagolub/pgxmock/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDishRepo_GetDishesByRestaurantBrandID(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	brandID := 1
	limit := 10
	offset := 0

	tests := []struct {
		name    string
		setup   func(mock pgxmock.PgxPoolIface)
		want    []domain.Dish
		wantErr bool
	}{
		{
			name: "Успех",
			setup: func(mock pgxmock.PgxPoolIface) {
				columns := []string{"id", "restaurant_brand_id", "name", "description", "image_url", "price", "created_at", "updated_at"}
				rows := pgxmock.NewRows(columns).
					AddRow(1, brandID, "Pizza", nil, nil, int64(500), now, now).
					AddRow(2, brandID, "Pasta", nil, nil, int64(600), now, now)

				mock.ExpectQuery(`SELECT (.+) FROM "dish" WHERE restaurant_brand_id = \$1`).
					WithArgs(brandID, limit, offset).
					WillReturnRows(rows)
			},
			want: []domain.Dish{
				{ID: 1, RestaurantBrandID: brandID, Name: "Pizza", Price: 500, CreatedAt: now, UpdatedAt: now},
				{ID: 2, RestaurantBrandID: brandID, Name: "Pasta", Price: 600, CreatedAt: now, UpdatedAt: now},
			},
			wantErr: false,
		},
		{
			name: "Ошибка бд",
			setup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery(`SELECT`).WithArgs(brandID, limit, offset).WillReturnError(errors.New("db error"))
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			repo := NewDishRepo(mock)
			tt.setup(mock)

			got, err := repo.GetDishesByRestaurantBrandID(ctx, brandID, limit, offset)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestDishRepo_GetDishByID(t *testing.T) {
	ctx := context.Background()
	dishID := 42
	now := time.Now()
	dbErr := errors.New("db error")

	tests := []struct {
		name    string
		setup   func(mock pgxmock.PgxPoolIface)
		want    domain.Dish
		wantErr error
	}{
		{
			name: "Успех",
			setup: func(mock pgxmock.PgxPoolIface) {
				columns := []string{"id", "restaurant_brand_id", "name", "description", "image_url", "price", "created_at", "updated_at"}
				rows := pgxmock.NewRows(columns).
					AddRow(dishID, 1, "Burger", nil, nil, int64(400), now, now)

				mock.ExpectQuery(`SELECT (.+) FROM "dish" WHERE id = \$1`).
					WithArgs(dishID).
					WillReturnRows(rows)
			},
			want:    domain.Dish{ID: dishID, RestaurantBrandID: 1, Name: "Burger", Price: 400, CreatedAt: now, UpdatedAt: now},
			wantErr: nil,
		},
		{
			name: "Ошибка: блюдо не найдено",
			setup: func(mock pgxmock.PgxPoolIface) {
				columns := []string{"id", "restaurant_brand_id", "name", "description", "image_url", "price", "created_at", "updated_at"}
				mock.ExpectQuery(`SELECT`).
					WithArgs(dishID).
					WillReturnRows(pgxmock.NewRows(columns)) // Возвращаем пустые строки
			},
			want:    domain.Dish{},
			wantErr: domain.ErrDishNotFound,
		},
		{
			name: "Ошибка query",
			setup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery(`SELECT`).WithArgs(dishID).WillReturnError(dbErr)
			},
			want:    domain.Dish{},
			wantErr: dbErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			repo := NewDishRepo(mock)
			tt.setup(mock)

			got, err := repo.GetDishByID(ctx, dishID)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
