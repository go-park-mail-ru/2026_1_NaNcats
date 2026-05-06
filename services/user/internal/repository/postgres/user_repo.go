package postgres

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/user/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/user/internal/repository"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/postgres"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type userRepo struct {
	pool postgres.PgxPool
}

func NewUserRepo(pool postgres.PgxPool) repository.UserRepository {
	return &userRepo{
		pool: pool,
	}
}

func (r *userRepo) CreateUser(ctx context.Context, user domain.User, idempotencyKey string) (int64, error) {
	user.Email = strings.ToLower(strings.TrimSpace(user.Email))

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	queryUser := `
		INSERT INTO "user" (name, email, password_hash, user_role, idempotency_key)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (idempotency_key) DO UPDATE 
		SET idempotency_key = EXCLUDED.idempotency_key
		RETURNING id;
	`

	queryClient := `
		INSERT INTO "client_profile" (account_id, idempotency_key)
		VALUES ($1, $2)
		ON CONFLICT (idempotency_key) DO UPDATE 
		SET idempotency_key = EXCLUDED.idempotency_key
	`

	var lastInsertedID int64
	err = tx.QueryRow(ctx, queryUser,
		user.Name,
		user.Email,
		user.PasswordHash,
		"client",
		idempotencyKey,
	).Scan(&lastInsertedID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation { // проверка на уникальность
			return 0, domain.ErrEmailAlreadyExists
		}
		return 0, err
	}

	_, err = tx.Exec(ctx, queryClient,
		lastInsertedID,
		idempotencyKey,
	)
	if err != nil {
		return 0, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return 0, err
	}

	return lastInsertedID, nil
}

func (r *userRepo) GetUserByEmail(ctx context.Context, email string) (domain.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	query := `
		SELECT id, name, email, password_hash, user_role, COALESCE(avatar_url, '')
		FROM "user"
		WHERE email = $1
	`

	var user domain.User
	var userRole string

	err := r.pool.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
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

func (r *userRepo) GetUserByID(ctx context.Context, id int64) (domain.User, error) {
	query := `
		SELECT id, name, email, password_hash, user_role, COALESCE(avatar_url, '')
		FROM "user"
		WHERE id = $1
	`

	var user domain.User
	var userRole string

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
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

func (r *userRepo) CheckUserByID(ctx context.Context, userID int64) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM "user" WHERE id = $1);`

	var isExists bool

	err := r.pool.QueryRow(ctx, query, userID).Scan(&isExists)
	if err != nil {
		return false, err
	}

	return isExists, nil
}

func (r *userRepo) UpdateProfile(ctx context.Context, userID int64, name, email *string) error {
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
		return domain.ErrNoChangesProvided
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
			default:
				return err
			}
		}
		return err
	}

	if tag.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}

func (r *userRepo) UpdateAvatarURL(ctx context.Context, userID int64, newAvatarURL string) error {
	query := `UPDATE "user" SET "avatar_url" = $1 WHERE id = $2`

	tag, err := r.pool.Exec(ctx, query, newAvatarURL, userID)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case pgerrcode.CheckViolation:
				return domain.ErrInvalidInput
			default:
				return err
			}
		}

		return err
	}

	if tag.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}

func (r *userRepo) UpdateUserRole(ctx context.Context, userID int64, newRole string, idempotencyKey string) (string, bool, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return "", false, err
	}
	defer tx.Rollback(ctx)

	// Достаем текущую роль и последний использованный ключ этого юзера
	var currentRole string
	var lastKey *string
	queryCheck := `SELECT user_role, idempotency_key FROM "user" WHERE id = $1 FOR UPDATE`
	err = tx.QueryRow(ctx, queryCheck, userID).Scan(&currentRole, &lastKey)
	if err != nil {
		return "", false, err
	}

	if lastKey != nil && *lastKey == idempotencyKey {
		return currentRole, false, nil
	}

	// Очистка старых профилей
	profileTables := []string{"courier_profile", "owner_profile", "client_profile"}
	for _, table := range profileTables {
		query := fmt.Sprintf(`DELETE FROM "%s" WHERE account_id = $1`, table)
		_, err := tx.Exec(ctx, query, userID)
		if err != nil {
			return "", false, fmt.Errorf("failed to cleanup profile in %s: %w", table, err)
		}
	}

	// Обновление роли и запись нового ключа
	queryUpdate := `
		UPDATE "user" 
		SET user_role = $1, 
		    idempotency_key = $2, 
		    updated_at = NOW() 
		WHERE id = $3`
	_, err = tx.Exec(ctx, queryUpdate, newRole, idempotencyKey, userID)
	if err != nil {
		return "", false, err
	}

	// Создание нового профиля
	if newRole == "courier" {
		_, err = tx.Exec(ctx, `INSERT INTO "courier_profile" (account_id, status) VALUES ($1, 'offline')`, userID)
	} else if newRole == "owner" {
		_, err = tx.Exec(ctx, `INSERT INTO "owner_profile" (account_id) VALUES ($1)`, userID)
	}

	if err != nil {
		return "", false, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return "", false, err
	}

	// Возвращаем старую роль и флаг true (нужно отправить событие)
	return currentRole, true, nil
}
