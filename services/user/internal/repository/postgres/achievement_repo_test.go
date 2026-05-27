package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/user/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/user/internal/repository"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAchievementRepo_ListAll(t *testing.T) {
	ctx := context.Background()

	type mockInit func(m pgxmock.PgxPoolIface)
	tests := []struct {
		name          string
		mockInit      mockInit
		expectedRes   []domain.Achievement
		expectedError error
	}{
		{
			name: "Успешное получение списка",
			mockInit: func(m pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "code", "title", "description", "icon", "sort_order"}).
					AddRow(int64(1), "FIRST_ORDER", "Первый заказ", "Описание 1", "icon1.png", 1).
					AddRow(int64(2), "TEN_ORDERS", "Десять заказов", "Описание 2", "icon2.png", 2)

				m.ExpectQuery(`SELECT id, code, title, description, icon, sort_order FROM "achievement"`).
					WillReturnRows(rows)
			},
			expectedRes: []domain.Achievement{
				{ID: 1, Code: "FIRST_ORDER", Title: "Первый заказ", Description: "Описание 1", Icon: "icon1.png", SortOrder: 1},
				{ID: 2, Code: "TEN_ORDERS", Title: "Десять заказов", Description: "Описание 2", Icon: "icon2.png", SortOrder: 2},
			},
			expectedError: nil,
		},
		{
			name: "Ошибка запроса к БД",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(`SELECT id, code, title, description, icon, sort_order FROM "achievement"`).
					WillReturnError(errors.New("db error"))
			},
			expectedRes:   nil,
			expectedError: errors.New("list achievements: db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			repo := NewAchievementRepo(mock)
			tt.mockInit(mock)

			res, err := repo.ListAll(ctx)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedRes, res)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestAchievementRepo_GetByCode(t *testing.T) {
	ctx := context.Background()
	code := "HERO"

	type mockInit func(m pgxmock.PgxPoolIface)
	tests := []struct {
		name          string
		mockInit      mockInit
		expectedRes   domain.Achievement
		expectedError error
	}{
		{
			name: "Успешное получение по коду",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(`SELECT id, code, title, description, icon, sort_order FROM "achievement" WHERE code = \$1`).
					WithArgs(code).
					WillReturnRows(pgxmock.NewRows([]string{"id", "code", "title", "description", "icon", "sort_order"}).
						AddRow(int64(3), code, "Герой", "Описание", "hero.png", 3))
			},
			expectedRes: domain.Achievement{
				ID:          3,
				Code:        code,
				Title:       "Герой",
				Description: "Описание",
				Icon:        "hero.png",
				SortOrder:   3,
			},
			expectedError: nil,
		},
		{
			name: "Достижение не найдено",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(`SELECT id, code, title, description, icon, sort_order FROM "achievement"`).
					WithArgs(code).
					WillReturnError(pgx.ErrNoRows)
			},
			expectedRes:   domain.Achievement{},
			expectedError: domain.ErrAchievementNotFound,
		},
		{
			name: "Ошибка БД",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(`SELECT id, code, title, description, icon, sort_order FROM "achievement"`).
					WithArgs(code).
					WillReturnError(errors.New("db error"))
			},
			expectedRes:   domain.Achievement{},
			expectedError: errors.New("get achievement by code: db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			repo := NewAchievementRepo(mock)
			tt.mockInit(mock)

			res, err := repo.GetByCode(ctx, code)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedRes, res)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestAchievementRepo_ListForUser(t *testing.T) {
	ctx := context.Background()
	var accountID int64 = 42
	now := time.Now()

	type mockInit func(m pgxmock.PgxPoolIface)
	tests := []struct {
		name          string
		mockInit      mockInit
		expectedRes   []domain.UserAchievement
		expectedError error
	}{
		{
			name: "Успешное получение достижений пользователя",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(`SELECT achievement_id, awarded_at FROM "account_achievement" WHERE account_id = \$1`).
					WithArgs(accountID).
					WillReturnRows(pgxmock.NewRows([]string{"achievement_id", "awarded_at"}).
						AddRow(int64(1), now))
			},
			expectedRes: []domain.UserAchievement{
				{AchievementID: 1, AwardedAt: now},
			},
			expectedError: nil,
		},
		{
			name: "Ошибка БД при получении списка",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(`SELECT achievement_id, awarded_at FROM "account_achievement"`).
					WithArgs(accountID).
					WillReturnError(errors.New("db error"))
			},
			expectedRes:   nil,
			expectedError: errors.New("list user achievements: db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			repo := NewAchievementRepo(mock)
			tt.mockInit(mock)

			res, err := repo.ListForUser(ctx, accountID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedRes, res)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestAchievementRepo_Award(t *testing.T) {
	ctx := context.Background()
	var accountID int64 = 42
	var achievementID int64 = 5

	type mockInit func(m pgxmock.PgxPoolIface)
	tests := []struct {
		name          string
		mockInit      mockInit
		expectedError error
	}{
		{
			name: "Успешная выдача достижения",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectExec(`INSERT INTO "account_achievement"`).
					WithArgs(accountID, achievementID).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
			},
			expectedError: nil,
		},
		{
			name: "Ошибка БД при выдаче",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectExec(`INSERT INTO "account_achievement"`).
					WithArgs(accountID, achievementID).
					WillReturnError(errors.New("db error"))
			},
			expectedError: errors.New("award achievement: db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			repo := NewAchievementRepo(mock)
			tt.mockInit(mock)

			err = repo.Award(ctx, accountID, achievementID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestAchievementRepo_IncrementPaidOrders(t *testing.T) {
	ctx := context.Background()
	var accountID int64 = 42
	paidAt := time.Now()

	type mockInit func(m pgxmock.PgxPoolIface)
	tests := []struct {
		name          string
		mockInit      mockInit
		expectedRes   repository.IncrementPaidOrderResult
		expectedError error
	}{
		{
			name: "Успешный инкремент",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(`WITH cur AS`).
					WithArgs(accountID, paidAt).
					WillReturnRows(pgxmock.NewRows([]string{"paid_orders_count", "streak_count"}).
						AddRow(15, 3))
			},
			expectedRes: repository.IncrementPaidOrderResult{
				NewPaidOrdersCount: 15,
				NewStreakCount:     3,
			},
			expectedError: nil,
		},
		{
			name: "Пользователь не найден",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(`WITH cur AS`).
					WithArgs(accountID, paidAt).
					WillReturnError(pgx.ErrNoRows)
			},
			expectedRes:   repository.IncrementPaidOrderResult{},
			expectedError: domain.ErrUserNotFound,
		},
		{
			name: "Ошибка БД при инкременте",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(`WITH cur AS`).
					WithArgs(accountID, paidAt).
					WillReturnError(errors.New("db error"))
			},
			expectedRes:   repository.IncrementPaidOrderResult{},
			expectedError: errors.New("increment paid orders: db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			repo := NewAchievementRepo(mock)
			tt.mockInit(mock)

			res, err := repo.IncrementPaidOrders(ctx, accountID, paidAt)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedRes, res)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestAchievementRepo_RegisterRestaurantForAccount(t *testing.T) {
	ctx := context.Background()
	var accountID int64 = 42
	var restaurantID int64 = 10
	now := time.Now()

	type mockInit func(m pgxmock.PgxPoolIface)
	tests := []struct {
		name          string
		mockInit      mockInit
		expectedRes   int
		expectedError error
	}{
		{
			name: "Успешная регистрация",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()
				m.ExpectExec(`INSERT INTO "account_restaurant"`).
					WithArgs(accountID, restaurantID, now).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
				m.ExpectQuery(`SELECT COUNT\(\*\) FROM "account_restaurant"`).
					WithArgs(accountID).
					WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(3))
				m.ExpectCommit()
			},
			expectedRes:   3,
			expectedError: nil,
		},
		{
			name: "Ошибка старта транзакции",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin().WillReturnError(errors.New("tx error"))
			},
			expectedRes:   0,
			expectedError: errors.New("begin tx: tx error"),
		},
		{
			name: "Ошибка вставки",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()
				m.ExpectExec(`INSERT INTO "account_restaurant"`).
					WithArgs(accountID, restaurantID, now).
					WillReturnError(errors.New("insert error"))
				m.ExpectRollback()
			},
			expectedRes:   0,
			expectedError: errors.New("register restaurant: insert error"),
		},
		{
			name: "Ошибка подсчета",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()
				m.ExpectExec(`INSERT INTO "account_restaurant"`).
					WithArgs(accountID, restaurantID, now).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
				m.ExpectQuery(`SELECT COUNT\(\*\) FROM "account_restaurant"`).
					WithArgs(accountID).
					WillReturnError(errors.New("count error"))
				m.ExpectRollback()
			},
			expectedRes:   0,
			expectedError: errors.New("count restaurants: count error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			repo := NewAchievementRepo(mock)
			tt.mockInit(mock)

			res, err := repo.RegisterRestaurantForAccount(ctx, accountID, restaurantID, now)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedRes, res)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
