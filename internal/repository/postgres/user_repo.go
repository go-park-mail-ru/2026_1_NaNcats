package postgres

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

type userRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) repository.UserRepository {
	return &userRepo{
		db: db,
	}
}

func (r *userRepo) CreateUser(ctx context.Context, user domain.User) (uuid.UUID, error) {
	user.Email = strings.ToLower(strings.TrimSpace(user.Email))

	user.ID = uuid.New()

	query := `
		INSERT INTO "user" (id, name, email, phone, password_hash, user_role)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.Name,
		user.Email,
		user.Phone,
		user.PasswordHash,
		"client",
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // 23505 - проверка на уникальность
			return uuid.Nil, domain.ErrEmailAlreadyExists
		}
		return uuid.Nil, err
	}

	return user.ID, nil
}

func (r *userRepo) GetUserByEmail(ctx context.Context, email string) (domain.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	query := `
		SELECT id, name, email, phone, password_hash, user_role
		FROM "user"
		WHERE email = $1
	`

	var user_role string // заглушка

	var user domain.User
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Phone,
		&user.PasswordHash,
		user_role,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.User{}, domain.ErrUserNotFound
		}
		return domain.User{}, err
	}

	return user, nil
}

func (r *userRepo) GetUserByID(ctx context.Context, id uuid.UUID) (domain.User, error) {
	query := `
		SELECT id, name, email, phone, password_hash, user_role
		FROM "user"
		WHERE id = $1
	`

	var user_role string // заглушка

	var user domain.User
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Phone,
		&user.PasswordHash,
		user_role,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.User{}, domain.ErrUserNotFound
		}
		return domain.User{}, err
	}

	return user, nil
}
