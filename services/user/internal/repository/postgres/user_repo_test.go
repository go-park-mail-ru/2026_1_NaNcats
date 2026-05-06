package postgres

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/user/internal/domain"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pashagolub/pgxmock/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserRepo_CreateUser(t *testing.T) {
	ctx := context.Background()
	user := domain.User{
		Name:         "Ivan",
		Email:        "Ivan@Mail.ru",
		PasswordHash: "hashed_pass",
	}
	idemKey := "create-user-key"
	cleanEmail := "ivan@mail.ru"
	var expectedID int64 = 1

	type mockInit func(m pgxmock.PgxPoolIface)
	tests := []struct {
		name          string
		mockInit      mockInit
		expectedID    int64
		expectedError error
	}{
		{
			name: "Успешное создание пользователя и профиля",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()
				m.ExpectQuery(`INSERT INTO "user"`).
					WithArgs(user.Name, cleanEmail, user.PasswordHash, "client", idemKey).
					WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(expectedID))

				m.ExpectExec(`INSERT INTO "client_profile"`).
					WithArgs(expectedID, idemKey).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))

				m.ExpectCommit()
			},
			expectedID: expectedID,
		},
		{
			name: "Ошибка: Email уже занят",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()
				m.ExpectQuery(`INSERT INTO "user"`).
					WithArgs(user.Name, cleanEmail, user.PasswordHash, "client", idemKey).
					WillReturnError(&pgconn.PgError{Code: pgerrcode.UniqueViolation})
				m.ExpectRollback()
			},
			expectedID:    0,
			expectedError: domain.ErrEmailAlreadyExists,
		},
		{
			name: "Ошибка при создании профиля (Rollback)",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()
				m.ExpectQuery(`INSERT INTO "user"`).
					WithArgs(user.Name, cleanEmail, user.PasswordHash, "client", idemKey).
					WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(expectedID))

				m.ExpectExec(`INSERT INTO "client_profile"`).
					WithArgs(expectedID, idemKey).
					WillReturnError(errors.New("profile creation failed"))
				m.ExpectRollback()
			},
			expectedID:    0,
			expectedError: errors.New("profile creation failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mock.Close()

			repo := NewUserRepo(mock)
			tt.mockInit(mock)

			id, err := repo.CreateUser(ctx, user, idemKey)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedID, id)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUserRepo_GetUserByEmail(t *testing.T) {
	ctx := context.Background()
	email := "TEST@mail.ru"
	cleanEmail := "test@mail.ru"
	columns := []string{"id", "name", "email", "password_hash", "user_role", "avatar_url"}

	type mockInit func(m pgxmock.PgxPoolIface)
	tests := []struct {
		name          string
		mockInit      mockInit
		expectedUser  domain.User
		expectedError error
	}{
		{
			name: "Успешное получение",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(`SELECT (.+) FROM "user" WHERE email = \$1`).
					WithArgs(cleanEmail).
					WillReturnRows(pgxmock.NewRows(columns).
						AddRow(int64(10), "Ivan", cleanEmail, "hash", "client", "avatar.png"))
			},
			expectedUser: domain.User{
				ID:           10,
				Name:         "Ivan",
				Email:        cleanEmail,
				PasswordHash: "hash",
				AvatarURL:    "avatar.png",
			},
		},
		{
			name: "Пользователь не найден",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(`SELECT (.+) FROM "user" WHERE email = \$1`).
					WithArgs(cleanEmail).
					WillReturnError(pgx.ErrNoRows)
			},
			expectedError: domain.ErrUserNotFound,
		},
		{
			name: "Ошибка выполнения SQL",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(`SELECT (.+) FROM "user" WHERE email = \$1`).
					WithArgs(cleanEmail).
					WillReturnError(errors.New("database failure"))
			},
			expectedError: errors.New("database failure"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, _ := pgxmock.NewPool()
			defer mock.Close()

			repo := NewUserRepo(mock)
			tt.mockInit(mock)

			res, err := repo.GetUserByEmail(ctx, email)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedUser, res)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUserRepo_GetUserByID(t *testing.T) {
	ctx := context.Background()
	var userID int64 = 1
	columns := []string{"id", "name", "email", "password_hash", "user_role", "avatar_url"}

	type mockInit func(m pgxmock.PgxPoolIface)
	tests := []struct {
		name          string
		mockInit      mockInit
		expectedUser  domain.User
		expectedError error
	}{
		{
			name: "Успешное получение пользователя",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(`SELECT (.+) FROM "user" WHERE id = \$1`).
					WithArgs(userID).
					WillReturnRows(pgxmock.NewRows(columns).
						AddRow(userID, "Ivan", "test@mail.ru", "hash", "client", "avatar.png"))
			},
			expectedUser: domain.User{
				ID:           userID,
				Name:         "Ivan",
				Email:        "test@mail.ru",
				PasswordHash: "hash",
				AvatarURL:    "avatar.png",
			},
		},
		{
			name: "Пользователь не найден",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(`SELECT (.+) FROM "user"`).
					WithArgs(userID).
					WillReturnError(pgx.ErrNoRows)
			},
			expectedError: domain.ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, _ := pgxmock.NewPool()
			defer mock.Close()

			repo := NewUserRepo(mock)
			tt.mockInit(mock)

			res, err := repo.GetUserByID(ctx, userID)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedUser, res)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUserRepo_CheckUserByID(t *testing.T) {
	ctx := context.Background()
	var userID int64 = 42

	type mockInit func(m pgxmock.PgxPoolIface)
	tests := []struct {
		name     string
		mockInit mockInit
		expected bool
		wantErr  bool
	}{
		{
			name: "Пользователь существует",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(`SELECT EXISTS`).
					WithArgs(userID).
					WillReturnRows(pgxmock.NewRows([]string{"exists"}).AddRow(true))
			},
			expected: true,
		},
		{
			name: "Пользователь не существует",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(`SELECT EXISTS`).
					WithArgs(userID).
					WillReturnRows(pgxmock.NewRows([]string{"exists"}).AddRow(false))
			},
			expected: false,
		},
		{
			name: "Ошибка выполнения запроса",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectQuery(`SELECT EXISTS`).WillReturnError(errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, _ := pgxmock.NewPool()
			defer mock.Close()

			repo := NewUserRepo(mock)
			tt.mockInit(mock)

			res, err := repo.CheckUserByID(ctx, userID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, res)
			}
		})
	}
}

func TestUserRepo_UpdateProfile(t *testing.T) {
	ctx := context.Background()
	var userID int64 = 1
	newName := "NewName"
	newEmail := "new@mail.ru"

	type mockInit func(m pgxmock.PgxPoolIface)
	tests := []struct {
		name          string
		uName         *string
		uEmail        *string
		mockInit      mockInit
		expectedError error
	}{
		{
			name:   "Успешное обновление всех полей",
			uName:  &newName,
			uEmail: &newEmail,
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectExec(`UPDATE "user" SET name = \$1, email = \$2 WHERE id = \$3`).
					WithArgs(newName, newEmail, userID).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
			},
		},
		{
			name:   "Ошибка: Email уже существует",
			uName:  nil,
			uEmail: &newEmail,
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectExec(`UPDATE "user" SET email = \$1 WHERE id = \$2`).
					WithArgs(newEmail, userID).
					WillReturnError(&pgconn.PgError{
						Code:           pgerrcode.UniqueViolation,
						ConstraintName: "user_email_key",
					})
			},
			expectedError: domain.ErrEmailAlreadyExists,
		},
		{
			name:   "Ошибка: нарушение check-constraint",
			uName:  &newName,
			uEmail: nil,
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectExec(`UPDATE "user" SET name = \$1 WHERE id = \$2`).
					WithArgs(newName, userID).
					WillReturnError(&pgconn.PgError{Code: pgerrcode.CheckViolation})
			},
			expectedError: domain.ErrInvalidInput,
		},
		{
			name:   "Пользователь не найден (0 строк затронуто)",
			uName:  &newName,
			uEmail: nil,
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectExec(`UPDATE "user" SET name = \$1 WHERE id = \$2`).
					WithArgs(newName, userID).
					WillReturnResult(pgxmock.NewResult("UPDATE", 0))
			},
			expectedError: domain.ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, _ := pgxmock.NewPool()
			defer mock.Close()

			repo := NewUserRepo(mock)
			tt.mockInit(mock)

			err := repo.UpdateProfile(ctx, userID, tt.uName, tt.uEmail)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUserRepo_UpdateAvatarURL(t *testing.T) {
	ctx := context.Background()
	var userID int64 = 1
	newAvatar := "https://s3.ru/avatars/new.webp"

	type mockInit func(m pgxmock.PgxPoolIface)
	tests := []struct {
		name          string
		mockInit      mockInit
		expectedError error
	}{
		{
			name: "Успешное обновление аватара",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectExec(`UPDATE "user" SET "avatar_url" = \$1 WHERE id = \$2`).
					WithArgs(newAvatar, userID).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
			},
		},
		{
			name: "Ошибка: нарушение check-constraint (слишком длинный URL)",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectExec(`UPDATE "user" SET "avatar_url" = \$1 WHERE id = \$2`).
					WithArgs(newAvatar, userID).
					WillReturnError(&pgconn.PgError{Code: pgerrcode.CheckViolation})
			},
			expectedError: domain.ErrInvalidInput,
		},
		{
			name: "Неизвестная ошибка БД",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectExec(`UPDATE "user"`).
					WithArgs(newAvatar, userID).
					WillReturnError(errors.New("fatal db error"))
			},
			expectedError: errors.New("fatal db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, _ := pgxmock.NewPool()
			defer mock.Close()

			repo := NewUserRepo(mock)
			tt.mockInit(mock)

			err := repo.UpdateAvatarURL(ctx, userID, newAvatar)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUserRepo_UpdateUserRole(t *testing.T) {
	ctx := context.Background()
	var userID int64 = 10
	newRole := "courier"
	idemKey := "role-change-key"

	type mockInit func(m pgxmock.PgxPoolIface)
	tests := []struct {
		name           string
		mockInit       mockInit
		expectedOld    string
		expectedNotify bool
		expectedError  string
	}{
		{
			name: "Успешная смена роли (Client -> Courier)",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()
				m.ExpectQuery(regexp.QuoteMeta(`SELECT user_role, idempotency_key FROM "user" WHERE id = $1 FOR UPDATE`)).
					WithArgs(userID).
					WillReturnRows(pgxmock.NewRows([]string{"user_role", "idempotency_key"}).AddRow("client", nil))

				m.ExpectExec(regexp.QuoteMeta(`DELETE FROM "courier_profile" WHERE account_id = $1`)).
					WithArgs(userID).WillReturnResult(pgxmock.NewResult("DELETE", 0))
				m.ExpectExec(regexp.QuoteMeta(`DELETE FROM "owner_profile" WHERE account_id = $1`)).
					WithArgs(userID).WillReturnResult(pgxmock.NewResult("DELETE", 0))
				m.ExpectExec(regexp.QuoteMeta(`DELETE FROM "client_profile" WHERE account_id = $1`)).
					WithArgs(userID).WillReturnResult(pgxmock.NewResult("DELETE", 1))

				m.ExpectExec(regexp.QuoteMeta(`UPDATE "user" SET user_role = $1, idempotency_key = $2, updated_at = NOW() WHERE id = $3`)).
					WithArgs(newRole, idemKey, userID).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))

				m.ExpectExec(regexp.QuoteMeta(`INSERT INTO "courier_profile" (account_id, status) VALUES ($1, 'offline')`)).
					WithArgs(userID).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))

				m.ExpectCommit()
			},
			expectedOld:    "client",
			expectedNotify: true,
		},
		{
			name: "Кейс идемпотентности: тот же ключ (без коммита, выход раньше)",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()
				m.ExpectQuery(regexp.QuoteMeta(`SELECT user_role, idempotency_key FROM "user" WHERE id = $1 FOR UPDATE`)).
					WithArgs(userID).
					WillReturnRows(pgxmock.NewRows([]string{"user_role", "idempotency_key"}).AddRow("courier", &idemKey))
				m.ExpectRollback()
			},
			expectedOld:    "courier",
			expectedNotify: false,
		},
		{
			name: "Ошибка: пользователь не найден",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()
				m.ExpectQuery(regexp.QuoteMeta(`SELECT user_role, idempotency_key FROM "user" WHERE id = $1 FOR UPDATE`)).
					WithArgs(userID).
					WillReturnError(pgx.ErrNoRows)
				m.ExpectRollback()
			},
			expectedError: "no rows in result set",
		},
		{
			name: "Ошибка при очистке профилей (откат транзакции)",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()
				m.ExpectQuery(regexp.QuoteMeta(`SELECT user_role, idempotency_key FROM "user" WHERE id = $1 FOR UPDATE`)).
					WithArgs(userID).
					WillReturnRows(pgxmock.NewRows([]string{"role", "key"}).AddRow("owner", nil))

				// Эмулируем сбой на первом же DELETE
				m.ExpectExec(regexp.QuoteMeta(`DELETE FROM "courier_profile" WHERE account_id = $1`)).
					WithArgs(userID).
					WillReturnError(errors.New("db cleanup error"))

				m.ExpectRollback()
			},
			expectedError: "failed to cleanup profile in courier_profile",
		},
		{
			name: "Успешная смена роли на Owner",
			mockInit: func(m pgxmock.PgxPoolIface) {
				m.ExpectBegin()
				m.ExpectQuery(regexp.QuoteMeta(`SELECT user_role, idempotency_key FROM "user" WHERE id = $1 FOR UPDATE`)).
					WithArgs(userID).
					WillReturnRows(pgxmock.NewRows([]string{"role", "key"}).AddRow("client", nil))

				// Цикл очистки
				m.ExpectExec(`DELETE FROM "courier_profile"`).WithArgs(userID).WillReturnResult(pgxmock.NewResult("DELETE", 0))
				m.ExpectExec(`DELETE FROM "owner_profile"`).WithArgs(userID).WillReturnResult(pgxmock.NewResult("DELETE", 0))
				m.ExpectExec(`DELETE FROM "client_profile"`).WithArgs(userID).WillReturnResult(pgxmock.NewResult("DELETE", 1))

				// Апдейт роли на owner
				m.ExpectExec(`UPDATE "user" SET user_role = \$1`).
					WithArgs("owner", idemKey, userID).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))

				// Создание профиля владельца
				m.ExpectExec(regexp.QuoteMeta(`INSERT INTO "owner_profile" (account_id) VALUES ($1)`)).
					WithArgs(userID).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))

				m.ExpectCommit()
			},
			expectedOld:    "client",
			expectedNotify: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, _ := pgxmock.NewPool()
			defer mock.Close()

			repo := NewUserRepo(mock)
			tt.mockInit(mock)

			// Для случая с owner меняем входной аргумент в самом тесте
			currentNewRole := newRole
			if tt.name == "Успешная смена роли на Owner" {
				currentNewRole = "owner"
			}

			oldRole, notify, err := repo.UpdateUserRole(ctx, userID, currentNewRole, idemKey)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedOld, oldRole)
				assert.Equal(t, tt.expectedNotify, notify)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
