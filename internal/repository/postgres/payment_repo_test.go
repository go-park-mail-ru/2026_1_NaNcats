package postgres

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pashagolub/pgxmock/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPaymentRepo_Create(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewPaymentRepo(mock)
	ctx := context.Background()

	method := domain.PaymentMethod{
		UserID:     1,
		ExternalID: "card_123",
		CardType:   "visa",
		Last4:      "4242",
		IssuerName: "Sber",
		IsDefault:  true,
	}

	tests := []struct {
		name    string
		setup   func()
		wantID  int
		wantErr error
	}{
		{
			name: "Успех",
			setup: func() {
				mock.ExpectQuery(`INSERT INTO "payment_method"`).
					WithArgs(method.UserID, method.ExternalID, method.CardType, method.Last4, method.IssuerName, method.IsDefault).
					WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(10))
			},
			wantID:  10,
			wantErr: nil,
		},
		{
			name: "Ошибка: payment method уже существует",
			setup: func() {
				mock.ExpectQuery(`INSERT INTO "payment_method"`).
					WithArgs(method.UserID, method.ExternalID, method.CardType, method.Last4, method.IssuerName, method.IsDefault).
					WillReturnError(&pgconn.PgError{Code: pgerrcode.UniqueViolation})
			},
			wantID:  0,
			wantErr: domain.ErrPaymentMethodAlreadyExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			id, err := repo.Create(ctx, method)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.wantID, id)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestPaymentRepo_Delete(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewPaymentRepo(mock)
	ctx := context.Background()

	tests := []struct {
		name    string
		cardID  string
		userID  int
		setup   func()
		wantErr error
	}{
		{
			name:   "Успешное удаление payment method",
			cardID: "card_1",
			userID: 1,
			setup: func() {
				mock.ExpectExec(`DELETE FROM "payment_method" WHERE external_id = \$1 AND user_id = \$2`).
					WithArgs("card_1", 1).
					WillReturnResult(pgxmock.NewResult("DELETE", 1))
			},
			wantErr: nil,
		},
		{
			name:   "Ошибка: не найден payment method",
			cardID: "card_999",
			userID: 1,
			setup: func() {
				mock.ExpectExec(`DELETE FROM "payment_method"`).
					WithArgs("card_999", 1).
					WillReturnResult(pgxmock.NewResult("DELETE", 0))
			},
			wantErr: domain.ErrPaymentMethodNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			err := repo.Delete(ctx, tt.cardID, tt.userID)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestPaymentRepo_GetByUserID(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewPaymentRepo(mock)
	ctx := context.Background()

	issuer := "Tinkoff"

	tests := []struct {
		name    string
		userID  int
		setup   func()
		want    []domain.PaymentMethod
		wantErr bool
	}{
		{
			name:   "Успех - карты найдены",
			userID: 1,
			setup: func() {
				rows := pgxmock.NewRows([]string{"id", "user_id", "external_id", "card_type", "last4", "issuer_name", "is_default"}).
					AddRow(1, 1, "ext_1", "visa", "1111", &issuer, true).
					AddRow(2, 1, "ext_2", "mastercard", "2222", nil, false)

				mock.ExpectQuery(`SELECT (.+) FROM "payment_method" WHERE user_id = \$1`).
					WithArgs(1).
					WillReturnRows(rows)
			},
			want: []domain.PaymentMethod{
				{ID: 1, UserID: 1, ExternalID: "ext_1", CardType: "visa", Last4: "1111", IssuerName: "Tinkoff", IsDefault: true},
				{ID: 2, UserID: 1, ExternalID: "ext_2", CardType: "mastercard", Last4: "2222", IssuerName: "", IsDefault: false},
			},
			wantErr: false,
		},
		{
			name:   "Ошибка query",
			userID: 1,
			setup: func() {
				mock.ExpectQuery(`SELECT`).
					WithArgs(1).
					WillReturnError(errors.New("db error"))
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			got, err := repo.GetByUserID(ctx, tt.userID)
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

func TestPaymentRepo_SetDefault(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewPaymentRepo(mock)
	ctx := context.Background()

	tests := []struct {
		name    string
		cardID  string
		userID  int
		setup   func()
		wantErr error
	}{
		{
			name:   "Успех",
			cardID: "new_default",
			userID: 1,
			setup: func() {
				mock.ExpectExec(`UPDATE "payment_method"`).
					WithArgs("new_default", 1).
					WillReturnResult(pgxmock.NewResult("UPDATE", 2)) // обновилось 2 строки: старая и новая карты
			},
			wantErr: nil,
		},
		{
			name:   "Ошибка: payment method не найден",
			cardID: "ghost_card",
			userID: 1,
			setup: func() {
				mock.ExpectExec(`UPDATE "payment_method"`).
					WithArgs("ghost_card", 1).
					WillReturnResult(pgxmock.NewResult("UPDATE", 0))
			},
			wantErr: domain.ErrPaymentMethodNotFound,
		},
		{
			name:   "Ошибка БД",
			cardID: "card",
			userID: 1,
			setup: func() {
				mock.ExpectExec(`UPDATE`).
					WithArgs("card", 1).
					WillReturnError(errors.New("fail"))
			},
			wantErr: fmt.Errorf("failed to set default payment method: %w", errors.New("fail")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			err := repo.SetDefault(ctx, tt.cardID, tt.userID)
			if tt.wantErr != nil {
				assert.Equal(t, tt.wantErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
