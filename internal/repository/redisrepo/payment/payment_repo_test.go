package payment

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	"github.com/gomodule/redigo/redis"
	"github.com/rafaeljusto/redigomock/v3"
	"github.com/stretchr/testify/assert"
)

func TestPaymentCacheRepo_SetPendingBinding(t *testing.T) {
	mock := redigomock.NewConn()
	pool := &redis.Pool{
		Dial: func() (redis.Conn, error) { return mock, nil },
	}
	repo := NewPaymentCacheRepo(pool)
	ctx := context.Background()

	tests := []struct {
		name      string
		paymentID string
		userID    int
		ttl       time.Duration
		setup     func()
		wantErr   error
	}{
		{
			name:      "Успешная установка значения",
			paymentID: "pay_123",
			userID:    1,
			ttl:       time.Minute,
			setup: func() {
				mock.Command("SET", "payment:pay_123", 1, "EX", 60).Expect("OK")
			},
			wantErr: nil,
		},
		{
			name:      "Ошибка Redis",
			paymentID: "pay_123",
			userID:    1,
			ttl:       time.Minute,
			setup: func() {
				mock.Command("SET", "payment:pay_123", 1, "EX", 60).ExpectError(errors.New("redis fail"))
			},
			wantErr: errors.New("redis fail"),
		},
		{
			name:      "Результат не OK",
			paymentID: "pay_123",
			userID:    1,
			ttl:       time.Minute,
			setup: func() {
				mock.Command("SET", "payment:pay_123", 1, "EX", 60).Expect("NOT_OK")
			},
			wantErr: domain.ErrRedisResultIsNotOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock.Clear()
			tt.setup()
			err := repo.SetPendingBinding(ctx, tt.paymentID, tt.userID, tt.ttl)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPaymentCacheRepo_DeletePendingBinding(t *testing.T) {
	mock := redigomock.NewConn()
	pool := &redis.Pool{
		Dial: func() (redis.Conn, error) { return mock, nil },
	}
	repo := NewPaymentCacheRepo(pool)
	ctx := context.Background()

	tests := []struct {
		name      string
		paymentID string
		setup     func()
		wantErr   bool
	}{
		{
			name:      "Успешное удаление",
			paymentID: "pay_123",
			setup: func() {
				mock.Command("DEL", "payment:pay_123").Expect(int64(1))
			},
			wantErr: false,
		},
		{
			name:      "Ошибка при удалении",
			paymentID: "pay_123",
			setup: func() {
				mock.Command("DEL", "payment:pay_123").ExpectError(errors.New("del fail"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock.Clear()
			tt.setup()
			err := repo.DeletePendingBinding(ctx, tt.paymentID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPaymentCacheRepo_GetUserIDByPaymentID(t *testing.T) {
	mock := redigomock.NewConn()
	pool := &redis.Pool{
		Dial: func() (redis.Conn, error) { return mock, nil },
	}
	repo := NewPaymentCacheRepo(pool)
	ctx := context.Background()

	tests := []struct {
		name      string
		paymentID string
		setup     func()
		want      int
		wantErr   bool
	}{
		{
			name:      "Успешное получение ID",
			paymentID: "pay_123",
			setup: func() {
				mock.Command("GET", "payment:pay_123").Expect(int64(10))
			},
			want:    10,
			wantErr: false,
		},
		{
			name:      "Ключ не найден",
			paymentID: "pay_none",
			setup: func() {
				mock.Command("GET", "payment:pay_none").Expect(nil)
			},
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock.Clear()
			tt.setup()
			got, err := repo.GetUserIDByPaymentID(ctx, tt.paymentID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
