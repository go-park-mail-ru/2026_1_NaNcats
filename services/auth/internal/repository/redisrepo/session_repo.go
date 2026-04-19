package redisrepo

import (
	"context"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/auth/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/auth/internal/repository"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/errutil"
	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
	"github.com/mailru/easyjson"
	"google.golang.org/grpc/codes"
)

const CSRFTokenTTL = 3600

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
		return errutil.New("failed to set session: redis result is not OK", codes.Internal)
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
		if err == redis.ErrNil {
			return domain.Session{}, domain.ErrSessionNotFound
		}
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
		return err
	}

	return nil
}

func (r *sessionRepo) SetCSRF(ctx context.Context, id uuid.UUID, token string) error {
	conn := r.redisPool.Get()
	defer conn.Close()

	mkey := "csrf:" + id.String()
	result, err := redis.String(conn.Do("SET", mkey, token, "EX", CSRFTokenTTL))
	if err != nil {
		return err
	}
	if result != "OK" {
		return errutil.New("failed to set csrf: redis result is not OK", codes.Internal)
	}

	return nil
}

func (r *sessionRepo) GetCSRF(ctx context.Context, id uuid.UUID) (string, error) {
	conn := r.redisPool.Get()
	defer conn.Close()

	mkey := "csrf:" + id.String()
	token, err := redis.String(conn.Do("GET", mkey))
	if err != nil {
		if err == redis.ErrNil {
			// ключа просто нет, это нормальное поведение
			return "", nil
		}
		return "", err
	}

	return token, nil
}
