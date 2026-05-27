package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/cart/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/cart/internal/repository"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

//go:generate mockgen -destination=mocks/restaurant_client_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/services/cart/internal/usecase RestaurantClient
type RestaurantClient interface {
	GetDishesByIDs(ctx context.Context, dishIDs []int64) ([]domain.Dish, error)
}

//go:generate mockgen -destination=mocks/cart_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/services/cart/internal/usecase CartUseCase
//go:generate gowrap gen -i CartUseCase -t ../../../../shared/templates/tracing.tmpl -o cart_tracing_mw.go -v TracerName=cart-service
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
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.Int64("user.id", userID))

	cart, err := u.cartRepo.GetCartByUserID(ctx, userID)
	if err != nil {
		return domain.Cart{}, 0, err
	}

	if len(cart.Items) == 0 {
		return cart, 0, nil
	}

	if cart.ID != "" {
		span.SetAttributes(attribute.String("cart.id", cart.ID))
	}

	dishIDs := make([]int64, 0, len(cart.Items))
	for _, item := range cart.Items {
		dishIDs = append(dishIDs, item.DishID)
	}

	dishes, err := u.restaurantClient.GetDishesByIDs(ctx, dishIDs)
	if err != nil {
		span.AddEvent("dishes_enrichment_failed", trace.WithAttributes(
			attribute.String("error", err.Error()),
		))
		return cart, 0, nil
	}

	dishMap := make(map[int64]domain.Dish)
	for _, d := range dishes {
		dishMap[d.ID] = d
	}

	validItems := make([]domain.CartItem, 0, len(cart.Items))
	var totalCost int64

	for i := range cart.Items {
		dishInfo, ok := dishMap[cart.Items[i].DishID]
		if !ok {
			span.AddEvent("dish_not_found_in_menu", trace.WithAttributes(
				attribute.Int64("dish.id", cart.Items[i].DishID),
			))
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

	span.SetAttributes(
		attribute.Int("cart.items_count", len(validItems)),
		attribute.Int64("cart.total_cost", totalCost),
	)

	return cart, totalCost, nil
}

func (u *cartUseCase) LockCart(ctx context.Context, cartID string, userID int64, idempotencyKey string) error {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.String("cart.id", cartID),
		attribute.Int64("user.id", userID),
		attribute.String("lock.idempotency_key", idempotencyKey),
	)

	err := u.cartRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := u.cartRepo.CheckAndSaveIdempotency(txCtx, userID, idempotencyKey, "LockCart"); err != nil {
			return err
		}

		var activeCartID string
		if cartID == "" {
			activeCart, err := u.cartRepo.GetActiveCartByUserID(txCtx, userID)
			if err != nil {
				return err
			}
			activeCartID = activeCart.ID
		} else {
			activeCartID = cartID
		}

		cart, err := u.cartRepo.GetCartByID(txCtx, activeCartID)
		if err != nil {
			return err
		}

		span.SetAttributes(attribute.Int64("cart.admin_id", cart.AdminID))

		if cart.AdminID != userID {
			return domain.ErrForbidden
		}

		for _, item := range cart.Items {
			if item.OwnerUserID == nil {
				span.AddEvent("unassigned_item_error", trace.WithAttributes(
					attribute.Int64("dish.id", item.DishID),
				))
				return domain.ErrUnassignedItems
			}
		}

		return u.cartRepo.LockCart(txCtx, activeCartID)
	})

	if errors.Is(err, domain.ErrIdempotencyConflict) {
		span.AddEvent("idempotency_hit_skipping")
		return nil
	}
	return err
}

func (u *cartUseCase) UnlockCart(ctx context.Context, cartID string, userID int64, idempotencyKey string) error {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.String("cart.id", cartID),
		attribute.Int64("user.id", userID),
		attribute.String("unlock.idempotency_key", idempotencyKey),
	)

	err := u.cartRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := u.cartRepo.CheckAndSaveIdempotency(txCtx, userID, idempotencyKey, "UnlockCart"); err != nil {
			return err
		}

		cart, err := u.cartRepo.GetCartByID(txCtx, cartID)
		if err != nil {
			return err
		}

		if !cart.HasMember(userID) {
			return domain.ErrForbidden
		}

		return u.cartRepo.UnlockCart(txCtx, cartID)
	})

	if errors.Is(err, domain.ErrIdempotencyConflict) {
		span.AddEvent("idempotency_hit_skipping")
		return nil
	}
	return err
}

func (u *cartUseCase) ClearCart(ctx context.Context, cartID string, userID int64, idempotencyKey string) error {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.String("cart.id", cartID),
		attribute.Int64("user.id", userID),
		attribute.String("clear.idempotency_key", idempotencyKey),
	)

	err := u.cartRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := u.cartRepo.CheckAndSaveIdempotency(txCtx, userID, idempotencyKey, "ClearCart"); err != nil {
			return err
		}

		cart, err := u.cartRepo.GetCartByID(txCtx, cartID)
		if err != nil {
			return err
		}

		if cart.AdminID != userID {
			return domain.ErrForbidden
		}

		return u.cartRepo.ClearCart(txCtx, cartID)
	})

	if errors.Is(err, domain.ErrIdempotencyConflict) {
		span.AddEvent("idempotency_hit_skipping")
		return nil
	}
	return err
}

func (u *cartUseCase) GenerateInvite(ctx context.Context, cartID string, adminID int64) (domain.CartInvite, error) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.String("cart.id", cartID),
		attribute.Int64("admin.id", adminID),
	)

	cart, err := u.cartRepo.GetCartByID(ctx, cartID)
	if err != nil {
		return domain.CartInvite{}, err
	}

	if cart.AdminID != adminID {
		err = domain.ErrForbidden
		return domain.CartInvite{}, err
	}

	if cart.Mode == domain.CartModeSolo {
		span.AddEvent("switching_to_shared_mode")
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

	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.String("cart.id", invite.CartID))

	if time.Now().After(invite.ExpiresAt) {
		return "", domain.ErrInviteExpired
	}

	if err := u.cartRepo.AddMember(ctx, invite.CartID, userID); err != nil {
		return "", err
	}

	return invite.CartID, nil
}

func (u *cartUseCase) KickMember(ctx context.Context, cartID string, adminID, targetUserID int64, idempotencyKey string) error {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.String("cart.id", cartID),
		attribute.Int64("admin.id", adminID),
		attribute.Int64("target_user.id", targetUserID),
		attribute.String("kick.idempotency_key", idempotencyKey),
	)

	err := u.cartRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := u.cartRepo.CheckAndSaveIdempotency(txCtx, adminID, idempotencyKey, "KickMember"); err != nil {
			return err
		}

		cart, err := u.cartRepo.GetCartByID(txCtx, cartID)
		if err != nil {
			return err
		}

		if cart.AdminID != adminID {
			return domain.ErrForbidden
		}

		if err := u.cartRepo.RemoveMember(txCtx, cartID, targetUserID); err != nil {
			return err
		}
		return u.cartRepo.OrphanUserItems(txCtx, cartID, targetUserID)
	})

	if errors.Is(err, domain.ErrIdempotencyConflict) {
		span.AddEvent("idempotency_hit_skipping")
		return nil
	}
	return err
}

func (u *cartUseCase) CloseSharedCart(ctx context.Context, cartID string, adminID int64, idempotencyKey string) error {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.String("cart.id", cartID),
		attribute.Int64("admin.id", adminID),
		attribute.String("close.idempotency_key", idempotencyKey),
	)

	err := u.cartRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := u.cartRepo.CheckAndSaveIdempotency(txCtx, adminID, idempotencyKey, "CloseSharedCart"); err != nil {
			return err
		}

		cart, err := u.cartRepo.GetCartByID(txCtx, cartID)
		if err != nil {
			return err
		}

		if cart.AdminID != adminID {
			return domain.ErrForbidden
		}

		return u.cartRepo.DowngradeToSolo(txCtx, cartID, adminID)
	})

	if errors.Is(err, domain.ErrIdempotencyConflict) {
		span.AddEvent("idempotency_hit_skipping")
		return nil
	}
	return err
}

func (u *cartUseCase) AddItem(ctx context.Context, cartID string, userID, dishID int64, quantity int32, idempotencyKey string) error {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.Int64("user.id", userID),
		attribute.Int64("dish.id", dishID),
		attribute.Int("quantity", int(quantity)),
	)

	if quantity <= 0 {
		return domain.ErrInvalidQuantity
	}

	dishes, err := u.restaurantClient.GetDishesByIDs(ctx, []int64{dishID})
	if err != nil || len(dishes) == 0 {
		return domain.ErrDishNotFound
	}
	dishBrandID := dishes[0].RestaurantBrandID
	span.SetAttributes(attribute.Int64("restaurant.brand_id", dishBrandID))

	err = u.cartRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := u.cartRepo.CheckAndSaveIdempotency(txCtx, userID, idempotencyKey, "AddItem"); err != nil {
			return err
		}

		var cart domain.Cart
		var err error

		if cartID != "" {
			span.SetAttributes(attribute.String("cart.id", cartID))
			cart, err = u.cartRepo.GetCartByID(txCtx, cartID)
			if err != nil {
				return err
			}
			if !cart.HasMember(userID) {
				return domain.ErrForbidden
			}
			if cart.Status != domain.CartStatusActive {
				return domain.ErrCartLocked
			}
		} else {
			cart, err = u.cartRepo.GetActiveCartByUserID(txCtx, userID)
			if err != nil {
				span.AddEvent("creating_new_cart")
				newCartID, createErr := u.cartRepo.CreateCart(txCtx, userID, dishBrandID)
				if createErr != nil {
					return createErr
				}
				cartID = newCartID

				cart = domain.Cart{
					ID:                cartID,
					RestaurantBrandID: dishBrandID,
				}
			} else {
				cartID = cart.ID
			}
			span.SetAttributes(attribute.String("cart.id", cartID))
		}

		if cart.RestaurantBrandID != dishBrandID {
			if len(cart.Items) == 0 {
				if err := u.cartRepo.SetCartRestaurantBrand(txCtx, cartID, dishBrandID); err != nil {
					return err
				}
				cart.RestaurantBrandID = dishBrandID
			} else {
				return domain.ErrMultipleRestaurants
			}
		}

		item := domain.CartItem{
			DishID:      dishID,
			Quantity:    quantity,
			OwnerUserID: &userID,
		}

		return u.cartRepo.AddItem(txCtx, cartID, item)
	})

	if errors.Is(err, domain.ErrIdempotencyConflict) {
		span.AddEvent("idempotency_hit_skipping")
		return nil
	}
	return err
}

func (u *cartUseCase) RemoveItem(ctx context.Context, cartID string, userID, dishID int64, idempotencyKey string) error {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.String("cart.id", cartID),
		attribute.Int64("user.id", userID),
		attribute.Int64("dish.id", dishID),
	)

	err := u.cartRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := u.cartRepo.CheckAndSaveIdempotency(txCtx, userID, idempotencyKey, "RemoveItem"); err != nil {
			return err
		}

		cart, err := u.cartRepo.GetCartByID(txCtx, cartID)
		if err != nil {
			return err
		}

		item := cart.GetItem(dishID, userID)
		if item == nil {
			span.AddEvent("item_already_absent")
			return nil
		}

		return u.cartRepo.RemoveItem(txCtx, cartID, dishID, userID)
	})

	if errors.Is(err, domain.ErrIdempotencyConflict) {
		span.AddEvent("idempotency_hit_skipping")
		return nil
	}
	return err
}

func (u *cartUseCase) UpdateItemQuantity(ctx context.Context, cartID string, userID, dishID int64, qty int32, idempotencyKey string) error {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.String("cart.id", cartID),
		attribute.Int64("user.id", userID),
		attribute.Int64("dish.id", dishID),
		attribute.Int64("quantity.new", int64(qty)),
	)

	if qty <= 0 {
		return domain.ErrInvalidQuantity
	}

	cart, err := u.cartRepo.GetCartByID(ctx, cartID)
	if err != nil {
		return err
	}

	// Меняем только свою позицию блюда: у каждого участника она своя.
	item := cart.GetItem(dishID, userID)
	if item == nil {
		span.AddEvent("item_not_found_in_cart")
		return domain.ErrDishNotFound
	}

	return u.cartRepo.UpdateItemQuantity(ctx, cartID, dishID, userID, qty)
}

func (u *cartUseCase) ReassignItemOwner(ctx context.Context, cartID string, adminID, dishID int64, newOwnerID *int64, idempotencyKey string) error {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.String("cart.id", cartID),
		attribute.Int64("admin.id", adminID),
		attribute.Int64("dish.id", dishID),
	)

	err := u.cartRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := u.cartRepo.CheckAndSaveIdempotency(txCtx, adminID, idempotencyKey, "ReassignItemOwner"); err != nil {
			return err
		}

		if newOwnerID != nil {
			span.SetAttributes(attribute.Int64("new_owner.id", *newOwnerID))
		} else {
			span.AddEvent("stripping_item_owner")
		}

		cart, err := u.cartRepo.GetCartByID(txCtx, cartID)
		if err != nil {
			return err
		}

		if cart.AdminID != adminID {
			return domain.ErrForbidden
		}

		if newOwnerID != nil && !cart.HasMember(*newOwnerID) {
			span.AddEvent("new_owner_not_in_cart")
			return domain.ErrUserNotInCart
		}

		return u.cartRepo.ReassignItemOwner(txCtx, cartID, dishID, newOwnerID)
	})

	if errors.Is(err, domain.ErrIdempotencyConflict) {
		span.AddEvent("idempotency_hit_skipping")
		return nil
	}
	return err
}
