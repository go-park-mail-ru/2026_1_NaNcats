package redisrepo

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
	"github.com/mailru/easyjson"
	"github.com/rafaeljusto/redigomock/v3"
	"github.com/stretchr/testify/assert"
)

func TestSessionRepo_Create(t *testing.T) {
	mock := redigomock.NewConn()
	pool := &redis.Pool{
		Dial: func() (redis.Conn, error) { return mock, nil },
	}
	repo := NewSessionRepo(pool)
	ctx := context.Background()

	sessID := uuid.New()
	session := domain.Session{ID: sessID, UserID: 1}
	ttl := time.Hour
	expectedData, _ := easyjson.Marshal(session)
	key := "sessions:" + sessID.String()

	tests := []struct {
		name    string
		setup   func()
		wantErr error
	}{
		{
			name: "Успешное создание сессии",
			setup: func() {
				mock.Command("SET", key, expectedData, "EX", int(ttl.Seconds())).
					Expect("OK")
			},
			wantErr: nil,
		},
		{
			name: "Ошибка Redis при записи",
			setup: func() {
				mock.Command("SET", key, expectedData, "EX", int(ttl.Seconds())).
					ExpectError(errors.New("redis connection lost"))
			},
			wantErr: errors.New("redis connection lost"),
		},
		{
			name: "Redis вернул не OK",
			setup: func() {
				mock.Command("SET", key, expectedData, "EX", int(ttl.Seconds())).
					Expect("NOT_OK")
			},
			wantErr: domain.ErrRedisResultIsNotOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock.Clear()
			tt.setup()
			err := repo.Create(ctx, session, ttl)
			if tt.wantErr != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSessionRepo_GetByID(t *testing.T) {
	mock := redigomock.NewConn()
	pool := &redis.Pool{
		Dial: func() (redis.Conn, error) { return mock, nil },
	}
	repo := NewSessionRepo(pool)
	ctx := context.Background()

	sessID := uuid.New()
	session := domain.Session{ID: sessID, UserID: 1}
	validData, _ := easyjson.Marshal(session)

	tests := []struct {
		name    string
		id      uuid.UUID
		setup   func()
		want    domain.Session
		wantErr bool
	}{
		{
			name: "Сессия найдена",
			id:   sessID,
			setup: func() {
				mock.Command("GET", "sessions:"+sessID.String()).Expect(validData)
			},
			want:    session,
			wantErr: false,
		},
		{
			name: "Сессия не найдена",
			id:   sessID,
			setup: func() {
				mock.Command("GET", "sessions:"+sessID.String()).ExpectError(redis.ErrNil)
			},
			want:    domain.Session{},
			wantErr: true,
		},
		{
			name: "Ошибка десериализации (битые данные)",
			id:   sessID,
			setup: func() {
				mock.Command("GET", "sessions:"+sessID.String()).Expect([]byte("invalid json"))
			},
			want:    domain.Session{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock.Clear()
			tt.setup()
			got, err := repo.GetByID(ctx, tt.id)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want.ID, got.ID)
				assert.Equal(t, tt.want.UserID, got.UserID)
			}
		})
	}
}

func TestSessionRepo_Delete(t *testing.T) {
	mock := redigomock.NewConn()
	pool := &redis.Pool{
		Dial: func() (redis.Conn, error) { return mock, nil },
	}
	repo := NewSessionRepo(pool)
	ctx := context.Background()

	sessID := uuid.New()

	tests := []struct {
		name    string
		id      uuid.UUID
		setup   func()
		wantErr bool
	}{
		{
			name: "Успешное удаление",
			id:   sessID,
			setup: func() {
				mock.Command("DEL", "sessions:"+sessID.String()).Expect(int64(1))
			},
			wantErr: false,
		},
		{
			name: "Ошибка Redis при удалении",
			id:   sessID,
			setup: func() {
				mock.Command("DEL", "sessions:"+sessID.String()).ExpectError(errors.New("fail"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock.Clear()
			tt.setup()
			err := repo.Delete(ctx, tt.id)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
