package cart

import (
	"context"
	"html"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/repository"
)

//go:generate mockgen -destination=mocks/cart_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/internal/usecase/cart CartUseCase
type CartUseCase interface {
	GetCart(ctx context.Context, userID int) (domain.Cart, int64, error) // Возвращает карту, полную стоимость коризины и ошибку
	UpdateCart(ctx context.Context, userID int, cartData domain.Cart) error
	LockCart(ctx context.Context, userID int64) error
	UnlockCart(ctx context.Context, userID int64) error
	ClearCart(ctx context.Context, userID int64) error
}

type cartUseCase struct {
	cartRepo repository.CartRepository
	// TODO: убрать dishRepo
	dishRepo           repository.DishRepository
	defaultFoodLogoURL string
}

func NewCartUseCase(cr repository.CartRepository, dr repository.DishRepository, dflurl string) *cartUseCase {
	return &cartUseCase{
		cartRepo:           cr,
		dishRepo:           dr,
		defaultFoodLogoURL: dflurl,
	}
}

func (u *cartUseCase) GetCart(ctx context.Context, userID int) (domain.Cart, int64, error) {
	cart, err := u.cartRepo.GetCartByUserID(ctx, userID)
	if err != nil {
		return domain.Cart{}, 0, err
	}

	var totalCost int64
	for i, cartItem := range cart.Items {
		if cartItem.ImageURL == "" {
			cart.Items[i].ImageURL = u.defaultFoodLogoURL
		}

		totalCost += cartItem.Price * int64(cartItem.Quantity)
	}

	return cart, totalCost, nil
}

func (u *cartUseCase) UpdateCart(ctx context.Context, userID int, cartData domain.Cart) error {
	if len(cartData.Items) == 0 {
		return u.cartRepo.ClearCart(ctx, userID)
	}

	dishIDs := make([]int, len(cartData.Items))
	for ind, item := range cartData.Items {
		if item.Quantity <= 0 {
			return domain.ErrInvalidQuantity
		}
		cartData.Items[ind].Name = html.EscapeString(cartData.Items[ind].Name)
		dishIDs[ind] = item.DishID
	}

	dishes, err := u.dishRepo.GetDishesByIDs(ctx, dishIDs)
	if err != nil {
		return err
	}

	if len(dishes) != len(dishIDs) {
		return domain.ErrDishNotFound
	}

	for _, dish := range dishes {
		if dish.RestaurantBrandID != cartData.RestaurantBrandID {
			return domain.ErrMultipleRestaurants
		}
	}

	return u.cartRepo.UpdateCart(ctx, userID, cartData.RestaurantBrandID, cartData.Items)
}

func (u *cartUseCase) LockCart(ctx context.Context, userID int64) error {
	return u.cartRepo.LockCart(ctx, userID)
}

func (u *cartUseCase) UnlockCart(ctx context.Context, userID int64) error {
	return u.cartRepo.UnlockCart(ctx, userID)
}

func (u *cartUseCase) ClearCart(ctx context.Context, userID int64) error {
	return u.cartRepo.ClearCart(ctx, userID)
}
