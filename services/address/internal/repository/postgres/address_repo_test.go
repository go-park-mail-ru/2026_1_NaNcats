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
			name: "Успешное создание адреса",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()
				// 1. Вставка локации (4 аргумента: text, lon, lat, idem)
				m.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "location"`)).
					WithArgs(addr.Location.AddressText, addr.Location.Longitude, addr.Location.Latitude, idemKey).
					WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(100))

				// 2. Вставка адреса клиента (9 аргументов: loc_id, user_id, apt, ent, floor, code, comment, label, idem)
				m.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "client_address"`)).
					WithArgs(100, userID, addr.Apartment, addr.Entrance, addr.Floor, addr.DoorCode, addr.CourierComment, addr.Label, idemKey).
					WillReturnRows(pgxmock.NewRows([]string{"public_id"}).AddRow("uuid-addr-777"))

				m.ExpectCommit()
			},
			expectedID: "uuid-addr-777",
		},
		{
			name: "Ошибка при вставке локации (откат)",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()
				m.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "location"`)).
					WithArgs(addr.Location.AddressText, addr.Location.Longitude, addr.Location.Latitude, idemKey).
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

	type mockInit func(m pgxmock.PgxPoolIface)
	tests := []struct {
		name          string
		mockInit      mockInit
		expectedError error
	}{
		{
			name: "Успешное удаление",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectExec(regexp.QuoteMeta(`UPDATE "client_address" SET is_active = false`)).
					WithArgs(publicID, userID).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
			},
		},
		{
			name: "Ошибка исполнения запроса",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectExec(regexp.QuoteMeta(`UPDATE "client_address"`)).
					WithArgs(publicID, userID).
					WillReturnError(errors.New("query err"))
			},
			expectedError: errors.New("query err"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, _ := pgxmock.NewPool()
			defer mock.Close()

			repo := NewAddressRepo(mock)
			tt.mockInit(mock)

			err := repo.DeleteAddress(ctx, userID, publicID)

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
				m.ExpectExec(regexp.QuoteMeta(`UPDATE "location"`)).
					WithArgs(addr.Location.AddressText, addr.Location.Longitude, addr.Location.Latitude, addr.PublicID, userID).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))

				m.ExpectExec(regexp.QuoteMeta(`UPDATE "client_address"`)).
					WithArgs(addr.Apartment, addr.Entrance, addr.Floor, addr.DoorCode, addr.CourierComment, addr.Label, addr.PublicID, userID).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))

				m.ExpectCommit()
			},
		},
		{
			name: "Адрес не найден (второй UPDATE затронул 0 строк)",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()
				m.ExpectExec(regexp.QuoteMeta(`UPDATE "location"`)).
					WithArgs(addr.Location.AddressText, addr.Location.Longitude, addr.Location.Latitude, addr.PublicID, userID).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))

				m.ExpectExec(regexp.QuoteMeta(`UPDATE "client_address"`)).
					WithArgs(addr.Apartment, addr.Entrance, addr.Floor, addr.DoorCode, addr.CourierComment, addr.Label, addr.PublicID, userID).
					WillReturnResult(pgxmock.NewResult("UPDATE", 0))

				m.ExpectRollback()
			},
			expectedError: domain.ErrAddressNotFound.Error(),
		},
		{
			name: "Ошибка при обновлении локации",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()
				m.ExpectExec(regexp.QuoteMeta(`UPDATE "location"`)).
					WithArgs(addr.Location.AddressText, addr.Location.Longitude, addr.Location.Latitude, addr.PublicID, userID).
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

			err = repo.UpdateAddress(ctx, userID, addr)

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
			expectedError: pgx.ErrNoRows,
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
