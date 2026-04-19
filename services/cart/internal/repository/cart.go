package repository

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/cart/internal/domain"
)

//go:generate mockgen -destination=mocks/cart_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/internal/repository CartRepository
type CartRepository interface {
	GetCartByUserID(ctx context.Context, userID int64) (domain.Cart, error)
	UpdateCart(ctx context.Context, userID int64, resID int64, items []domain.CartItem, idempotencyKey string) error
	ClearCart(ctx context.Context, userID int64, idempotencyKey string) error
	LockCart(ctx context.Context, userID int64, idempotencyKey string) error
	UnlockCart(ctx context.Context, userID int64, idempotencyKey string) error
}
