package redisrepo

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/repository"
	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
)

type sessionRepo struct {
	redisConn redis.Conn
}

func NewSessionRepo(conn redis.Conn) repository.SessionRepository {
	return &sessionRepo{
		redisConn: conn,
	}
}

func (r *sessionRepo) Create(ctx context.Context, session domain.Session, ttl time.Duration) error {
	dataSerializer, err := json.Marshal(session)
	if err != nil {
		return err
	}

	mkey := "sessions:" + session.ID.String()
	result, err := redis.String(r.redisConn.Do("SET", mkey, dataSerializer, "EX", int(ttl.Seconds())))
	if err != nil {
		return err
	}
	if result != "OK" {
		return fmt.Errorf("result is not OK")
	}

	return nil
}

func (r *sessionRepo) GetByID(ctx context.Context, id uuid.UUID) (domain.Session, error) {
	mkey := "sessions:" + id.String()
	data, err := redis.Bytes(r.redisConn.Do("GET", mkey))
	if err != nil {
		return domain.Session{}, err
	}

	session := &domain.Session{}
	err = json.Unmarshal(data, session)
	if err != nil {
		return domain.Session{}, err
	}

	return *session, nil
}

func (r *sessionRepo) Delete(ctx context.Context, id uuid.UUID) error {
	mkey := "sessions:" + id.String()
	_, err := redis.Int(r.redisConn.Do("DEL", mkey))
	if err != nil {
		return fmt.Errorf("redis error: %w", err)
	}

	return nil
}
