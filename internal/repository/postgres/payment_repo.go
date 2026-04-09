package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/repository"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type paymentMethodDB struct {
	ID         int     `db:"id"`
	UserID     int     `db:"user_id"`
	ExternalID string  `db:"external_id"`
	CardType   string  `db:"card_type"`
	Last4      string  `db:"last4"`
	IssuerName *string `db:"issuer_name"`
	IsDefault  bool    `db:"is_default"`
}

func (p paymentMethodDB) toDomain() domain.PaymentMethod {
	issuerName := ""
	if p.IssuerName != nil {
		issuerName = *p.IssuerName
	}

	return domain.PaymentMethod{
		ID:         p.ID,
		UserID:     p.UserID,
		ExternalID: p.ExternalID,
		CardType:   p.CardType,
		Last4:      p.Last4,
		IssuerName: issuerName,
		IsDefault:  p.IsDefault,
	}
}

type paymentRepo struct {
	pool *pgxpool.Pool
}

func NewPaymentRepo(pool *pgxpool.Pool) repository.PaymentRepository {
	return &paymentRepo{
		pool: pool,
	}
}

func (r *paymentRepo) Create(ctx context.Context, method domain.PaymentMethod) (int, error) {
	query := `
		INSERT INTO "payment_method" (user_id, external_id, card_type, last4, issuer_name, is_default)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id;
	`

	var lastInsertedID int
	err := r.pool.QueryRow(ctx, query,
		method.UserID,
		method.ExternalID,
		method.CardType,
		method.Last4,
		method.IssuerName,
		method.IsDefault,
	).Scan(&lastInsertedID)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation { // проверка на уникальность
			return 0, domain.ErrPaymentMethodAlreadyExists
		}
		return 0, err
	}

	return lastInsertedID, nil
}

func (r *paymentRepo) Delete(ctx context.Context, cardID string, userID int) error {
	query := `
		DELETE FROM "payment_method"
		WHERE external_id = $1 AND user_id = $2
	`

	tag, err := r.pool.Exec(ctx, query, cardID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete payment method: %w", err)
	}

	rowsAffected := tag.RowsAffected()
	if rowsAffected == 0 {
		return domain.ErrPaymentMethodNotFound
	}

	return nil
}

func (r *paymentRepo) GetByUserID(ctx context.Context, userID int) ([]domain.PaymentMethod, error) {
	query := `
		SELECT id, user_id, external_id, card_type, last4, issuer_name, is_default
		FROM "payment_method" WHERE user_id = $1
		ORDER BY is_default DESC, created_at DESC;
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	dbPaymentMethods, err := pgx.CollectRows(rows, pgx.RowToStructByName[paymentMethodDB])
	if err != nil {
		return nil, err
	}

	domainPaymentMethods := make([]domain.PaymentMethod, 0, len(dbPaymentMethods))
	for _, dbPaymentMethod := range dbPaymentMethods {
		domainPaymentMethods = append(domainPaymentMethods, dbPaymentMethod.toDomain())
	}

	return domainPaymentMethods, nil
}

func (r *paymentRepo) SetDefault(ctx context.Context, cardID string, userID int) error {
	query := `
		UPDATE "payment_method"
		SET is_default = (external_id = $1)
		WHERE user_id = $2 AND (is_default = true OR external_id = $1);
	`

	tag, err := r.pool.Exec(ctx, query, cardID, userID)
	if err != nil {
		return fmt.Errorf("failed to set default payment method: %w", err)
	}

	rowsAffected := tag.RowsAffected()
	if rowsAffected == 0 {
		return domain.ErrPaymentMethodNotFound
	}

	return nil
}
