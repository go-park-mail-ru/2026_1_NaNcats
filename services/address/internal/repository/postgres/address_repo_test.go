package postgres

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/address/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddressRepo_CreateAddress(t *testing.T) {
	ctx := context.Background()
	var userID int64 = 1
	idemKey := "idem-addr-123"
	addr := domain.Address{
		Location: domain.Location{
			AddressText: "Moscow, Red Square",
			Longitude:   37.61,
			Latitude:    55.75,
		},
		Apartment:      "1",
		Entrance:       "2",
		Floor:          "3",
		DoorCode:       "4",
		CourierComment: "Call me",
		Label:          "Home",
	}

	type mockInit func(m pgxmock.PgxPoolIface)
	tests := []struct {
		name          string
		mockInit      mockInit
		expectedID    string
		expectedError string
	}{
		{
			name: "Успешное создание адреса (первичное)",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()
				m.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "idempotency_records"`)).
					WithArgs(userID, idemKey).
					WillReturnRows(pgxmock.NewRows([]string{"response_payload"}).AddRow(nil))

				m.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "location"`)).
					WithArgs(addr.Location.AddressText, addr.Location.Longitude, addr.Location.Latitude).
					WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(100))

				m.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "client_address"`)).
					WithArgs(100, userID, addr.Apartment, addr.Entrance, addr.Floor, addr.DoorCode, addr.CourierComment, addr.Label).
					WillReturnRows(pgxmock.NewRows([]string{"public_id"}).AddRow("uuid-addr-777"))

				m.ExpectExec(regexp.QuoteMeta(`UPDATE "idempotency_records"`)).
					WithArgs(pgxmock.AnyArg(), userID, idemKey).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))

				m.ExpectCommit()
			},
			expectedID: "uuid-addr-777",
		},
		{
			name: "Возврат результата по ключу идемпотентности",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()
				m.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "idempotency_records"`)).
					WithArgs(userID, idemKey).
					WillReturnError(pgx.ErrNoRows)

				m.ExpectQuery(regexp.QuoteMeta(`SELECT response_payload FROM "idempotency_records"`)).
					WithArgs(userID, idemKey).
					WillReturnRows(pgxmock.NewRows([]string{"response_payload"}).AddRow([]byte(`{"public_id":"uuid-cached-888"}`)))

				m.ExpectRollback()
			},
			expectedID: "uuid-cached-888",
		},
		{
			name: "Ошибка при вставке локации (откат)",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()
				m.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "idempotency_records"`)).
					WithArgs(userID, idemKey).
					WillReturnRows(pgxmock.NewRows([]string{"response_payload"}).AddRow(nil))

				m.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "location"`)).
					WithArgs(addr.Location.AddressText, addr.Location.Longitude, addr.Location.Latitude).
					WillReturnError(errors.New("db error"))

				m.ExpectRollback()
			},
			expectedError: "insert location failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			repo := NewAddressRepo(mock)
			tt.mockInit(mock)

			res, err := repo.CreateAddress(ctx, userID, addr, idemKey)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedID, res)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestAddressRepo_GetAddressesByUserID(t *testing.T) {
	ctx := context.Background()
	var userID int64 = 1
	columns := []string{
		"public_id", "address_text", "lat", "lon",
		"apartment", "entrance", "floor_level", "door_code",
		"courier_comment", "label",
	}

	type mockInit func(m pgxmock.PgxPoolIface)
	tests := []struct {
		name          string
		mockInit      mockInit
		expectedCount int
		wantErr       bool
	}{
		{
			name: "Успешное получение списка адресов",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
					WithArgs(userID).
					WillReturnRows(pgxmock.NewRows(columns).
						AddRow("uuid-1", "Text 1", 55.0, 37.0, "1", "2", "3", "4", "c1", "Home").
						AddRow("uuid-2", "Text 2", 56.0, 38.0, "", "", "", "", "", "Work"))
			},
			expectedCount: 2,
		},
		{
			name: "Пустой список",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
					WithArgs(userID).
					WillReturnRows(pgxmock.NewRows(columns))
			},
			expectedCount: 0,
		},
		{
			name: "Ошибка базы данных",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
					WithArgs(userID).
					WillReturnError(errors.New("scan fail"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, _ := pgxmock.NewPool()
			defer mock.Close()

			repo := NewAddressRepo(mock)
			tt.mockInit(mock)

			res, err := repo.GetAddressesByUserID(ctx, userID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, res, tt.expectedCount)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestAddressRepo_DeleteAddress(t *testing.T) {
	ctx := context.Background()
	var userID int64 = 1
	publicID := "uuid-to-delete"
	idemKey := "idem-del-123"

	type mockInit func(m pgxmock.PgxPoolIface)
	tests := []struct {
		name          string
		mockInit      mockInit
		expectedError error
	}{
		{
			name: "Успешное удаление",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()
				m.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "idempotency_records"`)).
					WithArgs(userID, idemKey).
					WillReturnRows(pgxmock.NewRows([]string{"response_payload"}).AddRow(nil))

				m.ExpectExec(regexp.QuoteMeta(`UPDATE "client_address"`)).
					WithArgs(publicID, userID).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))

				m.ExpectExec(regexp.QuoteMeta(`UPDATE "idempotency_records"`)).
					WithArgs(userID, idemKey).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))

				m.ExpectCommit()
			},
		},
		{
			name: "Ретрай идемпотентного запроса",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()
				m.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "idempotency_records"`)).
					WithArgs(userID, idemKey).
					WillReturnError(pgx.ErrNoRows)
				m.ExpectRollback()
			},
		},
		{
			name: "Адрес не найден",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()
				m.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "idempotency_records"`)).
					WithArgs(userID, idemKey).
					WillReturnRows(pgxmock.NewRows([]string{"response_payload"}).AddRow(nil))

				m.ExpectExec(regexp.QuoteMeta(`UPDATE "client_address"`)).
					WithArgs(publicID, userID).
					WillReturnResult(pgxmock.NewResult("UPDATE", 0))

				m.ExpectExec(regexp.QuoteMeta(`DELETE FROM "idempotency_records"`)).
					WithArgs(userID, idemKey).
					WillReturnResult(pgxmock.NewResult("DELETE", 1))

				m.ExpectRollback()
			},
			expectedError: domain.ErrAddressNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			repo := NewAddressRepo(mock)
			tt.mockInit(mock)

			err = repo.DeleteAddress(ctx, userID, publicID, idemKey)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestAddressRepo_UpdateAddress(t *testing.T) {
	ctx := context.Background()
	userID := int64(1)
	idemKey := "idem-upd-123"
	addr := domain.Address{
		PublicID: "addr-uuid-123",
		Location: domain.Location{
			AddressText: "Новый адрес",
			Longitude:   37.0,
			Latitude:    55.0,
		},
		Apartment:      "10",
		Entrance:       "1",
		Floor:          "5",
		DoorCode:       "123",
		CourierComment: "звонить заранее",
		Label:          "Дом",
	}

	type mockInit func(m pgxmock.PgxPoolIface)
	tests := []struct {
		name          string
		mockInit      mockInit
		expectedError string
	}{
		{
			name: "Успешное обновление",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()
				m.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "idempotency_records"`)).
					WithArgs(userID, idemKey).
					WillReturnRows(pgxmock.NewRows([]string{"response_payload"}).AddRow(nil))

				m.ExpectQuery(regexp.QuoteMeta(`SELECT location_id FROM "client_address"`)).
					WithArgs(addr.PublicID, userID).
					WillReturnRows(pgxmock.NewRows([]string{"location_id"}).AddRow(100))

				m.ExpectExec(regexp.QuoteMeta(`UPDATE "location"`)).
					WithArgs(addr.Location.AddressText, addr.Location.Longitude, addr.Location.Latitude, 100).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))

				m.ExpectExec(regexp.QuoteMeta(`UPDATE "client_address"`)).
					WithArgs(addr.Apartment, addr.Entrance, addr.Floor, addr.DoorCode, addr.CourierComment, addr.Label, addr.PublicID, userID).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))

				m.ExpectExec(regexp.QuoteMeta(`UPDATE "idempotency_records"`)).
					WithArgs(userID, idemKey).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))

				m.ExpectCommit()
			},
		},
		{
			name: "Ретрай идемпотентного запроса",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()
				m.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "idempotency_records"`)).
					WithArgs(userID, idemKey).
					WillReturnError(pgx.ErrNoRows)
				m.ExpectRollback()
			},
		},
		{
			name: "Адрес не найден",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()
				m.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "idempotency_records"`)).
					WithArgs(userID, idemKey).
					WillReturnRows(pgxmock.NewRows([]string{"response_payload"}).AddRow(nil))

				m.ExpectQuery(regexp.QuoteMeta(`SELECT location_id FROM "client_address"`)).
					WithArgs(addr.PublicID, userID).
					WillReturnError(pgx.ErrNoRows)

				m.ExpectExec(regexp.QuoteMeta(`DELETE FROM "idempotency_records"`)).
					WithArgs(userID, idemKey).
					WillReturnResult(pgxmock.NewResult("DELETE", 1))

				m.ExpectRollback()
			},
			expectedError: domain.ErrAddressNotFound.Error(),
		},
		{
			name: "Ошибка при обновлении локации",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()
				m.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "idempotency_records"`)).
					WithArgs(userID, idemKey).
					WillReturnRows(pgxmock.NewRows([]string{"response_payload"}).AddRow(nil))

				m.ExpectQuery(regexp.QuoteMeta(`SELECT location_id FROM "client_address"`)).
					WithArgs(addr.PublicID, userID).
					WillReturnRows(pgxmock.NewRows([]string{"location_id"}).AddRow(100))

				m.ExpectExec(regexp.QuoteMeta(`UPDATE "location"`)).
					WithArgs(addr.Location.AddressText, addr.Location.Longitude, addr.Location.Latitude, 100).
					WillReturnError(errors.New("db crash"))

				m.ExpectRollback()
			},
			expectedError: "update location failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			repo := NewAddressRepo(mock)
			tt.mockInit(mock)

			err = repo.UpdateAddress(ctx, userID, addr, idemKey)

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

func TestAddressRepo_GetInternalIDByPublicID(t *testing.T) {
	ctx := context.Background()
	userID := int64(1)
	publicID := "uuid-123"

	type mockInit func(m pgxmock.PgxPoolIface)
	tests := []struct {
		name          string
		mockInit      mockInit
		expectedID    int
		expectedError error
	}{
		{
			name: "Успешное получение",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(regexp.QuoteMeta(`SELECT id FROM "client_address"`)).
					WithArgs(publicID, userID).
					WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(100))
			},
			expectedID: 100,
		},
		{
			name: "ID не найден",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(regexp.QuoteMeta(`SELECT id`)).
					WithArgs(publicID, userID).
					WillReturnError(pgx.ErrNoRows)
			},
			expectedError: domain.ErrAddressNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, _ := pgxmock.NewPool()
			defer mock.Close()

			repo := NewAddressRepo(mock)
			tt.mockInit(mock)

			res, err := repo.GetInternalIDByPublicID(ctx, userID, publicID)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedID, res)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestAddressRepo_CheckAddressExists(t *testing.T) {
	ctx := context.Background()
	userID := int64(1)
	publicID := "uuid-123"

	type mockInit func(m pgxmock.PgxPoolIface)
	tests := []struct {
		name          string
		mockInit      mockInit
		expectedError error
	}{
		{
			name: "Адрес существует",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(regexp.QuoteMeta(`SELECT 1 FROM "client_address"`)).
					WithArgs(userID, publicID).
					WillReturnRows(pgxmock.NewRows([]string{"?column?"}).AddRow(1))
			},
			expectedError: nil,
		},
		{
			name: "Адрес не найден",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(regexp.QuoteMeta(`SELECT 1 FROM "client_address"`)).
					WithArgs(userID, publicID).
					WillReturnError(pgx.ErrNoRows)
			},
			expectedError: domain.ErrAddressNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			repo := NewAddressRepo(mock)
			tt.mockInit(mock)

			err = repo.CheckAddressExists(ctx, userID, publicID)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
