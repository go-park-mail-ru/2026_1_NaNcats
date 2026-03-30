package postgres

import (
	"context"
	"errors"
	"strings"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type userRepo struct {
	pool   *pgxpool.Pool
	logger domain.Logger
}

func NewUserRepo(pool *pgxpool.Pool, logger domain.Logger) repository.UserRepository {
	return &userRepo{
		pool:   pool,
		logger: logger,
	}
}

func (r *userRepo) CreateUser(ctx context.Context, user domain.User) (int, error) {
	l := r.logger.WithContext(ctx)

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
			l.Info("email already exists", map[string]any{
				"query": "CreateUser",
				"email": user.Email,
			})
			return 0, domain.ErrEmailAlreadyExists
		}
		l.Error("database query execution failed", err, map[string]any{
			"query": "CreateUser",
			"email": user.Email,
		})
		return 0, err
	}

	l.Info("database trip successful", map[string]any{
		"query": "GetUserByEmail",
		"email": user.Email,
	})

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

	l := r.logger.WithContext(ctx)

	err := r.pool.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Phone,
		&user.PasswordHash,
		&userRole,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			l.Info("user not found in database", map[string]any{
				"query": "GetUserByEmail",
				"email": email,
			})
			return domain.User{}, domain.ErrUserNotFound
		}
		l.Error("query", err, map[string]any{
			"query": "GetUserByEmail",
			"email": email,
		})
		return domain.User{}, err
	}

	l.Info("database trip successful", map[string]any{
		"query": "GetUserByEmail",
		"email": email,
	})

	return user, nil
}

func (r *userRepo) GetUserByID(ctx context.Context, id int) (domain.User, error) {
	l := r.logger.WithContext(ctx)

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
		if errors.Is(err, pgx.ErrNoRows) {
			l.Info("user not found in database", map[string]any{
				"query":   "GetUserByID",
				"user_id": id,
			})
			return domain.User{}, domain.ErrUserNotFound
		}
		l.Error("database query execution failed", err, map[string]any{
			"query": "GetUserById",
			"id":    id,
		})
		return domain.User{}, err
	}

	l.Info("database trip successful", map[string]any{
		"query": "GetUserById",
		"id":    id,
	})

	return user, nil
}
