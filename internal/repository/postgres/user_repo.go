package postgres

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/repository"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
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
		RETURNING id;
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
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation { // проверка на уникальность
			return 0, domain.ErrEmailAlreadyExists
		}
		return 0, err
	}

	return lastInsertedID, nil
}

func (r *userRepo) GetUserByEmail(ctx context.Context, email string) (domain.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	query := `
		SELECT id, name, email, phone, password_hash, user_role, avatar_url
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
		&user.AvatarURL,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, domain.ErrUserNotFound
		}
		return domain.User{}, err
	}

	return user, nil
}

func (r *userRepo) GetUserByID(ctx context.Context, id int) (domain.User, error) {
	query := `
		SELECT id, name, email, phone, password_hash, user_role, avatar_url
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
		&user.AvatarURL,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, domain.ErrUserNotFound
		}
		return domain.User{}, err
	}

	return user, nil
}

func (r *userRepo) CheckUserByID(ctx context.Context, userID int) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM "user" WHERE id = $1);`

	var isExists bool

	err := r.pool.QueryRow(ctx, query, userID).Scan(&isExists)
	if err != nil {
		return false, err
	}

	return isExists, nil
}

func (r *userRepo) UpdateProfile(ctx context.Context, userID int, name, email *string) error {
	query := `UPDATE "user" SET `
	var setClauses []string
	var args []any
	argID := 1 // значение для нумерации аргументов

	if name != nil {
		setClauses = append(setClauses, "name = $"+strconv.Itoa(argID))
		args = append(args, *name)
		argID++
	}

	if email != nil {
		setClauses = append(setClauses, "email = $"+strconv.Itoa(argID))
		args = append(args, *email)
		argID++
	}

	if argID == 1 {
		// кто-то отправил пустой запрос, это не очень хорошо
		return domain.ErrEmptyDBQuery
	}

	query += strings.Join(setClauses, ", ") + " WHERE id = $" + strconv.Itoa(argID)
	args = append(args, userID)

	tag, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case pgerrcode.UniqueViolation: // 23505
				switch pgErr.ConstraintName { // такую реализацию можно будет легко масштабировать, если решим менять какие-то другие unique поля
				case "user_email_key":
					return domain.ErrEmailAlreadyExists
				}
			case pgerrcode.CheckViolation:
				return domain.ErrInvalidInput
			case pgerrcode.SyntaxError:
				return domain.ErrSQLSyntax
			case pgerrcode.DeadlockDetected:
				return domain.ErrSQLDeadlock
			case pgerrcode.LockNotAvailable:
				return domain.ErrSQLLockTimeout
			default:
				return err
			}
		}
	}

	if tag.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}

func (r *userRepo) UpdateAvatarURL(ctx context.Context, userID int, newAvatarURL string) error {
	query := `UPDATE "user" SET "avatar_url" = $1 WHERE id = $2`

	tag, err := r.pool.Exec(ctx, query, newAvatarURL, userID)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case pgerrcode.CheckViolation:
				return domain.ErrInvalidInput
			case pgerrcode.SyntaxError:
				return domain.ErrSQLSyntax
			case pgerrcode.DeadlockDetected:
				return domain.ErrSQLDeadlock
			case pgerrcode.LockNotAvailable:
				return domain.ErrSQLLockTimeout
			default:
				return err
			}
		}
	}

	if tag.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}
