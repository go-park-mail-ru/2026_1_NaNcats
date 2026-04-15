package repository

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
)

//go:generate mockgen -destination=mocks/cart_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/internal/repository CartRepository
type CartRepository interface {
	GetCartByUserID(ctx context.Context, userID int) (domain.Cart, error)
	UpdateCart(ctx context.Context, userID int, resID int, items []domain.CartItem) error
	ClearCart(ctx context.Context, userId int) error
}
