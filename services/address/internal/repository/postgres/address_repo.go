package postgres

import (
	"context"
	"fmt"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/address/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/address/internal/repository"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/postgres"
	"github.com/jackc/pgx/v5"
)

type addressDB struct {
	PublicID       string  `db:"public_id"`
	AddressText    string  `db:"address_text"`
	Latitude       float64 `db:"lat"`
	Longitude      float64 `db:"lon"`
	Apartment      string  `db:"apartment"`
	Entrance       string  `db:"entrance"`
	Floor          string  `db:"floor_level"`
	DoorCode       string  `db:"door_code"`
	CourierComment string  `db:"courier_comment"`
	Label          string  `db:"label"`
}

func (a addressDB) toDomain() domain.Address {
	return domain.Address{
		PublicID: a.PublicID,
		Location: domain.Location{
			AddressText: a.AddressText,
			Latitude:    a.Latitude,
			Longitude:   a.Longitude,
		},
		Apartment:      a.Apartment,
		Entrance:       a.Entrance,
		Floor:          a.Floor,
		DoorCode:       a.DoorCode,
		CourierComment: a.CourierComment,
		Label:          a.Label,
	}
}

type addressRepo struct {
	pool postgres.PgxPool
}

func NewAddressRepo(pool postgres.PgxPool) repository.AddressRepository {
	return &addressRepo{pool: pool}
}

func (r *addressRepo) CreateAddress(ctx context.Context, userID int64, addr domain.Address, idempotencyKey string) (string, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)

	locationQuery := `
		INSERT INTO "location" (address_text, coordinate, idempotency_key)
		VALUES ($1, ST_SetSRID(ST_MakePoint($2, $3), 4326), $4)
		ON CONFLICT (idempotency_key) DO UPDATE
		SET idempotency_key = EXCLUDED.idempotency_key
		RETURNING id;`

	clientAddressQuery := `
		INSERT INTO "client_address" (location_id, client_account_id, apartment, entrance, floor_level, door_code, courier_comment, label, idempotency_key)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (idempotency_key) DO UPDATE
		SET idempotency_key = EXCLUDED.idempotency_key
		RETURNING public_id;
		`

	var locationID int
	err = tx.QueryRow(ctx, locationQuery,
		addr.Location.AddressText,
		addr.Location.Longitude,
		addr.Location.Latitude,
		idempotencyKey).Scan(&locationID)
	if err != nil {
		return "", fmt.Errorf("insert location failed: %w", err)
	}

	var addressPublicID string
	err = tx.QueryRow(ctx, clientAddressQuery,
		locationID,
		userID,
		addr.Apartment,
		addr.Entrance,
		addr.Floor,
		addr.DoorCode,
		addr.CourierComment,
		addr.Label,
		idempotencyKey).Scan(&addressPublicID)
	if err != nil {
		return "", fmt.Errorf("insert address failed: %w", err)
	}

	return addressPublicID, tx.Commit(ctx)
}

func (r *addressRepo) GetAddressesByUserID(ctx context.Context, userID int64) ([]domain.Address, error) {
	query := `
        SELECT 
            a.public_id, 
            l.address_text, 
            ST_Y(l.coordinate::geometry) as lat, 
            ST_X(l.coordinate::geometry) as lon,
            COALESCE(a.apartment, '') AS apartment, 
            COALESCE(a.entrance, '') AS entrance, 
            COALESCE(a.floor_level, '') AS floor_level, 
            COALESCE(a.door_code, '') AS door_code, 
            COALESCE(a.courier_comment, '') AS courier_comment, 
            COALESCE(a.label, '') AS label
        FROM "client_address" a
        JOIN "location" l ON a.location_id = l.id
        WHERE a.client_account_id = $1 AND a.is_active = true
        ORDER BY a.created_at DESC;`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	dbAddresses, err := pgx.CollectRows(rows, pgx.RowToStructByName[addressDB])
	if err != nil {
		return nil, err
	}

	result := make([]domain.Address, 0, len(dbAddresses))
	for _, addr := range dbAddresses {
		result = append(result, addr.toDomain())
	}

	return result, nil
}

func (r *addressRepo) DeleteAddress(ctx context.Context, userID int64, publicID string) error {
	query := `
		UPDATE "client_address"
		SET is_active = false
		WHERE public_id = $1 AND client_account_id = $2;
	`

	res, err := r.pool.Exec(ctx, query, publicID, userID)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return domain.ErrNoChangesProvided
	}

	return nil
}

func (r *addressRepo) UpdateAddress(ctx context.Context, userID int64, addr domain.Address) error {
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
			WHERE public_id = $4 AND client_account_id = $5
		);`

	_, err = tx.Exec(ctx, queryLoc,
		addr.Location.AddressText,
		addr.Location.Longitude,
		addr.Location.Latitude,
		addr.PublicID,
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
		WHERE public_id = $7 AND client_account_id = $8;`

	result, err := tx.Exec(ctx, queryAddr,
		addr.Apartment,
		addr.Entrance,
		addr.Floor,
		addr.DoorCode,
		addr.CourierComment,
		addr.Label,
		addr.PublicID,
		userID,
	)
	if err != nil {
		return fmt.Errorf("update client_address failed: %w", err)
	}

	if result.RowsAffected() == 0 {
		// TODO: сделать что-то с этой ошибкой
		return domain.ErrAddressNotFound
	}

	return tx.Commit(ctx)
}

func (r *addressRepo) GetInternalIDByPublicID(ctx context.Context, userID int64, publicID string) (int, error) {
	query := `
		SELECT id FROM "client_address"
		WHERE public_id = $1 AND client_account_id = $2;
	`

	var internalID int
	err := r.pool.QueryRow(ctx, query, publicID, userID).Scan(&internalID)
	if err != nil {
		return 0, err
	}

	return internalID, err
}
