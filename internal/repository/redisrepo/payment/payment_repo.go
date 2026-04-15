package redisrepo

import (
	"context"
	"fmt"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/repository"
	"github.com/gomodule/redigo/redis"
)

type paymentCacheRepo struct {
	redisPool *redis.Pool
}

func NewPaymentCacheRepo(pool *redis.Pool) repository.PaymentCacheRepository {
	return &paymentCacheRepo{
		redisPool: pool,
	}
}

func (r *paymentCacheRepo) SetPendingBinding(ctx context.Context, paymentID string, userID int, ttl time.Duration) error {
	conn := r.redisPool.Get()
	defer conn.Close()

	mkey := "payment:" + paymentID
	result, err := redis.String(conn.Do("SET", mkey, userID, "EX", int(ttl.Seconds())))
	if err != nil {
		return err
	}
	if result != "OK" {
		return domain.ErrRedisResultIsNotOK
	}

	return nil
}

func (r *paymentCacheRepo) DeletePendingBinding(ctx context.Context, paymentID string) error {
	conn := r.redisPool.Get()
	defer conn.Close()

	mkey := "payment:" + paymentID
	_, err := redis.Int(conn.Do("DEL", mkey))
	if err != nil {
		return fmt.Errorf("redis error: %w", err)
	}

	return nil
}

func (r *paymentCacheRepo) GetUserIDByPaymentID(ctx context.Context, paymentID string) (int, error) {
	conn := r.redisPool.Get()
	defer conn.Close()

	mkey := "payment:" + paymentID
	userID, err := redis.Int(conn.Do("GET", mkey))
	if err != nil {
		return 0, err
	}

	return userID, nil
}
