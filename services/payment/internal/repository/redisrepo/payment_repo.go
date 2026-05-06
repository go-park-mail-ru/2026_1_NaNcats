package redisrepo

import (
	"context"
	"fmt"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/payment/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/payment/internal/repository"
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

func (r *paymentCacheRepo) SetPendingBinding(ctx context.Context, paymentID string, userID int64, ttl time.Duration) error {
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

func (r *paymentCacheRepo) GetUserIDByPaymentID(ctx context.Context, paymentID string) (int64, error) {
	conn := r.redisPool.Get()
	defer conn.Close()

	mkey := "payment:" + paymentID
	userID, err := redis.Int64(conn.Do("GET", mkey))
	if err != nil {
		if err == redis.ErrNil {
			return 0, domain.ErrUserIDNotFoundInCache
		}
		return 0, err
	}

	return userID, nil
}
