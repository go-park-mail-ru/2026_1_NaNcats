package postgres

import (
	"context"
	"fmt"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

type addressRepo struct {
	pool *pgxpool.Pool
}

func NewAddressRepo(pool *pgxpool.Pool) repository.AddressRepository {
	return &addressRepo{pool: pool}
}

func (r *addressRepo) CreateAddress(ctx context.Context, userID int, addr domain.Address) (int, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	var locationID int
	err = tx.QueryRow(ctx, `
		INSERT INTO "location" (address_text, coordinate)
		VALUES ($1, ST_SetSRID(ST_MakePoint($2, $3), 4326))
		RETURNING id`, 
		addr.Location.AddressText, addr.Location.Longitude, addr.Location.Latitude).Scan(&locationID)
	if err != nil {
		return 0, fmt.Errorf("insert location failed: %w", err)
	}

	var addressID int
	err = tx.QueryRow(ctx, `
		INSERT INTO "client_address" (location_id, client_account_id, apartment, entrance, floor_level, door_code, courier_comment, label)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id`,
		locationID, userID, addr.Apartment, addr.Entrance, addr.Floor, addr.DoorCode, addr.CourierComment, addr.Label).Scan(&addressID)
	if err != nil {
		return 0, fmt.Errorf("insert address failed: %w", err)
	}

	return addressID, tx.Commit(ctx)
}

func (r *addressRepo) GetAddressesByUserID(ctx context.Context, userID int) ([]domain.Address, error) {
	query := `
		SELECT a.id, l.address_text, ST_Y(l.coordinate::geometry) as lat, ST_X(l.coordinate::geometry) as lon,
		       a.apartment, a.entrance, a.floor_level, a.door_code, a.courier_comment, a.label
		FROM "client_address" a
		JOIN "location" l ON a.location_id = l.id
		WHERE a.client_account_id = $1
		ORDER BY a.created_at DESC`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var addresses []domain.Address
	for rows.Next() {
		var a domain.Address
		err := rows.Scan(
			&a.ID, &a.Location.AddressText, &a.Location.Latitude, &a.Location.Longitude,
			&a.Apartment, &a.Entrance, &a.Floor, &a.DoorCode, &a.CourierComment, &a.Label,
		)
		if err != nil {
			return nil, err
		}
		addresses = append(addresses, a)
	}
	return addresses, nil
}

func (r *addressRepo) DeleteAddress(ctx context.Context, userID int, addressID int) error {
	res, err := r.pool.Exec(ctx, `DELETE FROM "client_address" WHERE id = $1 AND client_account_id = $2`, addressID, userID)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return domain.ErrEmptyDBQuery
	}
	return nil
}
