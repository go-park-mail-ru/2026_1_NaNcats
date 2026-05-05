package postgres

import (
	"context"
	"errors"
	"testing"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/payment/internal/domain"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pashagolub/pgxmock/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPaymentRepo_Create(t *testing.T) {
	type mockBehavior func(mock pgxmock.PgxPoolIface, method domain.PaymentMethod, idemKey string)

	method := domain.PaymentMethod{
		UserID:      1,
		ExternalID:  "ext-123",
		First6:      "123456",
		Last4:       "7890",
		ExpiryMonth: "12",
		ExpiryYear:  "2025",
		CardType:    "Visa",
		IssuerName:  "Bank",
		IsDefault:   true,
	}

	tests := []struct {
		name         string
		method       domain.PaymentMethod
		idemKey      string
		mockBehavior mockBehavior
		expectedID   int64
		expectedErr  error
	}{
		{
			name:    "Успешное создание метода оплаты",
			method:  method,
			idemKey: "idem-key-1",
			mockBehavior: func(mock pgxmock.PgxPoolIface, method domain.PaymentMethod, idemKey string) {
				mock.ExpectQuery(`INSERT INTO "payment_method"`).
					WithArgs(
						method.UserID, method.ExternalID, method.First6, method.Last4,
						method.ExpiryMonth, method.ExpiryYear, method.CardType,
						method.IssuerName, method.IsDefault, idemKey,
					).
					WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(int64(42)))
			},
			expectedID:  42,
			expectedErr: nil,
		},
		{
			name:    "Ошибка: конфликт уникальности (такой метод уже есть)",
			method:  method,
			idemKey: "idem-key-2",
			mockBehavior: func(mock pgxmock.PgxPoolIface, method domain.PaymentMethod, idemKey string) {
				pgErr := &pgconn.PgError{
					Code:           pgerrcode.UniqueViolation,
					ConstraintName: "payment_method_user_id_external_id_key",
				}
				mock.ExpectQuery(`INSERT INTO "payment_method"`).
					WithArgs(
						method.UserID, method.ExternalID, method.First6, method.Last4,
						method.ExpiryMonth, method.ExpiryYear, method.CardType,
						method.IssuerName, method.IsDefault, idemKey,
					).
					WillReturnError(pgErr)
			},
			expectedID:  0,
			expectedErr: domain.ErrPaymentMethodAlreadyExists,
		},
		{
			name:    "Системная ошибка БД",
			method:  method,
			idemKey: "idem-key-3",
			mockBehavior: func(mock pgxmock.PgxPoolIface, method domain.PaymentMethod, idemKey string) {
				mock.ExpectQuery(`INSERT INTO "payment_method"`).
					WithArgs(
						method.UserID, method.ExternalID, method.First6, method.Last4,
						method.ExpiryMonth, method.ExpiryYear, method.CardType,
						method.IssuerName, method.IsDefault, idemKey,
					).
					WillReturnError(errors.New("db error"))
			},
			expectedID:  0,
			expectedErr: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			tt.mockBehavior(mock, tt.method, tt.idemKey)

			repo := NewPaymentRepo(mock)
			id, err := repo.Create(context.Background(), tt.method, tt.idemKey)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				if errors.Is(tt.expectedErr, domain.ErrPaymentMethodAlreadyExists) {
					assert.ErrorIs(t, err, domain.ErrPaymentMethodAlreadyExists)
				}
				assert.Equal(t, tt.expectedID, id)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedID, id)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestPaymentRepo_Delete(t *testing.T) {
	type mockBehavior func(mock pgxmock.PgxPoolIface, cardID string, userID int64)

	tests := []struct {
		name         string
		cardID       string
		userID       int64
		mockBehavior mockBehavior
		expectedErr  error
	}{
		{
			name:   "Успешное удаление карты",
			cardID: "ext-123",
			userID: 1,
			mockBehavior: func(mock pgxmock.PgxPoolIface, cardID string, userID int64) {
				mock.ExpectExec(`DELETE FROM "payment_method"`).
					WithArgs(cardID, userID).
					WillReturnResult(pgxmock.NewResult("DELETE", 1))
			},
			expectedErr: nil,
		},
		{
			name:   "Ошибка: карта не найдена",
			cardID: "ext-404",
			userID: 1,
			mockBehavior: func(mock pgxmock.PgxPoolIface, cardID string, userID int64) {
				mock.ExpectExec(`DELETE FROM "payment_method"`).
					WithArgs(cardID, userID).
					WillReturnResult(pgxmock.NewResult("DELETE", 0)) // 0 затронутых строк
			},
			expectedErr: domain.ErrPaymentMethodNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			tt.mockBehavior(mock, tt.cardID, tt.userID)

			repo := NewPaymentRepo(mock)
			err = repo.Delete(context.Background(), tt.cardID, tt.userID)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestPaymentRepo_GetByUserID(t *testing.T) {
	type mockBehavior func(mock pgxmock.PgxPoolIface, userID int64)

	issuerStr := "Tinkoff"

	tests := []struct {
		name         string
		userID       int64
		mockBehavior mockBehavior
		expectedLen  int
		expectedErr  bool
	}{
		{
			name:   "Успешное получение списка карт",
			userID: 1,
			mockBehavior: func(mock pgxmock.PgxPoolIface, userID int64) {
				// Описываем колонки, как они приходят из БД
				cols := []string{"id", "user_id", "external_id", "first6", "last4", "expiry_month", "expiry_year", "card_type", "issuer_name", "is_default"}

				rows := pgxmock.NewRows(cols).
					AddRow(int64(1), userID, "ext-1", "111111", "1111", "01", "2024", "Visa", nil, true).
					AddRow(int64(2), userID, "ext-2", "222222", "2222", "02", "2025", "MasterCard", &issuerStr, false)

				mock.ExpectQuery(`SELECT (.+) FROM "payment_method"`).
					WithArgs(userID).
					WillReturnRows(rows)
			},
			expectedLen: 2,
			expectedErr: false,
		},
		{
			name:   "Ошибка БД при запросе",
			userID: 1,
			mockBehavior: func(mock pgxmock.PgxPoolIface, userID int64) {
				mock.ExpectQuery(`SELECT (.+) FROM "payment_method"`).
					WithArgs(userID).
					WillReturnError(errors.New("db disconnect"))
			},
			expectedLen: 0,
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			tt.mockBehavior(mock, tt.userID)

			repo := NewPaymentRepo(mock)
			methods, err := repo.GetByUserID(context.Background(), tt.userID)

			if tt.expectedErr {
				assert.Error(t, err)
				assert.Nil(t, methods)
			} else {
				assert.NoError(t, err)
				assert.Len(t, methods, tt.expectedLen)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestPaymentRepo_SetDefault(t *testing.T) {
	type mockBehavior func(mock pgxmock.PgxPoolIface, cardID string, userID int64)

	tests := []struct {
		name         string
		cardID       string
		userID       int64
		mockBehavior mockBehavior
		expectedErr  error
	}{
		{
			name:   "Успешная установка карты по умолчанию",
			cardID: "ext-123",
			userID: 1,
			mockBehavior: func(mock pgxmock.PgxPoolIface, cardID string, userID int64) {
				mock.ExpectExec(`UPDATE "payment_method"`).
					WithArgs(cardID, userID).
					WillReturnResult(pgxmock.NewResult("UPDATE", 2)) // Допустим 1 сняли дефолт, 1 поставили
			},
			expectedErr: nil,
		},
		{
			name:   "Ошибка: карта не найдена",
			cardID: "ext-404",
			userID: 1,
			mockBehavior: func(mock pgxmock.PgxPoolIface, cardID string, userID int64) {
				mock.ExpectExec(`UPDATE "payment_method"`).
					WithArgs(cardID, userID).
					WillReturnResult(pgxmock.NewResult("UPDATE", 0)) // 0 затронутых строк
			},
			expectedErr: domain.ErrPaymentMethodNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			tt.mockBehavior(mock, tt.cardID, tt.userID)

			repo := NewPaymentRepo(mock)
			err = repo.SetDefault(context.Background(), tt.cardID, tt.userID)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
