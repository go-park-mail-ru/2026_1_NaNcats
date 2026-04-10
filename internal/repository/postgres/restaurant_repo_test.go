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

func TestRestaurantBrandRepo_GetRestaurantBrandsList(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewRestaurantBrandRepo(mock)
	ctx := context.Background()

	// Вспомогательные переменные для ссылок (т.к. в БД структуре указатели)
	desc1 := "Best burgers"
	logo1 := "logo1.png"

	tests := []struct {
		name    string
		limit   int
		offset  int
		setup   func()
		want    []domain.RestaurantBrand
		wantErr error
	}{
		{
			name:   "Успех с несколькими брендами",
			limit:  2,
			offset: 0,
			setup: func() {
				rows := pgxmock.NewRows([]string{"id", "owner_profile_id", "name", "description", "promotion_tier", "logo_url"}).
					AddRow(1, 101, "Burger King", &desc1, 5, &logo1).
					AddRow(2, 102, "Mac", nil, 3, nil) // Проверка обработки nil

				mock.ExpectQuery(`SELECT id, owner_profile_id, name, description, promotion_tier, logo_url FROM "restaurant_brand"`).
					WithArgs(2, 0).
					WillReturnRows(rows)
			},
			want: []domain.RestaurantBrand{
				{
					ID:             1,
					OwnerProfileID: 101,
					Name:           "Burger King",
					Description:    "Best burgers",
					PromotionTier:  5,
					LogoURL:        "logo1.png",
				},
				{
					ID:             2,
					OwnerProfileID: 102,
					Name:           "Mac",
					Description:    "", // Конвертировалось из nil
					PromotionTier:  3,
					LogoURL:        "", // Конвертировалось из nil
				},
			},
			wantErr: nil,
		},
		{
			name:   "Успех - пустое множество ресторанов",
			limit:  10,
			offset: 0,
			setup: func() {
				mock.ExpectQuery(`SELECT`).
					WithArgs(10, 0).
					WillReturnRows(pgxmock.NewRows([]string{"id", "owner_profile_id", "name", "description", "promotion_tier", "logo_url"}))
			},
			want:    []domain.RestaurantBrand{},
			wantErr: nil,
		},
		{
			name:   "Ошибка БД",
			limit:  10,
			offset: 0,
			setup: func() {
				mock.ExpectQuery(`SELECT`).
					WillReturnError(errors.New("connection failed"))
			},
			want:    nil,
			wantErr: errors.New("connection failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

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
