package postgres

import (
	"context"
	"errors"
	"testing"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pashagolub/pgxmock/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserRepo_CreateUser(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewUserRepo(mock)
	ctx := context.Background()

	tests := []struct {
		name    string
		user    domain.User
		mock    func()
		wantID  int
		wantErr error
	}{
		{
			name: "Успех",
			user: domain.User{Name: "Ivan", Email: "TEST@mail.ru", Phone: "7999", PasswordHash: "hash"},
			mock: func() {
				mock.ExpectBegin()

				mock.ExpectQuery(`INSERT INTO "user"`).
					WithArgs("Ivan", "test@mail.ru", "7999", "hash", "client").
					WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(1))

				mock.ExpectExec(`INSERT INTO "client_profile"`).
					WithArgs(1).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))

				mock.ExpectCommit()
			},
			wantID:  1,
			wantErr: nil,
		},
		{
			name: "Почта уже существует",
			user: domain.User{Email: "exists@mail.ru"},
			mock: func() {
				mock.ExpectBegin()

				mock.ExpectQuery(`INSERT INTO "user"`).
					WithArgs(pgxmock.AnyArg(), "exists@mail.ru", pgxmock.AnyArg(), pgxmock.AnyArg(), "client").
					WillReturnError(&pgconn.PgError{Code: pgerrcode.UniqueViolation})

				mock.ExpectRollback()
			},
			wantID:  0,
			wantErr: domain.ErrEmailAlreadyExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			id, err := repo.CreateUser(ctx, tt.user)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.wantID, id)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUserRepo_GetUserByEmail(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewUserRepo(mock)
	ctx := context.Background()

	tests := []struct {
		name    string
		email   string
		mock    func()
		wantRes domain.User
		wantErr error
	}{
		{
			name:  "Успех",
			email: "test@mail.ru",
			mock: func() {
				rows := pgxmock.NewRows([]string{"id", "name", "email", "phone", "password_hash", "user_role", "avatar_url"}).
					AddRow(1, "Ivan", "test@mail.ru", "7999", "hash", "client", "url")
				mock.ExpectQuery(`SELECT (.+) FROM "user" WHERE email = \$1`).
					WithArgs("test@mail.ru").
					WillReturnRows(rows)
			},
			wantRes: domain.User{ID: 1, Name: "Ivan", Email: "test@mail.ru", Phone: "7999", PasswordHash: "hash", AvatarURL: "url"},
			wantErr: nil,
		},
		{
			name:  "Ошибка: юзер не найден",
			email: "unknown@mail.ru",
			mock: func() {
				mock.ExpectQuery(`SELECT`).
					WithArgs("unknown@mail.ru").
					WillReturnError(pgx.ErrNoRows)
			},
			wantRes: domain.User{},
			wantErr: domain.ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			user, err := repo.GetUserByEmail(ctx, tt.email)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.wantRes, user)
		})
	}
}

func TestUserRepo_UpdateProfile(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewUserRepo(mock)
	ctx := context.Background()

	nameVal := "NewName"
	emailVal := "new@mail.ru"

	tests := []struct {
		name    string
		userID  int
		uName   *string
		uEmail  *string
		mock    func()
		wantErr error
	}{
		{
			name:   "Успех обновления имени",
			userID: 1,
			uName:  &nameVal,
			mock: func() {
				mock.ExpectExec(`UPDATE "user" SET name = \$1 WHERE id = \$2`).
					WithArgs(nameVal, 1).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
			},
			wantErr: nil,
		},
		{
			name:    "Ошибка: пустое обновление",
			userID:  1,
			uName:   nil,
			uEmail:  nil,
			mock:    func() {},
			wantErr: domain.ErrEmptyDBQuery,
		},
		{
			name:   "Ошибка: email занят",
			userID: 1,
			uEmail: &emailVal,
			mock: func() {
				mock.ExpectExec(`UPDATE "user"`).
					WithArgs(emailVal, 1).
					WillReturnError(&pgconn.PgError{
						Code:           pgerrcode.UniqueViolation,
						ConstraintName: "user_email_key",
					})
			},
			wantErr: domain.ErrEmailAlreadyExists,
		},
		{
			name:   "Ошибка: юзер не найден",
			userID: 404,
			uName:  &nameVal,
			mock: func() {
				mock.ExpectExec(`UPDATE "user"`).
					WithArgs(nameVal, 404).
					WillReturnResult(pgxmock.NewResult("UPDATE", 0))
			},
			wantErr: domain.ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			err := repo.UpdateProfile(ctx, tt.userID, tt.uName, tt.uEmail)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestUserRepo_CheckUserByID(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	repo := NewUserRepo(mock)
	ctx := context.Background()

	tests := []struct {
		name    string
		userID  int
		mock    func()
		want    bool
		wantErr bool
	}{
		{
			name:   "Успех",
			userID: 1,
			mock: func() {
				mock.ExpectQuery(`SELECT EXISTS`).WithArgs(1).
					WillReturnRows(pgxmock.NewRows([]string{"exists"}).AddRow(true))
			},
			want:    true,
			wantErr: false,
		},
		{
			name:   "Ошибка БД",
			userID: 1,
			mock: func() {
				mock.ExpectQuery(`SELECT EXISTS`).WillReturnError(errors.New("db error"))
			},
			want:    false,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			got, err := repo.CheckUserByID(ctx, tt.userID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
