package usecase

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/repository"
)

//go:generate mockgen -destination=mocks/cart_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/internal/usecase CartUseCase
type CartUseCase interface {
	GetCart(ctx context.Context, userID int) (domain.Cart, int64, error) // Возвращает карту, полную стоимость коризины и ошибку
	UpdateCart(ctx context.Context, userID int, cartData domain.Cart) error
}

type cartUseCase struct {
	cartRepo repository.CartRepository
	dishRepo repository.RestaurantBrandRepository
}

func NewCartUseCase(cr repository.CartRepository, dr repository.RestaurantBrandRepository) *cartUseCase {
	return &cartUseCase{
		cartRepo: cr,
		dishRepo: dr,
	}
}

func (u *cartUseCase) GetCart(ctx context.Context, userID int) (domain.Cart, int64, error) {
	cart, err := u.cartRepo.GetCartByUserID(ctx, userID)
	if err != nil {
		return domain.Cart{}, 0, err
	}

	var totalCost int64
	for _, cartItem := range cart.Items {
		totalCost += cartItem.Price * int64(cartItem.Quantity)
	}

	return cart, totalCost, nil
}

func (u *cartUseCase) UpdateCart(ctx context.Context, userID int, cartData domain.Cart) error {
	if len(cartData.Items) == 0 {
		return u.cartRepo.ClearCart(ctx, userID)
	}

	for _, cartItem := range cartData.Items {
		if cartItem.Quantity == 0 {
			return domain.ErrInvalidQuantity
		}

		// dish, err := u.dishRepo.GetDishByID(ctx, cartItem.DishID)
		// if err != nil {
		// 	return err
		// }

		// if dish.RestaurantID != cartData.RestaurantBrandID {
		// 	return domain.ErrMultipleRestaurants
		// }
	}

	return u.cartRepo.UpdateCart(ctx, userID, cartData.RestaurantBrandID, cartData.Items)
}
