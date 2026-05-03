package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
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
		// Не валим всю ручку — возвращаем корзину без обогащённых данных.
		// Иначе при недоступности restaurant-сервиса (или временной ошибке)
		// фронт не может ни прочитать корзину, ни добавить новый товар.
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
	if cartID == "" {
		activeCart, err := u.cartRepo.GetActiveCartByUserID(ctx, userID)
		if err != nil {
			return err
		}
		cartID = activeCart.ID
	}

	cart, err := u.cartRepo.GetCartByID(ctx, cartID)
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

	// Сначала лочим (саге нужен ивент CartLocked для перехода к payment),
	// потом сразу очищаем — корзина больше не нужна, все данные заказа
	// уже в order_db. Это позволяет пользователю оформлять следующий заказ
	// не дожидаясь завершения оплаты предыдущего и не висеть с залоченной
	// корзиной если saga умрёт где-то в payment.
	if err := u.cartRepo.LockCart(ctx, cartID); err != nil {
		return err
	}
	if err := u.cartRepo.ClearCart(ctx, cartID); err != nil {
		// Не валим всю сагу — ивент CartLocked уже отправлен. Очистка
		// неудалась — корзина останется пустой+locked, фронт сам разлочит
		// через CART_LOCKED auto-recover при следующей попытке добавить.
		span.AddEvent("post_lock_clear_failed", trace.WithAttributes(
			attribute.String("error", err.Error()),
		))
	}
	return nil
}

func (u *cartUseCase) UnlockCart(ctx context.Context, cartID string, userID int64, idempotencyKey string) error {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.String("cart.id", cartID),
		attribute.Int64("user.id", userID),
		attribute.String("unlock.idempotency_key", idempotencyKey),
	)

	cart, err := u.cartRepo.GetCartByID(ctx, cartID)
	if err != nil {
		return err
	}

	if !cart.HasMember(userID) {
		err = domain.ErrForbidden
		return err
	}

	return u.cartRepo.UnlockCart(ctx, cartID)
}

func (u *cartUseCase) ClearCart(ctx context.Context, cartID string, userID int64, idempotencyKey string) error {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.String("cart.id", cartID),
		attribute.Int64("user.id", userID),
		attribute.String("clear.idempotency_key", idempotencyKey),
	)

	cart, err := u.cartRepo.GetCartByID(ctx, cartID)
	if err != nil {
		return err
	}

	if cart.AdminID != userID {
		err = domain.ErrForbidden
		return err
	}

	return u.cartRepo.ClearCart(ctx, cartID)
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

	// Переводим корзину в режим Shared, если она еще не там
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
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.String("cart.id", cartID),
		attribute.Int64("admin.id", adminID),
		attribute.String("close.idempotency_key", idempotencyKey),
	)

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

	var cart domain.Cart

	if cartID != "" {
		span.SetAttributes(attribute.String("cart.id", cartID))
		cart, err = u.cartRepo.GetCartByID(ctx, cartID)
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
		cart, err = u.cartRepo.GetActiveCartByUserID(ctx, userID)
		if err != nil {
			// Если корзины нет создаем новую
			span.AddEvent("creating_new_cart")
			newCartID, createErr := u.cartRepo.CreateCart(ctx, userID, dishBrandID)
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
		// Пустая корзина может быть привязана к ресторану от прошлой попытки
		// (после ClearCart cart_dish=∅ но restaurant_brand_id остаётся).
		// Перепривязываем к ресторану нового блюда — это корректно, так как
		// корзина буквально пустая и ничего не теряем.
		if len(cart.Items) == 0 {
			if err := u.cartRepo.SetCartRestaurantBrand(ctx, cartID, dishBrandID); err != nil {
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

	return u.cartRepo.AddItem(ctx, cartID, item)
}

func (u *cartUseCase) RemoveItem(ctx context.Context, cartID string, userID, dishID int64, idempotencyKey string) error {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.String("cart.id", cartID),
		attribute.Int64("user.id", userID),
		attribute.Int64("dish.id", dishID),
	)

	cart, err := u.cartRepo.GetCartByID(ctx, cartID)
	if err != nil {
		return err
	}

	item := cart.GetItem(dishID)
	if item == nil {
		span.AddEvent("item_already_absent") // Событие для идемпотентного выхода
		return nil
	}

	if cart.AdminID != userID {
		if item.OwnerUserID == nil || *item.OwnerUserID != userID {
			return domain.ErrForbidden
		}
	}

	return u.cartRepo.RemoveItem(ctx, cartID, dishID)
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

	item := cart.GetItem(dishID)
	if item == nil {
		span.AddEvent("item_not_found_in_cart")
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
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.String("cart.id", cartID),
		attribute.Int64("admin.id", adminID),
		attribute.Int64("dish.id", dishID),
	)

	if newOwnerID != nil {
		span.SetAttributes(attribute.Int64("new_owner.id", *newOwnerID))
	} else {
		span.AddEvent("stripping_item_owner")
	}

	cart, err := u.cartRepo.GetCartByID(ctx, cartID)
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

	return u.cartRepo.ReassignItemOwner(ctx, cartID, dishID, newOwnerID)
}
