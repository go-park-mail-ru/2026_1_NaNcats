package repository

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/cart/internal/domain"
)

//go:generate mockgen -destination=mocks/cart_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/services/cart/internal/repository CartRepository
type CartRepository interface {
	// Базовые операции с корзиной
	GetCartByUserID(ctx context.Context, userID int64) (domain.Cart, error)
	GetCartByID(ctx context.Context, cartID string) (domain.Cart, error)
	LockCart(ctx context.Context, cartID string) error
	UnlockCart(ctx context.Context, cartID string) error
	ClearCart(ctx context.Context, cartID string) error
	UpdateCartMode(ctx context.Context, cartID string, mode string) error
	SetCartRestaurantBrand(ctx context.Context, cartID string, brandID int64) error
	GetActiveCartByUserID(ctx context.Context, userID int64) (domain.Cart, error)
	CreateCart(ctx context.Context, adminID int64, brandID int64) (string, error)
	// Гранулярные операции с позициями
	AddItem(ctx context.Context, cartID string, item domain.CartItem) error
	RemoveItem(ctx context.Context, cartID string, dishID int64) error
	UpdateItemQuantity(ctx context.Context, cartID string, dishID int64, quantity int32) error
	ReassignItemOwner(ctx context.Context, cartID string, dishID int64, newOwnerID *int64) error
	OrphanUserItems(ctx context.Context, cartID string, targetUserID int64) error // Делает позиции кикнутого юзера ничейными
	// Управление Shared Cart
	SaveInvite(ctx context.Context, invite domain.CartInvite) error
	GetInviteByToken(ctx context.Context, token string) (domain.CartInvite, error)
	AddMember(ctx context.Context, cartID string, userID int64) error
	RemoveMember(ctx context.Context, cartID string, userID int64) error
	DowngradeToSolo(ctx context.Context, cartID string, adminID int64) error
}
