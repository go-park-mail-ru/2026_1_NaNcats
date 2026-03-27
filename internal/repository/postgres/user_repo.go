package postgres

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/repository"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type userRepo struct {
	pool *pgxpool.Pool
}

func NewUserRepo(pool *pgxpool.Pool) repository.UserRepository {
	return &userRepo{
		pool: pool,
	}
}

func (r *userRepo) CreateUser(ctx context.Context, user domain.User) (int, error) {
	user.Email = strings.ToLower(strings.TrimSpace(user.Email))

	query := `
		INSERT INTO "user" (name, email, phone, password_hash, user_role)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	var lastInsertedID int
	err := r.pool.QueryRow(ctx, query,
		user.Name,
		user.Email,
		user.Phone,
		user.PasswordHash,
		"client",
	).Scan(&lastInsertedID)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // 23505 - проверка на уникальность
			return 0, domain.ErrEmailAlreadyExists
		}
		return 0, err
	}

	return lastInsertedID, nil
}

func (r *userRepo) GetUserByEmail(ctx context.Context, email string) (domain.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	query := `
		SELECT id, name, email, phone, password_hash, user_role
		FROM "user"
		WHERE email = $1
	`

	var user domain.User
	var userRole string // заглушка

	err := r.pool.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Phone,
		&user.PasswordHash,
		&userRole,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.User{}, domain.ErrUserNotFound
		}
		return domain.User{}, err
	}

	return user, nil
}

func (r *userRepo) GetUserByID(ctx context.Context, id int) (domain.User, error) {
	query := `
		SELECT id, name, email, phone, password_hash, user_role
		FROM "user"
		WHERE id = $1
	`

	var user domain.User
	var userRole string // заглушка

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Phone,
		&user.PasswordHash,
		&userRole,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.User{}, domain.ErrUserNotFound
		}
		return domain.User{}, err
	}

	return user, nil
}
