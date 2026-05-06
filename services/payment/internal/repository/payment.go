package repository

import (
	"context"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/payment/internal/domain"
)

//go:generate mockgen -destination=mocks/payment_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/services/payment/internal/repository PaymentRepository,PaymentCacheRepository

type PaymentRepository interface {
	Create(ctx context.Context, method domain.PaymentMethod, idempotencyKey string) (int64, error)
	Delete(ctx context.Context, cardID string, userID int64) error
	GetByUserID(ctx context.Context, userID int64) ([]domain.PaymentMethod, error)
	SetDefault(ctx context.Context, cardID string, userID int64) error
}

type PaymentCacheRepository interface {
	SetPendingBinding(ctx context.Context, paymentID string, userID int64, ttl time.Duration) error
	DeletePendingBinding(ctx context.Context, paymentID string) error
	GetUserIDByPaymentID(ctx context.Context, paymentID string) (int64, error)
}
