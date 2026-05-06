package redisrepo

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/auth/internal/domain"
	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
	"github.com/mailru/easyjson"
	"github.com/rafaeljusto/redigomock/v3"
	"github.com/stretchr/testify/assert"
)

func setupMockPool(conn *redigomock.Conn) *redis.Pool {
	return &redis.Pool{
		Dial: func() (redis.Conn, error) {
			return conn, nil
		},
	}
}

func TestSessionRepo_Create(t *testing.T) {
	sessionID := uuid.New()
	session := domain.Session{
		ID:        sessionID,
		UserID:    1,
		Role:      "user",
		UserAgent: "Mozilla",
		ExpiresAt: time.Now().Add(time.Hour),
	}
	ttl := time.Hour
	serialized, _ := easyjson.Marshal(session)
	mkey := "sessions:" + sessionID.String()

	tests := []struct {
		name         string
		mockBehavior func(conn *redigomock.Conn)
		expectedErr  bool
	}{
		{
			name: "Успешное создание сессии",
			mockBehavior: func(conn *redigomock.Conn) {
				conn.Command("SET", mkey, serialized, "EX", int(ttl.Seconds())).
					Expect("OK")
			},
			expectedErr: false,
		},
		{
			name: "Ошибка Redis при выполнении SET",
			mockBehavior: func(conn *redigomock.Conn) {
				conn.Command("SET", mkey, serialized, "EX", int(ttl.Seconds())).
					ExpectError(errors.New("redis connection error"))
			},
			expectedErr: true,
		},
		{
			name: "Redis вернул результат отличный от OK",
			mockBehavior: func(conn *redigomock.Conn) {
				conn.Command("SET", mkey, serialized, "EX", int(ttl.Seconds())).
					Expect("NOT_OK")
			},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockConn := redigomock.NewConn()
			pool := setupMockPool(mockConn)
			repo := NewSessionRepo(pool)

			tt.mockBehavior(mockConn)

			err := repo.Create(context.Background(), session, ttl)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSessionRepo_GetByID(t *testing.T) {
	sessionID := uuid.New()
	mkey := "sessions:" + sessionID.String()
	session := domain.Session{ID: sessionID, UserID: 1}
	serialized, _ := easyjson.Marshal(session)

	tests := []struct {
		name          string
		mockBehavior  func(conn *redigomock.Conn)
		expectedSess  domain.Session
		expectedErrIs error
	}{
		{
			name: "Успешное получение сессии",
			mockBehavior: func(conn *redigomock.Conn) {
				conn.Command("GET", mkey).Expect(serialized)
			},
			expectedSess:  session,
			expectedErrIs: nil,
		},
		{
			name: "Сессия не найдена (ErrNil)",
			mockBehavior: func(conn *redigomock.Conn) {
				conn.Command("GET", mkey).ExpectError(redis.ErrNil)
			},
			expectedSess:  domain.Session{},
			expectedErrIs: domain.ErrSessionNotFound,
		},
		{
			name: "Ошибка десериализации (битые данные)",
			mockBehavior: func(conn *redigomock.Conn) {
				conn.Command("GET", mkey).Expect([]byte("invalid json"))
			},
			expectedSess:  domain.Session{},
			expectedErrIs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockConn := redigomock.NewConn()
			pool := setupMockPool(mockConn)
			repo := NewSessionRepo(pool)

			tt.mockBehavior(mockConn)

			sess, err := repo.GetByID(context.Background(), sessionID)

			if tt.expectedErrIs != nil {
				assert.ErrorIs(t, err, tt.expectedErrIs)
			} else if tt.name == "Ошибка десериализации (битые данные)" {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedSess.ID, sess.ID)
			}
		})
	}
}

func TestSessionRepo_Delete(t *testing.T) {
	sessionID := uuid.New()
	mkey := "sessions:" + sessionID.String()

	tests := []struct {
		name         string
		mockBehavior func(conn *redigomock.Conn)
		expectedErr  bool
	}{
		{
			name: "Успешное удаление",
			mockBehavior: func(conn *redigomock.Conn) {
				conn.Command("DEL", mkey).Expect(int64(1))
			},
			expectedErr: false,
		},
		{
			name: "Ошибка Redis при удалении",
			mockBehavior: func(conn *redigomock.Conn) {
				conn.Command("DEL", mkey).ExpectError(errors.New("redis error"))
			},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockConn := redigomock.NewConn()
			pool := setupMockPool(mockConn)
			repo := NewSessionRepo(pool)

			tt.mockBehavior(mockConn)

			err := repo.Delete(context.Background(), sessionID)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSessionRepo_SetCSRF(t *testing.T) {
	sessionID := uuid.New()
	token := "csrf-token-123"
	mkey := "csrf:" + sessionID.String()

	tests := []struct {
		name         string
		mockBehavior func(conn *redigomock.Conn)
		expectedErr  bool
	}{
		{
			name: "Успешная установка CSRF",
			mockBehavior: func(conn *redigomock.Conn) {
				conn.Command("SET", mkey, token, "EX", CSRFTokenTTL).
					Expect("OK")
			},
			expectedErr: false,
		},
		{
			name: "Redis вернул не OK",
			mockBehavior: func(conn *redigomock.Conn) {
				conn.Command("SET", mkey, token, "EX", CSRFTokenTTL).
					Expect("FAIL")
			},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockConn := redigomock.NewConn()
			pool := setupMockPool(mockConn)
			repo := NewSessionRepo(pool)

			tt.mockBehavior(mockConn)

			err := repo.SetCSRF(context.Background(), sessionID, token)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSessionRepo_GetCSRF(t *testing.T) {
	sessionID := uuid.New()
	mkey := "csrf:" + sessionID.String()
	token := "token-val"

	tests := []struct {
		name          string
		mockBehavior  func(conn *redigomock.Conn)
		expectedToken string
		expectedErr   bool
	}{
		{
			name: "Успешное получение CSRF",
			mockBehavior: func(conn *redigomock.Conn) {
				conn.Command("GET", mkey).Expect(token)
			},
			expectedToken: token,
			expectedErr:   false,
		},
		{
			name: "Токен не найден (нормальное поведение)",
			mockBehavior: func(conn *redigomock.Conn) {
				conn.Command("GET", mkey).ExpectError(redis.ErrNil)
			},
			expectedToken: "",
			expectedErr:   false,
		},
		{
			name: "Ошибка Redis",
			mockBehavior: func(conn *redigomock.Conn) {
				conn.Command("GET", mkey).ExpectError(errors.New("redis fail"))
			},
			expectedToken: "",
			expectedErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockConn := redigomock.NewConn()
			pool := setupMockPool(mockConn)
			repo := NewSessionRepo(pool)

			tt.mockBehavior(mockConn)

			result, err := repo.GetCSRF(context.Background(), sessionID)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedToken, result)
			}
		})
	}
}
