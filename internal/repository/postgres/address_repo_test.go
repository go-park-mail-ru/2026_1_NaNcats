package postgres

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	"github.com/pashagolub/pgxmock/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddressRepo_CreateAddress(t *testing.T) {
	ctx := context.Background()
	userID := 1
	addr := domain.Address{
		Location: domain.Location{
			AddressText: "Moscow",
			Longitude:   37.6,
			Latitude:    55.7,
		},
		Apartment:      "10",
		Entrance:       "1",
		Floor:          "5",
		DoorCode:       "123",
		CourierComment: "call me",
		Label:          "home",
	}

	tests := []struct {
		name    string
		setup   func(mock pgxmock.PgxPoolIface)
		want    string
		wantErr error
	}{
		{
			name: "Успех",
			setup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectBegin()
				mock.ExpectQuery(`INSERT INTO "location"`).
					WithArgs(addr.Location.AddressText, addr.Location.Longitude, addr.Location.Latitude).
					WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(100))
				mock.ExpectQuery(`INSERT INTO "client_address"`).
					WithArgs(100, userID, addr.Apartment, addr.Entrance, addr.Floor, addr.DoorCode, addr.CourierComment, addr.Label).
					WillReturnRows(pgxmock.NewRows([]string{"public_id"}).AddRow("addr-uuid"))
				mock.ExpectCommit()
			},
			want:    "addr-uuid",
			wantErr: nil,
		},
		{
			name: "Ошибка внесения локации",
			setup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectBegin()
				mock.ExpectQuery(`INSERT INTO "location"`).
					WithArgs(addr.Location.AddressText, addr.Location.Longitude, addr.Location.Latitude).
					WillReturnError(errors.New("db error"))
				mock.ExpectRollback()
			},
			want:    "",
			wantErr: fmt.Errorf("insert location failed: %w", errors.New("db error")),
		},
		{
			name: "Ошибка внесения адреса",
			setup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectBegin()
				mock.ExpectQuery(`INSERT INTO "location"`).
					WithArgs(addr.Location.AddressText, addr.Location.Longitude, addr.Location.Latitude).
					WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(100))
				mock.ExpectQuery(`INSERT INTO "client_address"`).
					WithArgs(100, userID, addr.Apartment, addr.Entrance, addr.Floor, addr.DoorCode, addr.CourierComment, addr.Label).
					WillReturnError(errors.New("db error"))
				mock.ExpectRollback()
			},
			want:    "",
			wantErr: fmt.Errorf("insert address failed: %w", errors.New("db error")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			repo := NewAddressRepo(mock)
			tt.setup(mock)

			got, err := repo.CreateAddress(ctx, userID, addr)
			if tt.wantErr != nil {
				assert.Equal(t, tt.wantErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestAddressRepo_GetAddressesByUserID(t *testing.T) {
	ctx := context.Background()
	userID := 1

	tests := []struct {
		name    string
		setup   func(mock pgxmock.PgxPoolIface)
		want    []domain.Address
		wantErr bool
	}{
		{
			name: "Успех",
			setup: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"public_id", "address_text", "lat", "lon", "apartment", "entrance", "floor_level", "door_code", "courier_comment", "label"}).
					AddRow("uuid-1", "Text 1", 55.0, 37.0, "1", "2", "3", "4", "5", "home").
					AddRow("uuid-2", "Text 2", 56.0, 38.0, "", "", "", "", "", "work")
				mock.ExpectQuery(`SELECT (.+) FROM "client_address"`).
					WithArgs(userID).
					WillReturnRows(rows)
			},
			want: []domain.Address{
				{PublicID: "uuid-1", Location: domain.Location{AddressText: "Text 1", Latitude: 55.0, Longitude: 37.0}, Apartment: "1", Entrance: "2", Floor: "3", DoorCode: "4", CourierComment: "5", Label: "home"},
				{PublicID: "uuid-2", Location: domain.Location{AddressText: "Text 2", Latitude: 56.0, Longitude: 38.0}, Apartment: "", Entrance: "", Floor: "", DoorCode: "", CourierComment: "", Label: "work"},
			},
			wantErr: false,
		},
		{
			name: "Успех: пустой список",
			setup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery(`SELECT`).WithArgs(userID).WillReturnRows(pgxmock.NewRows([]string{"public_id"}))
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "Ошибка запроса",
			setup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery(`SELECT`).
					WithArgs(userID).
					WillReturnError(errors.New("fail"))
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			repo := NewAddressRepo(mock)
			tt.setup(mock)

			got, err := repo.GetAddressesByUserID(ctx, userID)
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

func TestAddressRepo_DeleteAddress(t *testing.T) {
	ctx := context.Background()
	userID := 1
	publicID := "addr-uuid"
	dbErr := errors.New("db error")

	tests := []struct {
		name    string
		setup   func(mock pgxmock.PgxPoolIface)
		wantErr error
	}{
		{
			name: "Успех",
			setup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec(`DELETE FROM "client_address"`).
					WithArgs(publicID, userID).
					WillReturnResult(pgxmock.NewResult("DELETE", 1))
			},
			wantErr: nil,
		},
		{
			name: "Ошибка: адрес не найден",
			setup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec(`DELETE FROM "client_address"`).
					WithArgs(publicID, userID).
					WillReturnResult(pgxmock.NewResult("DELETE", 0))
			},
			wantErr: domain.ErrEmptyDBQuery,
		},
		{
			name: "Ошибка",
			setup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec(`DELETE`).
					WithArgs(publicID, userID).
					WillReturnError(dbErr)
			},
			wantErr: dbErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			repo := NewAddressRepo(mock)
			tt.setup(mock)

			err = repo.DeleteAddress(ctx, userID, publicID)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
