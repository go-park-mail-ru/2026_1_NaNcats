package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	"github.com/pashagolub/pgxmock/v5"
	"github.com/stretchr/testify/assert"
)

func TestRestaurantBrandRepo_GetRestaurantBrandsList(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	desc1 := "Best burgers"
	logo1 := "logo1.png"

	tests := []struct {
		name    string
		limit   int
		offset  int
		setup   func(mock pgxmock.PgxPoolIface)
		want    []domain.RestaurantBrand
		wantErr error
	}{
		{
			name:   "Успех с несколькими брендами",
			limit:  2,
			offset: 0,
			setup: func(mock pgxmock.PgxPoolIface) {
				columns := []string{"id", "owner_profile_id", "name", "description", "promotion_tier", "logo_url", "created_at", "updated_at"}
				rows := pgxmock.NewRows(columns).
					AddRow(1, 101, "Burger King", &desc1, 5, &logo1, now, now).
					AddRow(2, 102, "Mac", nil, 3, nil, now, now)

				// Используем (.+), чтобы не прописывать весь длинный SQL вручную
				mock.ExpectQuery(`SELECT (.+) FROM "restaurant_brand" (.+) LIMIT \$1 OFFSET \$2`).
					WithArgs(2, 0).
					WillReturnRows(rows)
			},
			want: []domain.RestaurantBrand{
				{ID: 1, OwnerProfileID: 101, Name: "Burger King", Description: "Best burgers", PromotionTier: 5, LogoURL: "logo1.png", CreatedAt: now, UpdatedAt: now},
				{ID: 2, OwnerProfileID: 102, Name: "Mac", Description: "", PromotionTier: 3, LogoURL: "", CreatedAt: now, UpdatedAt: now},
			},
			wantErr: nil,
		},
		{
			name:   "Успех - пустое множество ресторанов",
			limit:  10,
			offset: 0,
			setup: func(mock pgxmock.PgxPoolIface) {
				columns := []string{"id", "owner_profile_id", "name", "description", "promotion_tier", "logo_url", "created_at", "updated_at"}
				mock.ExpectQuery(`SELECT (.+) FROM "restaurant_brand"`).
					WithArgs(10, 0).
					WillReturnRows(pgxmock.NewRows(columns))
			},
			want:    []domain.RestaurantBrand{},
			wantErr: nil,
		},
		{
			name:   "Ошибка БД",
			limit:  10,
			offset: 0,
			setup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery(`SELECT (.+) FROM "restaurant_brand"`).
					WithArgs(10, 0).
					WillReturnError(errors.New("connection failed"))
			},
			want:    nil,
			wantErr: errors.New("connection failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, _ := pgxmock.NewPool()
			defer mock.Close()
			repo := NewRestaurantBrandRepo(mock)

			tt.setup(mock)

			got, err := repo.GetRestaurantBrandsList(ctx, tt.limit, tt.offset)

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
