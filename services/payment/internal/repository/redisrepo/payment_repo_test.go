package redisrepo

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/payment/internal/domain"
	"github.com/gomodule/redigo/redis"
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

func TestPaymentCacheRepo_SetPendingBinding(t *testing.T) {
	type mockBehavior func(conn *redigomock.Conn, paymentID string, userID int64, ttl time.Duration)

	tests := []struct {
		name         string
		paymentID    string
		userID       int64
		ttl          time.Duration
		mockBehavior mockBehavior
		expectedErr  error
	}{
		{
			name:      "Успешное сохранение биндинга",
			paymentID: "pay-123",
			userID:    10,
			ttl:       60 * time.Second,
			mockBehavior: func(conn *redigomock.Conn, paymentID string, userID int64, ttl time.Duration) {
				mkey := "payment:" + paymentID
				conn.Command("SET", mkey, userID, "EX", int(ttl.Seconds())).Expect("OK")
			},
			expectedErr: nil,
		},
		{
			name:      "Ошибка сети Redis",
			paymentID: "pay-123",
			userID:    10,
			ttl:       60 * time.Second,
			mockBehavior: func(conn *redigomock.Conn, paymentID string, userID int64, ttl time.Duration) {
				mkey := "payment:" + paymentID
				conn.Command("SET", mkey, userID, "EX", int(ttl.Seconds())).ExpectError(errors.New("redis timeout"))
			},
			expectedErr: errors.New("redis timeout"),
		},
		{
			name:      "Redis вернул не OK",
			paymentID: "pay-123",
			userID:    10,
			ttl:       60 * time.Second,
			mockBehavior: func(conn *redigomock.Conn, paymentID string, userID int64, ttl time.Duration) {
				mkey := "payment:" + paymentID
				conn.Command("SET", mkey, userID, "EX", int(ttl.Seconds())).Expect("FAIL")
			},
			expectedErr: domain.ErrRedisResultIsNotOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockConn := redigomock.NewConn()
			pool := setupMockPool(mockConn)

			tt.mockBehavior(mockConn, tt.paymentID, tt.userID, tt.ttl)

			repo := NewPaymentCacheRepo(pool)
			err := repo.SetPendingBinding(context.Background(), tt.paymentID, tt.userID, tt.ttl)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				if errors.Is(tt.expectedErr, domain.ErrRedisResultIsNotOK) {
					assert.ErrorIs(t, err, domain.ErrRedisResultIsNotOK)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPaymentCacheRepo_DeletePendingBinding(t *testing.T) {
	type mockBehavior func(conn *redigomock.Conn, paymentID string)

	tests := []struct {
		name         string
		paymentID    string
		mockBehavior mockBehavior
		expectedErr  bool
	}{
		{
			name:      "Успешное удаление",
			paymentID: "pay-123",
			mockBehavior: func(conn *redigomock.Conn, paymentID string) {
				mkey := "payment:" + paymentID
				conn.Command("DEL", mkey).Expect(int64(1))
			},
			expectedErr: false,
		},
		{
			name:      "Ошибка при удалении",
			paymentID: "pay-err",
			mockBehavior: func(conn *redigomock.Conn, paymentID string) {
				mkey := "payment:" + paymentID
				conn.Command("DEL", mkey).ExpectError(errors.New("redis connection reset"))
			},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockConn := redigomock.NewConn()
			pool := setupMockPool(mockConn)

			tt.mockBehavior(mockConn, tt.paymentID)

			repo := NewPaymentCacheRepo(pool)
			err := repo.DeletePendingBinding(context.Background(), tt.paymentID)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPaymentCacheRepo_GetUserIDByPaymentID(t *testing.T) {
	type mockBehavior func(conn *redigomock.Conn, paymentID string)

	tests := []struct {
		name         string
		paymentID    string
		mockBehavior mockBehavior
		expectedID   int64
		expectedErr  error
	}{
		{
			name:      "Успешное получение ID пользователя",
			paymentID: "pay-123",
			mockBehavior: func(conn *redigomock.Conn, paymentID string) {
				mkey := "payment:" + paymentID
				conn.Command("GET", mkey).Expect(int64(42))
			},
			expectedID:  42,
			expectedErr: nil,
		},
		{
			name:      "Ошибка: ключ не найден (Nil)",
			paymentID: "pay-404",
			mockBehavior: func(conn *redigomock.Conn, paymentID string) {
				mkey := "payment:" + paymentID
				conn.Command("GET", mkey).ExpectError(redis.ErrNil)
			},
			expectedID:  0,
			expectedErr: domain.ErrUserIDNotFoundInCache,
		},
		{
			name:      "Системная ошибка Redis",
			paymentID: "pay-err",
			mockBehavior: func(conn *redigomock.Conn, paymentID string) {
				mkey := "payment:" + paymentID
				conn.Command("GET", mkey).ExpectError(errors.New("connection failed"))
			},
			expectedID:  0,
			expectedErr: errors.New("connection failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockConn := redigomock.NewConn()
			pool := setupMockPool(mockConn)

			tt.mockBehavior(mockConn, tt.paymentID)

			repo := NewPaymentCacheRepo(pool)
			id, err := repo.GetUserIDByPaymentID(context.Background(), tt.paymentID)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				if errors.Is(tt.expectedErr, domain.ErrUserIDNotFoundInCache) {
					assert.ErrorIs(t, err, domain.ErrUserIDNotFoundInCache)
				}
				assert.Equal(t, tt.expectedID, id)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedID, id)
			}
		})
	}
}
