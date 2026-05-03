package postgres

import (
	"context"
	"errors"
	"regexp"
	"testing"

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
	columns := []string{"account_id", "bonus_balance", "streak_count"}

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
				m.ExpectQuery(regexp.QuoteMeta(`SELECT account_id, bonus_balance, streak_count FROM "client_profile" WHERE account_id = $1`)).
					WithArgs(accountID).
					WillReturnRows(pgxmock.NewRows(columns).AddRow(accountID, int64(1000), 5))
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
