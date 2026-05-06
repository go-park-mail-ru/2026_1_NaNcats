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

func TestDishRepo_SearchDishes(t *testing.T) {
	ctx := context.Background()
	searchQuery := "pizza"
	pattern := "%pizza%"
	limit := 10
	now := time.Now()
	columns := []string{"id", "restaurant_brand_id", "name", "description", "image_url", "price", "created_at", "updated_at"}

	type mockInit func(m pgxmock.PgxPoolIface)
	tests := []struct {
		name          string
		mockInit      mockInit
		expectedLen   int
		expectedError string
	}{
		{
			name: "Успешный поиск блюд",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(regexp.QuoteMeta(`WHERE name ILIKE $1 OR description ILIKE $1`)).
					WithArgs(pattern, limit).
					WillReturnRows(pgxmock.NewRows(columns).
						AddRow(int64(1), int64(10), "Pizza Margherita", ptrString("Cheese and tomato"), ptrString("url"), int64(500000000), now, now))
			},
			expectedLen: 1,
		},
		{
			name: "Ошибка выполнения запроса",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
					WithArgs(pattern, limit).
					WillReturnError(errors.New("db error"))
			},
			expectedError: "search dishes: db error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			repo := NewDishRepo(mock)
			tt.mockInit(mock)

			res, err := repo.SearchDishes(ctx, searchQuery, limit)

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

func TestDishRepo_SearchDishesByBrand(t *testing.T) {
	ctx := context.Background()
	brandID := int64(10)
	searchQuery := "burger"
	pattern := "%burger%"
	limit := 5
	now := time.Now()
	columns := []string{"id", "restaurant_brand_id", "name", "description", "image_url", "price", "created_at", "updated_at"}

	type mockInit func(m pgxmock.PgxPoolIface)
	tests := []struct {
		name          string
		mockInit      mockInit
		expectedLen   int
		expectedError string
	}{
		{
			name: "Успешный поиск блюд бренда",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(regexp.QuoteMeta(`WHERE restaurant_brand_id = $1 AND (name ILIKE $2 OR description ILIKE $2)`)).
					WithArgs(brandID, pattern, limit).
					WillReturnRows(pgxmock.NewRows(columns).
						AddRow(int64(1), brandID, "Cheese Burger", ptrString("Double meat"), nil, int64(600000000), now, now))
			},
			expectedLen: 1,
		},
		{
			name: "Ошибка сканирования результатов",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
					WithArgs(brandID, pattern, limit).
					WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow("not-an-int"))
			},
			expectedError: "scan search dishes by brand",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, _ := pgxmock.NewPool()
			defer mock.Close()

			repo := NewDishRepo(mock)
			tt.mockInit(mock)

			res, err := repo.SearchDishesByBrand(ctx, brandID, searchQuery, limit)

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

func TestDishRepo_GetDishesByRestaurantBrandID(t *testing.T) {
	ctx := context.Background()
	brandID := int64(15)
	limit, offset := 20, 0
	now := time.Now()
	columns := []string{"id", "restaurant_brand_id", "name", "description", "image_url", "price", "created_at", "updated_at"}

	type mockInit func(m pgxmock.PgxPoolIface)
	tests := []struct {
		name          string
		mockInit      mockInit
		expectedLen   int
		expectedError string
	}{
		{
			name: "Успешное получение блюд по ID бренда",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(regexp.QuoteMeta(`WHERE restaurant_brand_id = $1 ORDER BY id ASC LIMIT $2 OFFSET $3`)).
					WithArgs(brandID, limit, offset).
					WillReturnRows(pgxmock.NewRows(columns).
						AddRow(int64(101), brandID, "Sushi", nil, ptrString("img.png"), int64(800000000), now, now).
						AddRow(int64(102), brandID, "Rolls", ptrString("Fresh"), nil, int64(700000000), now, now))
			},
			expectedLen: 2,
		},
		{
			name: "Ошибка подключения к базе",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
					WithArgs(brandID, limit, offset).
					WillReturnError(errors.New("connection reset"))
			},
			expectedError: "postgres query error: connection reset",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, _ := pgxmock.NewPool()
			defer mock.Close()

			repo := NewDishRepo(mock)
			tt.mockInit(mock)

			res, err := repo.GetDishesByRestaurantBrandID(ctx, brandID, limit, offset)

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

func TestDishRepo_GetDishByID(t *testing.T) {
	ctx := context.Background()
	var dishID int64 = 100
	now := time.Now()
	columns := []string{"id", "restaurant_brand_id", "name", "description", "image_url", "price", "created_at", "updated_at"}

	type mockInit func(m pgxmock.PgxPoolIface)
	tests := []struct {
		name          string
		mockInit      mockInit
		expectedName  string
		expectedError error
	}{
		{
			name: "Успешное получение блюда",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(regexp.QuoteMeta(`SELECT id, restaurant_brand_id, name, description, image_url, price, created_at, updated_at FROM "dish" WHERE id = $1`)).
					WithArgs(dishID).
					WillReturnRows(pgxmock.NewRows(columns).
						AddRow(dishID, int64(1), "Burger", ptrString("Tasty"), ptrString("url"), int64(500000000), now, now))
			},
			expectedName: "Burger",
		},
		{
			name: "Блюдо не найдено",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(regexp.QuoteMeta(`WHERE id = $1`)).
					WithArgs(dishID).
					WillReturnRows(pgxmock.NewRows(columns))
			},
			expectedError: domain.ErrDishNotFound,
		},
		{
			name: "Ошибка выполнения запроса",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(regexp.QuoteMeta(`WHERE id = $1`)).
					WithArgs(dishID).
					WillReturnError(errors.New("db crash"))
			},
			expectedError: errors.New("query dish by id"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			repo := NewDishRepo(mock)
			tt.mockInit(mock)

			res, err := repo.GetDishByID(ctx, dishID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedName, res.Name)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestDishRepo_GetDishesByIDs(t *testing.T) {
	ctx := context.Background()
	ids := []int64{1, 2, 3}
	now := time.Now()
	columns := []string{"id", "restaurant_brand_id", "name", "description", "image_url", "price", "created_at", "updated_at"}

	type mockInit func(m pgxmock.PgxPoolIface)
	tests := []struct {
		name          string
		mockInit      mockInit
		expectedLen   int
		expectedError string
	}{
		{
			name: "Успешное получение нескольких блюд",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(regexp.QuoteMeta(`WHERE id=ANY($1)`)).
					WithArgs(ids).
					WillReturnRows(pgxmock.NewRows(columns).
						AddRow(int64(1), int64(10), "Dish 1", nil, nil, int64(100), now, now).
						AddRow(int64(2), int64(10), "Dish 2", nil, nil, int64(200), now, now))
			},
			expectedLen: 2,
		},
		{
			name: "Ошибка сканирования строк",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(regexp.QuoteMeta(`WHERE id=ANY($1)`)).
					WithArgs(ids).
					WillReturnError(errors.New("fatal db error"))
			},
			expectedError: "query dishes by ids",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, _ := pgxmock.NewPool()
			defer mock.Close()

			repo := NewDishRepo(mock)
			tt.mockInit(mock)

			res, err := repo.GetDishesByIDs(ctx, ids)

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

func TestDishRepo_Create(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	idemKey := "dish-idem-key"
	dish := domain.Dish{
		RestaurantBrandID: 10,
		Name:              "Burger",
		Description:       "Tasty",
		Price:             500000000,
		ImageURL:          "burger.png",
	}

	columns := []string{"id", "restaurant_brand_id", "name", "description", "price", "image_url", "created_at", "updated_at"}

	type mockInit func(m pgxmock.PgxPoolIface)
	tests := []struct {
		name          string
		mockInit      mockInit
		expectedError error
	}{
		{
			name: "Успешное создание блюда",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "dish" (restaurant_brand_id, name, description, price, image_url, idempotency_key) VALUES ($1, $2, $3, $4, $5, $6)`)).
					WithArgs(dish.RestaurantBrandID, dish.Name, dish.Description, dish.Price, dish.ImageURL, idemKey).
					WillReturnRows(pgxmock.NewRows(columns).
						AddRow(int64(1), dish.RestaurantBrandID, dish.Name, ptrString(dish.Description), dish.Price, ptrString(dish.ImageURL), now, now))
			},
		},
		{
			name: "Ошибка выполнения QueryRow/Scan",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "dish"`)).
					WithArgs(dish.RestaurantBrandID, dish.Name, dish.Description, dish.Price, dish.ImageURL, idemKey).
					WillReturnError(errors.New("insert failed"))
			},
			expectedError: errors.New("insert failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			repo := NewDishRepo(mock)
			tt.mockInit(mock)

			res, err := repo.Create(ctx, dish, idemKey)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, int64(1), res.ID)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestDishRepo_Update(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	dish := domain.Dish{
		ID:                1,
		Name:              "New Name",
		Description:       "New Desc",
		Price:             600000000,
		ImageURL:          "new.png",
		RestaurantBrandID: 10,
	}

	columns := []string{"id", "restaurant_brand_id", "name", "description", "price", "image_url", "created_at", "updated_at"}

	type mockInit func(m pgxmock.PgxPoolIface)
	tests := []struct {
		name          string
		mockInit      mockInit
		expectedError error
	}{
		{
			name: "Успешное обновление",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(regexp.QuoteMeta(`UPDATE "dish" SET name = $1, description = $2, price = $3, image_url = $4, updated_at = NOW() WHERE id = $5`)).
					WithArgs(dish.Name, dish.Description, dish.Price, dish.ImageURL, dish.ID).
					WillReturnRows(pgxmock.NewRows(columns).
						AddRow(dish.ID, dish.RestaurantBrandID, dish.Name, ptrString(dish.Description), dish.Price, ptrString(dish.ImageURL), now, now))
			},
		},
		{
			name: "Ошибка: блюдо не найдено (ErrNoRows)",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(regexp.QuoteMeta(`UPDATE "dish"`)).
					WithArgs(dish.Name, dish.Description, dish.Price, dish.ImageURL, dish.ID).
					WillReturnError(pgx.ErrNoRows)
			},
			expectedError: domain.ErrDishNotFound,
		},
		{
			name: "Внутренняя ошибка базы данных",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(regexp.QuoteMeta(`UPDATE "dish"`)).
					WithArgs(dish.Name, dish.Description, dish.Price, dish.ImageURL, dish.ID).
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

			repo := NewDishRepo(mock)
			tt.mockInit(mock)

			res, err := repo.Update(ctx, dish)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, dish.ID, res.ID)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestDishRepo_Delete(t *testing.T) {
	ctx := context.Background()
	var dishID int64 = 42

	type mockInit func(m pgxmock.PgxPoolIface)
	tests := []struct {
		name     string
		mockInit mockInit
		wantErr  bool
	}{
		{
			name: "Успешное удаление блюда",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectExec(regexp.QuoteMeta(`DELETE FROM "dish" WHERE id = $1`)).
					WithArgs(dishID).
					WillReturnResult(pgxmock.NewResult("DELETE", 1))
			},
			wantErr: false,
		},
		{
			name: "Ошибка при удалении (например, нарушение связей)",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectExec(regexp.QuoteMeta(`DELETE FROM "dish"`)).
					WithArgs(dishID).
					WillReturnError(errors.New("foreign key constraint"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			repo := NewDishRepo(mock)
			tt.mockInit(mock)

			err = repo.Delete(ctx, dishID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
