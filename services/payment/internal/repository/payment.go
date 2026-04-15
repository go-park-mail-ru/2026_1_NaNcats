package repository

import (
	"context"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
)

type PaymentRepository interface {
	Create(ctx context.Context, method domain.PaymentMethod) (int, error)
	Delete(ctx context.Context, cardID string, userID int) error
	GetByUserID(ctx context.Context, userID int) ([]domain.PaymentMethod, error)
	SetDefault(ctx context.Context, cardID string, userID int) error
}

type PaymentCacheRepository interface {
	SetPendingBinding(ctx context.Context, paymentID string, userID int, ttl time.Duration) error
	DeletePendingBinding(ctx context.Context, paymentID string) error
	GetUserIDByPaymentID(ctx context.Context, paymentID string) (int, error)
}
