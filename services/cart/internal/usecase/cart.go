package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/cart/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/cart/internal/repository"
	"google.golang.org/grpc/codes"

	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/errutil"
)

//go:generate mockgen -destination=mocks/restaurant_client_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/services/cart/internal/usecase RestaurantClient
type RestaurantClient interface {
	GetDishesByIDs(ctx context.Context, dishIDs []int64) ([]domain.Dish, error)
}

//go:generate mockgen -destination=mocks/cart_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/services/cart/internal/usecase CartUseCase
type CartUseCase interface {
	// Базовые
	GetCart(ctx context.Context, userID int64) (domain.Cart, int64, error)
	LockCart(ctx context.Context, cartID string, userID int64, idempotencyKey string) error
	UnlockCart(ctx context.Context, cartID string, userID int64, idempotencyKey string) error
	ClearCart(ctx context.Context, cartID string, userID int64, idempotencyKey string) error

	// Управление комнатой
	GenerateInvite(ctx context.Context, cartID string, adminID int64) (domain.CartInvite, error)
	JoinCart(ctx context.Context, token string, userID int64) (string, error)
	KickMember(ctx context.Context, cartID string, adminID, targetUserID int64, idempotencyKey string) error
	CloseSharedCart(ctx context.Context, cartID string, adminID int64, idempotencyKey string) error

	// Товары
	AddItem(ctx context.Context, cartID string, userID, dishID int64, quantity int32, idempotencyKey string) error
	RemoveItem(ctx context.Context, cartID string, userID, dishID int64, idempotencyKey string) error
	UpdateItemQuantity(ctx context.Context, cartID string, userID, dishID int64, qty int32, idempotencyKey string) error
	ReassignItemOwner(ctx context.Context, cartID string, adminID, dishID int64, newOwnerID *int64, idempotencyKey string) error
}

type cartUseCase struct {
	cartRepo           repository.CartRepository
	restaurantClient   RestaurantClient
	defaultFoodLogoURL string
}

func NewCartUseCase(cr repository.CartRepository, rc RestaurantClient, dflurl string) *cartUseCase {
	return &cartUseCase{
		cartRepo:           cr,
		restaurantClient:   rc,
		defaultFoodLogoURL: dflurl,
	}
}

func (u *cartUseCase) GetCart(ctx context.Context, userID int64) (domain.Cart, int64, error) {
	cart, err := u.cartRepo.GetCartByUserID(ctx, userID)
	if err != nil {
		return domain.Cart{}, 0, err
	}

	if len(cart.Items) == 0 {
		return cart, 0, nil
	}

	dishIDs := make([]int64, 0, len(cart.Items))
	for _, item := range cart.Items {
		dishIDs = append(dishIDs, item.DishID)
	}

	dishes, err := u.restaurantClient.GetDishesByIDs(ctx, dishIDs)
	if err != nil {
		return domain.Cart{}, 0, errutil.Wrap("INTERNAL_SERVER_ERROR", "failed to reach restaurant service", err, codes.Internal)
	}

	dishMap := make(map[int64]domain.Dish)
	for _, d := range dishes {
		dishMap[d.ID] = d
	}

	var totalCost int64
	validItems := make([]domain.CartItem, 0, len(cart.Items))

	for i := range cart.Items {
		dishInfo, ok := dishMap[cart.Items[i].DishID]
		if !ok {
			continue // Блюдо пропало из меню
		}

		cart.Items[i].Name = dishInfo.Name
		cart.Items[i].Price = dishInfo.Price
		cart.Items[i].ImageURL = dishInfo.ImageURL
		if cart.Items[i].ImageURL == "" {
			cart.Items[i].ImageURL = u.defaultFoodLogoURL
		}

		totalCost += cart.Items[i].Price * int64(cart.Items[i].Quantity)
		validItems = append(validItems, cart.Items[i])
	}

	cart.Items = validItems
	return cart, totalCost, nil
}

func (u *cartUseCase) LockCart(ctx context.Context, cartID string, userID int64, idempotencyKey string) error {
	cart, err := u.cartRepo.GetCartByID(ctx, cartID)
	if err != nil {
		return err
	}

	if !cart.HasMember(userID) {
		return domain.ErrForbidden
	}

	// проверка на ничейные позиции
	for _, item := range cart.Items {
		if item.OwnerUserID == nil {
			return domain.ErrUnassignedItems
		}
	}

	return u.cartRepo.LockCart(ctx, cartID)
}

func (u *cartUseCase) UnlockCart(ctx context.Context, cartID string, userID int64, idempotencyKey string) error {
	cart, err := u.cartRepo.GetCartByID(ctx, cartID)
	if err != nil {
		return err
	}

	if !cart.HasMember(userID) {
		return domain.ErrForbidden
	}

	return u.cartRepo.UnlockCart(ctx, cartID)
}

func (u *cartUseCase) ClearCart(ctx context.Context, cartID string, userID int64, idempotencyKey string) error {
	cart, err := u.cartRepo.GetCartByID(ctx, cartID)
	if err != nil {
		return err
	}

	if cart.AdminID != userID {
		return domain.ErrForbidden
	}

	return u.cartRepo.ClearCart(ctx, cartID)
}

func (u *cartUseCase) GenerateInvite(ctx context.Context, cartID string, adminID int64) (domain.CartInvite, error) {
	cart, err := u.cartRepo.GetCartByID(ctx, cartID)
	if err != nil {
		return domain.CartInvite{}, err
	}

	if cart.AdminID != adminID {
		return domain.CartInvite{}, domain.ErrForbidden
	}

	// Переводим корзину в режим Shared, если она еще не там
	if cart.Mode == domain.CartModeSolo {
		if err := u.cartRepo.UpdateCartMode(ctx, cartID, domain.CartModeShared); err != nil {
			return domain.CartInvite{}, err
		}
	}

	b := make([]byte, 16)
	_, _ = rand.Read(b)
	token := hex.EncodeToString(b)

	invite := domain.CartInvite{
		Token:     token,
		CartID:    cartID,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	if err := u.cartRepo.SaveInvite(ctx, invite); err != nil {
		return domain.CartInvite{}, err
	}

	return invite, nil
}

func (u *cartUseCase) JoinCart(ctx context.Context, token string, userID int64) (string, error) {
	invite, err := u.cartRepo.GetInviteByToken(ctx, token)
	if err != nil {
		return "", err
	}

	if time.Now().After(invite.ExpiresAt) {
		return "", domain.ErrInviteExpired
	}

	if err := u.cartRepo.AddMember(ctx, invite.CartID, userID); err != nil {
		return "", err
	}

	return invite.CartID, nil
}

func (u *cartUseCase) KickMember(ctx context.Context, cartID string, adminID, targetUserID int64, idempotencyKey string) error {
	cart, err := u.cartRepo.GetCartByID(ctx, cartID)
	if err != nil {
		return err
	}

	if cart.AdminID != adminID {
		return domain.ErrForbidden
	}

	if err := u.cartRepo.RemoveMember(ctx, cartID, targetUserID); err != nil {
		return err
	}

	// обезличиваем позиции кикнутого
	return u.cartRepo.OrphanUserItems(ctx, cartID, targetUserID)
}

func (u *cartUseCase) CloseSharedCart(ctx context.Context, cartID string, adminID int64, idempotencyKey string) error {
	cart, err := u.cartRepo.GetCartByID(ctx, cartID)
	if err != nil {
		return err
	}

	if cart.AdminID != adminID {
		return domain.ErrForbidden
	}

	return u.cartRepo.DowngradeToSolo(ctx, cartID, adminID)
}

func (u *cartUseCase) AddItem(ctx context.Context, cartID string, userID, dishID int64, quantity int32, idempotencyKey string) error {
	if quantity <= 0 {
		return domain.ErrInvalidQuantity
	}

	cart, err := u.cartRepo.GetCartByID(ctx, cartID)
	if err != nil {
		return err
	}

	if !cart.HasMember(userID) {
		return domain.ErrForbidden
	}

	dishes, err := u.restaurantClient.GetDishesByIDs(ctx, []int64{dishID})
	if err != nil || len(dishes) == 0 {
		return domain.ErrDishNotFound
	}

	if dishes[0].RestaurantBrandID != cart.RestaurantBrandID {
		return domain.ErrMultipleRestaurants
	}

	item := domain.CartItem{
		DishID:      dishID,
		Quantity:    quantity,
		OwnerUserID: &userID,
	}

	return u.cartRepo.AddItem(ctx, cartID, item)
}

func (u *cartUseCase) RemoveItem(ctx context.Context, cartID string, userID, dishID int64, idempotencyKey string) error {
	cart, err := u.cartRepo.GetCartByID(ctx, cartID)
	if err != nil {
		return err
	}

	// Гость может удалять только свои, Админ может удалять любые
	item := cart.GetItem(dishID)
	if item == nil {
		return nil // идемпотентность
	}

	if cart.AdminID != userID {
		if item.OwnerUserID == nil || *item.OwnerUserID != userID {
			return domain.ErrForbidden
		}
	}

	return u.cartRepo.RemoveItem(ctx, cartID, dishID)
}

func (u *cartUseCase) UpdateItemQuantity(ctx context.Context, cartID string, userID, dishID int64, qty int32, idempotencyKey string) error {
	if qty <= 0 {
		return domain.ErrInvalidQuantity
	}

	cart, err := u.cartRepo.GetCartByID(ctx, cartID)
	if err != nil {
		return err
	}

	item := cart.GetItem(dishID)
	if item == nil {
		return domain.ErrDishNotFound
	}

	// Гость не может менять чужое или ничейное
	if cart.AdminID != userID {
		if item.OwnerUserID == nil || *item.OwnerUserID != userID {
			return domain.ErrForbidden
		}
	}

	return u.cartRepo.UpdateItemQuantity(ctx, cartID, dishID, qty)
}

func (u *cartUseCase) ReassignItemOwner(ctx context.Context, cartID string, adminID, dishID int64, newOwnerID *int64, idempotencyKey string) error {
	cart, err := u.cartRepo.GetCartByID(ctx, cartID)
	if err != nil {
		return err
	}

	if cart.AdminID != adminID {
		return domain.ErrForbidden
	}

	if newOwnerID != nil && !cart.HasMember(*newOwnerID) {
		return domain.ErrUserNotInCart
	}

	return u.cartRepo.ReassignItemOwner(ctx, cartID, dishID, newOwnerID)
}
