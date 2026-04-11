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
	return &addressRepo{
		pool: pool,
	}
}

func (r *addressRepo) CreateAddress(ctx context.Context, userID int, addr domain.Address) (string, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)

	var locationID int
	err = tx.QueryRow(ctx, `
		INSERT INTO "location" (address_text, coordinate)
		VALUES ($1, ST_SetSRID(ST_MakePoint($2, $3), 4326))
		RETURNING id`,
		addr.Location.AddressText, addr.Location.Longitude, addr.Location.Latitude).Scan(&locationID)
	if err != nil {
		return "", fmt.Errorf("insert location failed: %w", err)
	}

	var addressPublicID string
	err = tx.QueryRow(ctx, `
		INSERT INTO "client_address" (location_id, client_account_id, apartment, entrance, floor_level, door_code, courier_comment, label)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING public_id`,
		locationID, userID, addr.Apartment, addr.Entrance, addr.Floor, addr.DoorCode, addr.CourierComment, addr.Label).Scan(&addressPublicID)
	if err != nil {
		return "", fmt.Errorf("insert address failed: %w", err)
	}

	return addressPublicID, tx.Commit(ctx)
}

func (r *addressRepo) GetAddressesByUserID(ctx context.Context, userID int) ([]domain.Address, error) {
	query := `
		SELECT a.public_id, l.address_text, ST_Y(l.coordinate::geometry) as lat, ST_X(l.coordinate::geometry) as lon,
		       COALESCE(a.apartment, ''), COALESCE(a.entrance, ''), COALESCE(a.floor_level, ''), 
			   COALESCE(a.door_code, ''), COALESCE(a.courier_comment, ''), COALESCE(a.label, '')
		FROM "client_address" a
		JOIN "location" l ON a.location_id = l.id
		WHERE a.client_account_id = $1
		ORDER BY a.created_at DESC;`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var addresses []domain.Address
	for rows.Next() {
		var a domain.Address
		err := rows.Scan(
			&a.PublicID, &a.Location.AddressText, &a.Location.Latitude, &a.Location.Longitude,
			&a.Apartment, &a.Entrance, &a.Floor, &a.DoorCode, &a.CourierComment, &a.Label,
		)
		if err != nil {
			return nil, err
		}
		addresses = append(addresses, a)
	}
	return addresses, nil
}

func (r *addressRepo) DeleteAddress(ctx context.Context, userID int, publicID string) error {
	res, err := r.pool.Exec(ctx, `DELETE FROM "client_address" WHERE public_id = $1 AND client_account_id = $2`, publicID, userID)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return domain.ErrEmptyDBQuery
	}

	return nil
}

func (r *addressRepo) UpdateAddress(ctx context.Context, userID int, addr domain.Address) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	queryLoc := `
		UPDATE "location"
		SET address_text = $1, 
		    coordinate = ST_SetSRID(ST_MakePoint($2, $3), 4326)
		WHERE id = (
			SELECT location_id 
			FROM "client_address" 
			WHERE id = $4 AND client_account_id = $5
		)`

	_, err = tx.Exec(ctx, queryLoc,
		addr.Location.AddressText,
		addr.Location.Longitude,
		addr.Location.Latitude,
		addr.ID,
		userID,
	)
	if err != nil {
		return fmt.Errorf("update location failed: %w", err)
	}

	queryAddr := `
		UPDATE "client_address"
		SET apartment = $1, 
		    entrance = $2, 
		    floor_level = $3, 
		    door_code = $4, 
		    courier_comment = $5, 
		    label = $6
		WHERE id = $7 AND client_account_id = $8`

	result, err := tx.Exec(ctx, queryAddr,
		addr.Apartment,
		addr.Entrance,
		addr.Floor,
		addr.DoorCode,
		addr.CourierComment,
		addr.Label,
		addr.ID,
		userID,
	)
	if err != nil {
		return fmt.Errorf("update client_address failed: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrAddressNotFound
	}

	return tx.Commit(ctx)
}

func (r *addressRepo) GetInternalIDByPublicID(ctx context.Context, userID int, publicID string) (int, error) {
	query := `
		SELECT id FROM "client_address"
		WHERE public_id = $1 AND client_account_id = $2
	`

	var internalID int
	err := r.pool.QueryRow(ctx, query, publicID, userID).Scan(&internalID)
	if err != nil {
		return 0, err
	}

	return internalID, err
}
