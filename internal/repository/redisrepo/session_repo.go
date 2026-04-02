package redisrepo

import (
	"context"
	"fmt"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/repository"
	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
	"github.com/mailru/easyjson"
)

type sessionRepo struct {
	redisPool *redis.Pool
}

func NewSessionRepo(pool *redis.Pool) repository.SessionRepository {
	return &sessionRepo{
		redisPool: pool,
	}
}

func (r *sessionRepo) Create(ctx context.Context, session domain.Session, ttl time.Duration) error {
	dataSerializer, err := easyjson.Marshal(session)
	if err != nil {
		return err
	}

	conn := r.redisPool.Get()
	defer conn.Close()

	mkey := "sessions:" + session.ID.String()
	result, err := redis.String(conn.Do("SET", mkey, dataSerializer, "EX", int(ttl.Seconds())))
	if err != nil {
		return err
	}
	if result != "OK" {
		return fmt.Errorf("result is not OK")
	}

	return nil
}

func (r *sessionRepo) GetByID(ctx context.Context, id uuid.UUID) (domain.Session, error) {
	conn := r.redisPool.Get()
	defer conn.Close()

	mkey := "sessions:" + id.String()
	data, err := redis.Bytes(conn.Do("GET", mkey))
	if err != nil {
		return domain.Session{}, err
	}

	session := &domain.Session{}
	err = easyjson.Unmarshal(data, session)
	if err != nil {
		return domain.Session{}, err
	}

	return *session, nil
}

func (r *sessionRepo) Delete(ctx context.Context, id uuid.UUID) error {
	conn := r.redisPool.Get()
	defer conn.Close()

	mkey := "sessions:" + id.String()
	_, err := redis.Int(conn.Do("DEL", mkey))
	if err != nil {
		return fmt.Errorf("redis error: %w", err)
	}

	return nil
}
