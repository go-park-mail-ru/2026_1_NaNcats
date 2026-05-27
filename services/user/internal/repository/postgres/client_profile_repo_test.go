package postgres

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/user/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClientProfileRepo_Create(t *testing.T) {
	ctx := context.Background()
	var accountID int64 = 101
	idemKey := "profile-create-key"

	type mockInit func(m pgxmock.PgxPoolIface)
	tests := []struct {
		name     string
		mockInit mockInit
		wantErr  bool
	}{
		{
			name: "Успешное создание профиля",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectExec(regexp.QuoteMeta(`INSERT INTO "client_profile" (account_id, idempotency_key) VALUES ($1, $2)`)).
					WithArgs(accountID, idemKey).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
			},
			wantErr: false,
		},
		{
			name: "Кейс идемпотентности (ON CONFLICT DO UPDATE)",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectExec(regexp.QuoteMeta(`INSERT INTO "client_profile"`)).
					WithArgs(accountID, idemKey).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
			},
			wantErr: false,
		},
		{
			name: "Ошибка базы данных",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectExec(regexp.QuoteMeta(`INSERT INTO "client_profile"`)).
					WithArgs(accountID, idemKey).
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

			repo := NewClientProfileRepo(mock)
			tt.mockInit(mock)

			err = repo.Create(ctx, accountID, idemKey)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestClientProfileRepo_GetByAccountID(t *testing.T) {
	ctx := context.Background()
	var accountID int64 = 42
	columns := []string{"account_id", "bonus_balance", "streak_count", "last_order_date", "streak_freeze_active"}

	type mockInit func(m pgxmock.PgxPoolIface)
	tests := []struct {
		name          string
		mockInit      mockInit
		expectedRes   domain.ClientProfile
		expectedError error
	}{
		{
			name: "Успешное получение профиля",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(regexp.QuoteMeta(`SELECT account_id, bonus_balance, streak_count, last_order_date, streak_freeze_active FROM "client_profile" WHERE account_id = $1`)).
					WithArgs(accountID).
					WillReturnRows(pgxmock.NewRows(columns).AddRow(accountID, int64(1000), 5, nil, false))
			},
			expectedRes: domain.ClientProfile{
				AccountID:    accountID,
				BonusBalance: 1000,
				StreakCount:  5,
			},
		},
		{
			name: "Профиль не найден",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(regexp.QuoteMeta(`SELECT account_id`)).
					WithArgs(accountID).
					WillReturnError(pgx.ErrNoRows)
			},
			expectedError: domain.ErrUserNotFound,
		},
		{
			name: "Внутренняя ошибка базы",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(regexp.QuoteMeta(`SELECT account_id`)).
					WithArgs(accountID).
					WillReturnError(errors.New("conn fail"))
			},
			expectedError: errors.New("conn fail"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, _ := pgxmock.NewPool()
			defer mock.Close()

			repo := NewClientProfileRepo(mock)
			tt.mockInit(mock)

			res, err := repo.GetByAccountID(ctx, accountID)

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

func TestClientProfileRepo_UpdateStreakFreeze(t *testing.T) {
	ctx := context.Background()
	var accountID int64 = 42

	type mockInit func(m pgxmock.PgxPoolIface)
	tests := []struct {
		name          string
		active        bool
		mockInit      mockInit
		expectedError error
	}{
		{
			name:   "Успешная активация заморозки",
			active: true,
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectExec(`UPDATE "client_profile" SET streak_freeze_active = true`).
					WithArgs(accountID).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
			},
			expectedError: nil,
		},
		{
			name:   "Успешная деактивация заморозки",
			active: false,
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectExec(`UPDATE "client_profile" SET streak_freeze_active = false`).
					WithArgs(accountID).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
			},
			expectedError: nil,
		},
		{
			name:   "Профиль не найден",
			active: true,
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectExec(`UPDATE "client_profile"`).
					WithArgs(accountID).
					WillReturnResult(pgxmock.NewResult("UPDATE", 0))
			},
			expectedError: domain.ErrUserNotFound,
		},
		{
			name:   "Ошибка базы данных",
			active: true,
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectExec(`UPDATE "client_profile"`).
					WithArgs(accountID).
					WillReturnError(errors.New("db error"))
			},
			expectedError: errors.New("failed to update streak freeze: db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			repo := NewClientProfileRepo(mock)
			tt.mockInit(mock)

			err = repo.UpdateStreakFreeze(ctx, accountID, tt.active)

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

func TestClientProfileRepo_IncrementStreak(t *testing.T) {
	ctx := context.Background()
	var accountID int64 = 42

	type mockInit func(m pgxmock.PgxPoolIface)
	tests := []struct {
		name          string
		mockInit      mockInit
		expectedError error
	}{
		{
			name: "Успешное увеличение серии",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectExec(`UPDATE "client_profile" SET streak_count = streak_count \+ 1`).
					WithArgs(accountID).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
			},
			expectedError: nil,
		},
		{
			name: "Профиль не найден",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectExec(`UPDATE "client_profile"`).
					WithArgs(accountID).
					WillReturnResult(pgxmock.NewResult("UPDATE", 0))
			},
			expectedError: domain.ErrUserNotFound,
		},
		{
			name: "Ошибка базы данных",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectExec(`UPDATE "client_profile"`).
					WithArgs(accountID).
					WillReturnError(errors.New("db error"))
			},
			expectedError: errors.New("failed to increment streak: db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			repo := NewClientProfileRepo(mock)
			tt.mockInit(mock)

			err = repo.IncrementStreak(ctx, accountID)

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

func TestClientProfileRepo_ResetStreak(t *testing.T) {
	ctx := context.Background()
	var accountID int64 = 42

	type mockInit func(m pgxmock.PgxPoolIface)
	tests := []struct {
		name          string
		mockInit      mockInit
		expectedError error
	}{
		{
			name: "Успешный сброс серии",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectExec(`UPDATE "client_profile" SET streak_count = 0`).
					WithArgs(accountID).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
			},
			expectedError: nil,
		},
		{
			name: "Профиль не найден",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectExec(`UPDATE "client_profile"`).
					WithArgs(accountID).
					WillReturnResult(pgxmock.NewResult("UPDATE", 0))
			},
			expectedError: domain.ErrUserNotFound,
		},
		{
			name: "Ошибка базы данных",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectExec(`UPDATE "client_profile"`).
					WithArgs(accountID).
					WillReturnError(errors.New("db error"))
			},
			expectedError: errors.New("failed to reset streak in db: db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			repo := NewClientProfileRepo(mock)
			tt.mockInit(mock)

			err = repo.ResetStreak(ctx, accountID)

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

func TestClientProfileRepo_ClaimWheelSpin(t *testing.T) {
	ctx := context.Background()
	var accountID int64 = 42
	past25h := time.Now().Add(-25 * time.Hour)
	past12h := time.Now().Add(-12 * time.Hour)

	type mockInit func(m pgxmock.PgxPoolIface)
	tests := []struct {
		name          string
		mockInit      mockInit
		expectedError error
	}{
		{
			name: "Успешная прокрутка, кулдауна не было",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()
				m.ExpectQuery(`SELECT last_wheel_spin_at`).
					WithArgs(accountID).
					WillReturnRows(pgxmock.NewRows([]string{"last_wheel_spin_at"}).AddRow(nil))
				m.ExpectExec(`UPDATE "client_profile" SET last_wheel_spin_at = NOW\(\)`).
					WithArgs(accountID).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
				m.ExpectCommit()
			},
			expectedError: nil,
		},
		{
			name: "Успешная прокрутка, кулдаун прошел",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()
				m.ExpectQuery(`SELECT last_wheel_spin_at`).
					WithArgs(accountID).
					WillReturnRows(pgxmock.NewRows([]string{"last_wheel_spin_at"}).AddRow(&past25h))
				m.ExpectExec(`UPDATE "client_profile"`).
					WithArgs(accountID).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
				m.ExpectCommit()
			},
			expectedError: nil,
		},
		{
			name: "Кулдаун еще активен",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()
				m.ExpectQuery(`SELECT last_wheel_spin_at`).
					WithArgs(accountID).
					WillReturnRows(pgxmock.NewRows([]string{"last_wheel_spin_at"}).AddRow(&past12h))
				m.ExpectRollback()
			},
			expectedError: domain.ErrWheelCooldownActive,
		},
		{
			name: "Профиль не найден",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()
				m.ExpectQuery(`SELECT last_wheel_spin_at`).
					WithArgs(accountID).
					WillReturnError(pgx.ErrNoRows)
				m.ExpectRollback()
			},
			expectedError: domain.ErrUserNotFound,
		},
		{
			name: "Ошибка получения кулдауна",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()
				m.ExpectQuery(`SELECT last_wheel_spin_at`).
					WithArgs(accountID).
					WillReturnError(errors.New("select error"))
				m.ExpectRollback()
			},
			expectedError: errors.New("failed to select last_wheel_spin_at: select error"),
		},
		{
			name: "Ошибка старта транзакции",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin().WillReturnError(errors.New("tx error"))
			},
			expectedError: errors.New("begin transaction: tx error"),
		},
		{
			name: "Ошибка обновления даты",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()
				m.ExpectQuery(`SELECT last_wheel_spin_at`).
					WithArgs(accountID).
					WillReturnRows(pgxmock.NewRows([]string{"last_wheel_spin_at"}).AddRow(nil))
				m.ExpectExec(`UPDATE "client_profile"`).
					WithArgs(accountID).
					WillReturnError(errors.New("update error"))
				m.ExpectRollback()
			},
			expectedError: errors.New("failed to update last_wheel_spin_at: update error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			repo := NewClientProfileRepo(mock)
			tt.mockInit(mock)

			err = repo.ClaimWheelSpin(ctx, accountID)

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

func TestClientProfileRepo_ResetWheelSpinCooldown(t *testing.T) {
	ctx := context.Background()
	var accountID int64 = 42

	type mockInit func(m pgxmock.PgxPoolIface)
	tests := []struct {
		name          string
		mockInit      mockInit
		expectedError error
	}{
		{
			name: "Успешный сброс кулдауна",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectExec(`UPDATE "client_profile" SET last_wheel_spin_at = NULL`).
					WithArgs(accountID).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
			},
			expectedError: nil,
		},
		{
			name: "Профиль не найден",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectExec(`UPDATE "client_profile"`).
					WithArgs(accountID).
					WillReturnResult(pgxmock.NewResult("UPDATE", 0))
			},
			expectedError: domain.ErrUserNotFound,
		},
		{
			name: "Ошибка базы данных",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectExec(`UPDATE "client_profile"`).
					WithArgs(accountID).
					WillReturnError(errors.New("db error"))
			},
			expectedError: errors.New("failed to reset wheel spin cooldown in db: db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			repo := NewClientProfileRepo(mock)
			tt.mockInit(mock)

			err = repo.ResetWheelSpinCooldown(ctx, accountID)

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
