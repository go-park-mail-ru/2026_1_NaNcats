package repository

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/order/internal/domain"
)

//go:generate mockgen -destination=mocks/promo_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/services/order/internal/repository PromoRepository

// PromoRepository - доступ к промокодам в БД OrderService
type PromoRepository interface {
	// GetUserPromocodes - промокоды, привязанные к пользователю и пока доступные ему
	GetUserPromocodes(ctx context.Context, userID int64) ([]domain.Promocode, error)
	// GetRestaurantPromocodes - действующие промокоды конкретного бренда.
	GetRestaurantPromocodes(ctx context.Context, brandID int64) ([]domain.Promocode, error)
	// GetPromocodeByCode - промокод по его коду.
	GetPromocodeByCode(ctx context.Context, code string) (domain.Promocode, error)
	// BindPromocodeToUser - привязывает промокод к пользователю. Возвращает false,
	// если такая привязка уже существовала.
	BindPromocodeToUser(ctx context.Context, userID, promoID int64) (bool, error)
	// CountUserPromocodeUsage - сколько раз пользователь уже воспользовался промокодом.
	CountUserPromocodeUsage(ctx context.Context, promoID, userID int64) (int, error)
	// RecordPromocodeUsage - фиксирует использование промокода.
	RecordPromocodeUsage(ctx context.Context, promoID, userID int64, orderID *int64) error
	// ResolveOrderInternalID - внутренний id заказа по его публичному UUID.
	ResolveOrderInternalID(ctx context.Context, orderPublicID string) (*int64, error)
	// Создание уникального промокода
	CreatePromocode(ctx context.Context, promo domain.Promocode) (domain.Promocode, error)
}

// ErrPromocodeNotFound возвращается, когда промокод с заданным кодом не найден
var ErrPromocodeNotFound = errPromocodeNotFound{}

type errPromocodeNotFound struct{}

func (errPromocodeNotFound) Error() string { return "promocode not found" }
