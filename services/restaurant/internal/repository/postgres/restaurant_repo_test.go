package restaurant

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/restaurant/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func ptrString(s string) *string { return &s }

func TestRestaurantBrandRepo_GetRestaurantBrandsList(t *testing.T) {
	ctx := context.Background()
	limit, offset := 10, 0
	now := time.Now()
	columns := []string{"id", "owner_profile_id", "name", "description", "promotion_tier", "logo_url", "created_at", "updated_at"}

	type mockInit func(m pgxmock.PgxPoolIface)
	tests := []struct {
		name          string
		mockInit      mockInit
		expectedLen   int
		expectedError string
	}{
		{
			name: "Успешное получение списка",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(regexp.QuoteMeta(`SELECT id, owner_profile_id, name, description, promotion_tier, logo_url, created_at, updated_at FROM "restaurant_brand"`)).
					WithArgs(limit, offset).
					WillReturnRows(pgxmock.NewRows(columns).
						AddRow(int64(1), int64(10), "Brand 1", ptrString("Desc 1"), 3, ptrString("url1"), now, now).
						AddRow(int64(2), int64(10), "Brand 2", nil, 1, nil, now, now))
			},
			expectedLen: 2,
		},
		{
			name: "Ошибка выполнения запроса",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
					WithArgs(limit, offset).
					WillReturnError(errors.New("query error"))
			},
			expectedError: "query restaurant brands list",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			repo := NewRestaurantBrandRepo(mock)
			tt.mockInit(mock)

			res, err := repo.GetRestaurantBrandsList(ctx, limit, offset)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Len(t, res, tt.expectedLen)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestRestaurantBrandRepo_GetByID(t *testing.T) {
	ctx := context.Background()
	var brandID int64 = 1
	now := time.Now()
	columns := []string{"id", "owner_profile_id", "name", "description", "promotion_tier", "logo_url", "created_at", "updated_at"}

	type mockInit func(m pgxmock.PgxPoolIface)
	tests := []struct {
		name          string
		mockInit      mockInit
		expectedName  string
		expectedError error
	}{
		{
			name: "Успешное получение по ID",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(regexp.QuoteMeta(`SELECT id, owner_profile_id, name, description, promotion_tier, logo_url, created_at, updated_at FROM "restaurant_brand" WHERE id = $1`)).
					WithArgs(brandID).
					WillReturnRows(pgxmock.NewRows(columns).
						AddRow(brandID, int64(10), "KFC", ptrString("Fried Chicken"), 2, ptrString("logo.png"), now, now))
			},
			expectedName: "KFC",
		},
		{
			name: "Ресторан не найден",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
					WithArgs(brandID).
					WillReturnError(pgx.ErrNoRows)
			},
			expectedError: domain.ErrRestaurantNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, _ := pgxmock.NewPool()
			defer mock.Close()

			repo := NewRestaurantBrandRepo(mock)
			tt.mockInit(mock)

			res, err := repo.GetByID(ctx, brandID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedName, res.Name)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestRestaurantBrandRepo_GetRestaurantBrandsByIDs(t *testing.T) {
	ctx := context.Background()
	ids := []int64{1, 2}
	now := time.Now()
	columns := []string{"id", "owner_profile_id", "name", "description", "promotion_tier", "logo_url", "created_at", "updated_at"}

	type mockInit func(m pgxmock.PgxPoolIface)
	tests := []struct {
		name          string
		mockInit      mockInit
		expectedLen   int
		expectedError error
	}{
		{
			name: "Успешное пакетное получение",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(regexp.QuoteMeta(`SELECT id, owner_profile_id, name, description, promotion_tier, logo_url, created_at, updated_at FROM "restaurant_brand" WHERE id = ANY($1)`)).
					WithArgs(ids).
					WillReturnRows(pgxmock.NewRows(columns).
						AddRow(int64(1), int64(10), "B1", nil, 1, nil, now, now).
						AddRow(int64(2), int64(10), "B2", nil, 1, nil, now, now))
			},
			expectedLen: 2,
		},
		{
			name: "Частичное совпадение (найден один из двух)",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
					WithArgs(ids).
					WillReturnRows(pgxmock.NewRows(columns).
						AddRow(int64(1), int64(10), "B1", nil, 1, nil, now, now))
			},
			expectedLen: 1,
		},
		{
			name: "Ошибка выполнения запроса",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
					WithArgs(ids).
					WillReturnError(errors.New("db failure"))
			},
			expectedError: errors.New("postgres query error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			repo := NewRestaurantBrandRepo(mock)
			tt.mockInit(mock)

			res, err := repo.GetRestaurantBrandsByIDs(ctx, ids)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Len(t, res, tt.expectedLen)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestRestaurantBrandRepo_Create(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	idemKey := "brand-create-key"
	brand := domain.RestaurantBrand{
		OwnerProfileID: 10,
		Name:           "Burger Heroes",
		Description:    "Best in town",
		PromotionTier:  1,
		LogoURL:        "logo.png",
	}

	type mockInit func(m pgxmock.PgxPoolIface)
	tests := []struct {
		name          string
		mockInit      mockInit
		expectedRes   domain.RestaurantBrand
		expectedError error
	}{
		{
			name: "Успешное создание бренда",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "restaurant_brand"`)).
					WithArgs(brand.OwnerProfileID, brand.Name, brand.Description, brand.PromotionTier, brand.LogoURL, idemKey).
					WillReturnRows(pgxmock.NewRows([]string{"id", "owner_profile_id", "name", "description", "promotion_tier", "logo_url", "created_at", "updated_at"}).
						AddRow(int64(1), brand.OwnerProfileID, brand.Name, ptrString(brand.Description), brand.PromotionTier, ptrString(brand.LogoURL), now, now))
			},
			expectedRes: domain.RestaurantBrand{
				ID:             1,
				OwnerProfileID: brand.OwnerProfileID,
				Name:           brand.Name,
				Description:    brand.Description,
				PromotionTier:  brand.PromotionTier,
				LogoURL:        brand.LogoURL,
				CreatedAt:      now,
				UpdatedAt:      now,
			},
		},
		{
			name: "Ошибка базы данных",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "restaurant_brand"`)).
					WithArgs(brand.OwnerProfileID, brand.Name, brand.Description, brand.PromotionTier, brand.LogoURL, idemKey).
					WillReturnError(errors.New("db fail"))
			},
			expectedError: errors.New("db fail"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			repo := NewRestaurantBrandRepo(mock)
			tt.mockInit(mock)

			res, err := repo.Create(ctx, brand, idemKey)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedRes, res)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestRestaurantBrandRepo_Delete(t *testing.T) {
	ctx := context.Background()
	var brandID int64 = 42

	type mockInit func(m pgxmock.PgxPoolIface)
	tests := []struct {
		name     string
		mockInit mockInit
		wantErr  bool
	}{
		{
			name: "Успешное удаление",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectExec(regexp.QuoteMeta(`DELETE FROM "restaurant_brand" WHERE id = $1`)).
					WithArgs(brandID).
					WillReturnResult(pgxmock.NewResult("DELETE", 1))
			},
		},
		{
			name: "Ошибка при удалении",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectExec(regexp.QuoteMeta(`DELETE FROM "restaurant_brand" WHERE id = $1`)).
					WithArgs(brandID).
					WillReturnError(errors.New("foreign key violation"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, _ := pgxmock.NewPool()
			defer mock.Close()

			repo := NewRestaurantBrandRepo(mock)
			tt.mockInit(mock)

			err := repo.Delete(ctx, brandID)
			assert.Equal(t, tt.wantErr, err != nil)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestRestaurantBrandRepo_Update(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	brand := domain.RestaurantBrand{
		ID:            1,
		Name:          "New Name",
		Description:   "New Desc",
		LogoURL:       "new.png",
		PromotionTier: 3,
	}

	type mockInit func(m pgxmock.PgxPoolIface)
	tests := []struct {
		name          string
		mockInit      mockInit
		expectedError error
	}{
		{
			name: "Успешное обновление",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(regexp.QuoteMeta(`UPDATE "restaurant_brand" SET name = $1, description = $2, logo_url = $3, promotion_tier = $4, updated_at = NOW() WHERE id = $5 RETURNING`)).
					WithArgs(brand.Name, brand.Description, brand.LogoURL, brand.PromotionTier, brand.ID).
					WillReturnRows(pgxmock.NewRows([]string{"id", "owner_profile_id", "name", "description", "promotion_tier", "logo_url", "created_at", "updated_at"}).
						AddRow(brand.ID, int64(10), brand.Name, ptrString(brand.Description), brand.PromotionTier, ptrString(brand.LogoURL), now, now))
			},
		},
		{
			name: "Ресторан не найден",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(regexp.QuoteMeta(`UPDATE "restaurant_brand"`)).
					WithArgs(brand.Name, brand.Description, brand.LogoURL, brand.PromotionTier, brand.ID).
					WillReturnError(pgx.ErrNoRows)
			},
			expectedError: domain.ErrRestaurantNotFound,
		},
		{
			name: "Ошибка базы данных",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(regexp.QuoteMeta(`UPDATE "restaurant_brand"`)).
					WithArgs(brand.Name, brand.Description, brand.LogoURL, brand.PromotionTier, brand.ID).
					WillReturnError(errors.New("db connection error"))
			},
			expectedError: errors.New("db connection error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, _ := pgxmock.NewPool()
			defer mock.Close()

			repo := NewRestaurantBrandRepo(mock)
			tt.mockInit(mock)

			res, err := repo.Update(ctx, brand)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, brand.ID, res.ID)
				assert.Equal(t, brand.Name, res.Name)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestRestaurantBrandRepo_GetRestaurantBrandsByCategory(t *testing.T) {
	ctx := context.Background()
	var categoryID int64 = 5
	limit, offset := 10, 0
	now := time.Now()
	columns := []string{"id", "owner_profile_id", "name", "description", "promotion_tier", "logo_url", "created_at", "updated_at"}

	type mockInit func(m pgxmock.PgxPoolIface)
	tests := []struct {
		name          string
		mockInit      mockInit
		expectedLen   int
		expectedError string
	}{
		{
			name: "Успешное получение по ID категории",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(regexp.QuoteMeta(`SELECT rb.id, rb.owner_profile_id, rb.name, rb.description, rb.promotion_tier, rb.logo_url, rb.created_at, rb.updated_at`)).
					WithArgs(categoryID, limit, offset).
					WillReturnRows(pgxmock.NewRows(columns).
						AddRow(int64(1), int64(10), "Pizza Place", ptrString("Tasty"), 2, ptrString("url"), now, now))
			},
			expectedLen: 1,
		},
		{
			name: "Ошибка выполнения запроса",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
					WithArgs(categoryID, limit, offset).
					WillReturnError(errors.New("db error"))
			},
			expectedError: "query brands by category",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			repo := NewRestaurantBrandRepo(mock)
			tt.mockInit(mock)

			res, err := repo.GetRestaurantBrandsByCategory(ctx, categoryID, limit, offset)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Len(t, res, tt.expectedLen)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestRestaurantBrandRepo_GetRestaurantBrandsByCategoryName(t *testing.T) {
	ctx := context.Background()
	categoryName := "Бургеры"
	limit, offset := 5, 0
	now := time.Now()
	columns := []string{"id", "owner_profile_id", "name", "description", "promotion_tier", "logo_url", "created_at", "updated_at"}

	type mockInit func(m pgxmock.PgxPoolIface)
	tests := []struct {
		name          string
		mockInit      mockInit
		expectedError string
	}{
		{
			name: "Успешное получение по имени категории",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(regexp.QuoteMeta(`LOWER(c.name) = LOWER($1)`)).
					WithArgs(categoryName, limit, offset).
					WillReturnRows(pgxmock.NewRows(columns).
						AddRow(int64(1), int64(10), "Burger King", nil, 3, nil, now, now))
			},
		},
		{
			name: "Ошибка при сканировании строк",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(regexp.QuoteMeta(`LOWER(c.name)`)).
					WithArgs(categoryName, limit, offset).
					WillReturnRows(pgxmock.NewRows([]string{"bad_column"}).AddRow("garbage"))
			},
			expectedError: "scan brands by category name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, _ := pgxmock.NewPool()
			defer mock.Close()

			repo := NewRestaurantBrandRepo(mock)
			tt.mockInit(mock)

			_, err := repo.GetRestaurantBrandsByCategoryName(ctx, categoryName, limit, offset)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestRestaurantBrandRepo_SearchRestaurantBrands(t *testing.T) {
	ctx := context.Background()
	searchQuery := "king"
	pattern := "%king%"
	limit, offset := 10, 0
	now := time.Now()
	columns := []string{"id", "owner_profile_id", "name", "description", "promotion_tier", "logo_url", "created_at", "updated_at"}

	type mockInit func(m pgxmock.PgxPoolIface)
	tests := []struct {
		name          string
		mockInit      mockInit
		expectedLen   int
		expectedError string
	}{
		{
			name: "Успешный поиск",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(regexp.QuoteMeta(`WHERE name ILIKE $1 OR description ILIKE $1`)).
					WithArgs(pattern, limit, offset).
					WillReturnRows(pgxmock.NewRows(columns).
						AddRow(int64(1), int64(10), "Burger King", ptrString("The King"), 3, nil, now, now))
			},
			expectedLen: 1,
		},
		{
			name: "Ничего не найдено",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(regexp.QuoteMeta(`ILIKE`)).
					WithArgs(pattern, limit, offset).
					WillReturnRows(pgxmock.NewRows(columns))
			},
			expectedLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, _ := pgxmock.NewPool()
			defer mock.Close()

			repo := NewRestaurantBrandRepo(mock)
			tt.mockInit(mock)

			res, err := repo.SearchRestaurantBrands(ctx, searchQuery, limit, offset)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Len(t, res, tt.expectedLen)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestRestaurantBrandRepo_GetAllCategories(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	columns := []string{"id", "name", "emoji", "created_at", "updated_at"}

	type mockInit func(m pgxmock.PgxPoolIface)
	tests := []struct {
		name          string
		mockInit      mockInit
		expectedLen   int
		expectedError string
	}{
		{
			name: "Успешное получение всех категорий",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, COALESCE(emoji, '') as emoji, created_at, updated_at FROM "category"`)).
					WillReturnRows(pgxmock.NewRows(columns).
						AddRow(int64(1), "Бургеры", "🍔", now, now).
						AddRow(int64(2), "Пицца", "🍕", now, now))
			},
			expectedLen: 2,
		},
		{
			name: "Успех: категорий нет в базе",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
					WillReturnRows(pgxmock.NewRows(columns))
			},
			expectedLen: 0,
		},
		{
			name: "Ошибка выполнения запроса",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
					WillReturnError(errors.New("db disconnect"))
			},
			expectedError: "get categories: db disconnect",
		},
		{
			name: "Ошибка сканирования (несовпадение типов)",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
					WillReturnRows(pgxmock.NewRows(columns).
						AddRow("not-an-int", "name", "emoji", now, now))
			},
			expectedError: "scan categories",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			repo := NewRestaurantBrandRepo(mock)
			tt.mockInit(mock)

			res, err := repo.GetAllCategories(ctx)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Len(t, res, tt.expectedLen)
				if tt.expectedLen > 0 {
					assert.Equal(t, "Бургеры", res[0].Name)
					assert.Equal(t, "🍔", res[0].Emoji)
				}
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
