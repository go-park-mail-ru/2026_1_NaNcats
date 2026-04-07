package repository

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
)

type CartRepository interface {
	GetCartByUserID(ctx context.Context, userID int) (domain.Cart, error)
	UpdateCart(ctx context.Context, userID int, resID int, items []domain.CartItem) error
	ClearCart(ctx context.Context, userId int) error
}
