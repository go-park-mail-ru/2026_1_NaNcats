package postgres

import (
	"context"
	"errors"
	"testing"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v5"
	"github.com/stretchr/testify/assert"
)

func TestClientProfileRepo_Create(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		accountID int
		setup     func(mock pgxmock.PgxPoolIface)
		wantErr   error
	}{
		{
			name:      "Успех",
			accountID: 1,
			setup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec(`INSERT INTO "client_profile"`).
					WithArgs(1).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
			},
			wantErr: nil,
		},
		{
			name:      "Ошибка db",
			accountID: 2,
			setup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec(`INSERT INTO "client_profile"`).
					WithArgs(2).
					WillReturnError(errors.New("conn failed"))
			},
			wantErr: errors.New("conn failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, _ := pgxmock.NewPool()
			defer mock.Close()
			repo := NewClientProfileRepo(mock)

			tt.setup(mock)
			err := repo.Create(ctx, tt.accountID)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestClientProfileRepo_GetByAccountID(t *testing.T) {
	ctx := context.Background()
	errDB := errors.New("db error")

	tests := []struct {
		name      string
		accountID int
		setup     func(mock pgxmock.PgxPoolIface)
		want      domain.ClientProfile
		wantErr   error
	}{
		{
			name:      "Успех",
			accountID: 1,
			setup: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"account_id", "bonus_balance", "streak_count"}).
					AddRow(1, 100, 5)
				mock.ExpectQuery(`SELECT account_id, bonus_balance, streak_count FROM "client_profile"`).
					WithArgs(1).
					WillReturnRows(rows)
			},
			want: domain.ClientProfile{
				AccountID:    1,
				BonusBalance: 100,
				StreakCount:  5,
			},
			wantErr: nil,
		},
		{
			name:      "Ошибка not Found",
			accountID: 404,
			setup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery(`SELECT`).
					WithArgs(404).
					WillReturnError(pgx.ErrNoRows)
			},
			want:    domain.ClientProfile{},
			wantErr: pgx.ErrNoRows,
		},
		{
			name:      "Ошибка internal Error",
			accountID: 500,
			setup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery(`SELECT`).
					WithArgs(500).
					WillReturnError(errDB)
			},
			want:    domain.ClientProfile{},
			wantErr: errDB,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, _ := pgxmock.NewPool()
			defer mock.Close()
			repo := NewClientProfileRepo(mock)

			tt.setup(mock)
			res, err := repo.GetByAccountID(ctx, tt.accountID)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, res)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
