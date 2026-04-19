package usecase

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/cart/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/cart/internal/repository"
	"google.golang.org/grpc/codes"

	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/errutil"
)

//go:generate mockgen -destination=mocks/restaurant_client_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/services/cart/internal/usecase RestaurantClient
type RestaurantClient interface {
	GetDishesByIDs(ctx context.Context, dishIDs []int64) ([]domain.Dish, error)
}

//go:generate mockgen -destination=mocks/cart_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/internal/usecase/cart CartUseCase
type CartUseCase interface {
	GetCart(ctx context.Context, userID int64) (domain.Cart, int64, error) // Возвращает карту, полную стоимость коризины и ошибку
	UpdateCart(ctx context.Context, userID int64, cartData domain.Cart, idempotencyKey string) error
	LockCart(ctx context.Context, userID int64, idempotencyKey string) error
	UnlockCart(ctx context.Context, userID int64, idempotencyKey string) error
	ClearCart(ctx context.Context, userID int64, idempotencyKey string) error
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
		return domain.Cart{}, 0, errutil.Wrap("failed to reach restaurant service", err, codes.Internal)
	}

	// Мапка для быстрого поиска данных о блюде
	dishMap := make(map[int64]domain.Dish)
	for _, d := range dishes {
		dishMap[d.ID] = d
	}

	var totalCost int64
	validItems := make([]domain.CartItem, 0, len(cart.Items))

	for i := range cart.Items {
		dishInfo, ok := dishMap[cart.Items[i].DishID]
		if !ok {
			// Блюдо было в корзине, но исчезло из меню ресторана
			continue
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

func (u *cartUseCase) UpdateCart(ctx context.Context, userID int64, cartData domain.Cart, idempotencyKey string) error {
	if len(cartData.Items) == 0 {
		return u.cartRepo.ClearCart(ctx, userID)
	}

	dishIDs := make([]int64, len(cartData.Items))
	for ind, item := range cartData.Items {
		if item.Quantity <= 0 {
			return domain.ErrInvalidQuantity
		}
		dishIDs[ind] = item.DishID
	}

	dishes, err := u.restaurantClient.GetDishesByIDs(ctx, dishIDs)
	if err != nil {
		return errutil.Wrap("validation: failed to reach restaurant service", err, codes.Internal)
	}

	if len(dishes) != len(dishIDs) {
		return domain.ErrDishNotFound
	}

	for _, d := range dishes {
		if d.RestaurantBrandID != cartData.RestaurantBrandID {
			return errutil.New("all items must be from the same restaurant", codes.InvalidArgument)
		}
	}

	return u.cartRepo.UpdateCart(ctx, userID, cartData.RestaurantBrandID, cartData.Items)
}

func (u *cartUseCase) LockCart(ctx context.Context, userID int64, idempotencyKey string) error {
	return u.cartRepo.LockCart(ctx, userID)
}

func (u *cartUseCase) UnlockCart(ctx context.Context, userID int64, idempotencyKey string) error {
	return u.cartRepo.UnlockCart(ctx, userID)
}

func (u *cartUseCase) ClearCart(ctx context.Context, userID int64, idempotencyKey string) error {
	return u.cartRepo.ClearCart(ctx, userID)
}
