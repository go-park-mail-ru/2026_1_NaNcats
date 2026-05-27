package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/order/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/order/internal/repository"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/errutil"
	"google.golang.org/grpc/codes"
)

//go:generate mockgen -destination=mocks/promo_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/services/order/internal/usecase PromoUseCase

type PromoUseCase interface {
	GetUserPromos(ctx context.Context, userID int64) ([]domain.Promocode, error)
	GetRestaurantPromos(ctx context.Context, brandID int64) ([]domain.Promocode, error)
	BindPromo(ctx context.Context, userID int64, code string) (domain.Promocode, error)
	ValidatePromo(ctx context.Context, userID int64, code string, brandID, orderAmount, deliveryCost, serviceFee int64) (domain.PromoValidation, error)
	UsePromo(ctx context.Context, userID int64, code, orderPublicID string) error
	CreateAndBindWheelPromo(ctx context.Context, userID int64, title string, discountAmount *int64, discountPercent *int, brandID *int64, minOrderAmount *int64) (domain.Promocode, error)
}

type promoUseCase struct {
	repo repository.PromoRepository
}

func NewPromoUseCase(repo repository.PromoRepository) PromoUseCase {
	return &promoUseCase{repo: repo}
}

func (u *promoUseCase) GetUserPromos(ctx context.Context, userID int64) ([]domain.Promocode, error) {
	promos, err := u.repo.GetUserPromocodes(ctx, userID)
	if err != nil {
		return nil, errutil.Internal("failed to get user promocodes", err)
	}
	return promos, nil
}

func (u *promoUseCase) GetRestaurantPromos(ctx context.Context, brandID int64) ([]domain.Promocode, error) {
	promos, err := u.repo.GetRestaurantPromocodes(ctx, brandID)
	if err != nil {
		return nil, errutil.Internal("failed to get restaurant promocodes", err)
	}
	return promos, nil
}

func (u *promoUseCase) BindPromo(ctx context.Context, userID int64, code string) (domain.Promocode, error) {
	promo, err := u.repo.GetPromocodeByCode(ctx, code)
	if err != nil {
		if errors.Is(err, repository.ErrPromocodeNotFound) {
			return domain.Promocode{}, errutil.New("PROMO_NOT_FOUND", "promocode not found", codes.NotFound)
		}
		return domain.Promocode{}, errutil.Internal("failed to fetch promocode", err)
	}

	// Защита от привязки чужого персонального промокода
	if promo.UserID != nil && *promo.UserID != userID {
		return domain.Promocode{}, errutil.New("PROMO_FORBIDDEN", "promocode is tied to another user", codes.PermissionDenied)
	}

	if time.Now().After(promo.ExpiresAt) {
		return domain.Promocode{}, errutil.New("PROMO_EXPIRED", "promocode has expired", codes.FailedPrecondition)
	}

	bound, err := u.repo.BindPromocodeToUser(ctx, userID, promo.ID)
	if err != nil {
		return domain.Promocode{}, errutil.Internal("failed to bind promocode", err)
	}
	if !bound {
		return domain.Promocode{}, errutil.New("PROMO_ALREADY_BOUND", "promocode is already bound", codes.AlreadyExists)
	}

	return promo, nil
}

func (u *promoUseCase) ValidatePromo(ctx context.Context, userID int64, code string, brandID, orderAmount, deliveryCost, serviceFee int64) (domain.PromoValidation, error) {
	promo, err := u.repo.GetPromocodeByCode(ctx, code)
	if err != nil {
		if errors.Is(err, repository.ErrPromocodeNotFound) {
			return domain.PromoValidation{Valid: false, Reason: "promo not found"}, nil
		}
		return domain.PromoValidation{}, errutil.Internal("failed to fetch promocode", err)
	}

	if time.Now().After(promo.ExpiresAt) {
		return domain.PromoValidation{Valid: false, Reason: "promo has expired"}, nil
	}

	// Минимальная сумма считается только по товарам (без доставки и сервисного сбора),
	// чтобы поведение совпадало с проверкой в CreateOrder (cartTotalCost)
	if promo.MinOrderAmount != nil && orderAmount < *promo.MinOrderAmount {
		return domain.PromoValidation{Valid: false, Reason: "order amount is below minimum"}, nil
	}

	if promo.RestaurantBrandID != nil && brandID > 0 && *promo.RestaurantBrandID != brandID {
		return domain.PromoValidation{Valid: false, Reason: "promo is not valid for this restaurant"}, nil
	}

	if promo.UserID != nil && *promo.UserID != userID {
		return domain.PromoValidation{Valid: false, Reason: "promo is tied to another user"}, nil
	}

	usedByUser, err := u.repo.CountUserPromocodeUsage(ctx, promo.ID, userID)
	if err != nil {
		return domain.PromoValidation{}, errutil.Internal("failed to check promo usage", err)
	}
	if usedByUser > 0 {
		return domain.PromoValidation{Valid: false, Reason: "promo already used"}, nil
	}
	if promo.MaxUses != nil && promo.CurrentUses >= *promo.MaxUses {
		return domain.PromoValidation{Valid: false, Reason: "max uses reached"}, nil
	}

	return domain.PromoValidation{Valid: true, Discount: computeDiscount(promo, orderAmount)}, nil
}

func (u *promoUseCase) UsePromo(ctx context.Context, userID int64, code, orderPublicID string) error {
	promo, err := u.repo.GetPromocodeByCode(ctx, code)
	if err != nil {
		if errors.Is(err, repository.ErrPromocodeNotFound) {
			return errutil.New("PROMO_NOT_FOUND", "promocode not found", codes.NotFound)
		}
		return errutil.Internal("failed to fetch promocode", err)
	}

	usedByUser, err := u.repo.CountUserPromocodeUsage(ctx, promo.ID, userID)
	if err != nil {
		return errutil.Internal("failed to check promo usage", err)
	}
	if usedByUser > 0 {
		return nil
	}
	if promo.MaxUses != nil && promo.CurrentUses >= *promo.MaxUses {
		return errutil.New("PROMO_LIMIT_REACHED", "promocode usage limit reached", codes.FailedPrecondition)
	}

	var orderID *int64
	if orderPublicID != "" {
		orderID, err = u.repo.ResolveOrderInternalID(ctx, orderPublicID)
		if err != nil {
			return errutil.Internal("failed to resolve order", err)
		}
	}

	if err := u.repo.RecordPromocodeUsage(ctx, promo.ID, userID, orderID); err != nil {
		return errutil.Internal("failed to record promo usage", err)
	}
	return nil
}

func (u *promoUseCase) CreateAndBindWheelPromo(ctx context.Context, userID int64, title string, discountAmount *int64, discountPercent *int, brandID *int64, minOrderAmount *int64) (domain.Promocode, error) {
	code, err := generateWheelPromoCode()
	if err != nil {
		return domain.Promocode{}, errutil.Internal("failed to generate wheel promo code", err)
	}

	maxUses := 1 // Одноразовый промокод

	promo := domain.Promocode{
		Code:              code,
		Title:             title,
		DiscountPercent:   discountPercent,
		DiscountAmount:    discountAmount,
		MaxUses:           &maxUses,
		MinOrderAmount:    minOrderAmount,
		UserID:            &userID,
		RestaurantBrandID: brandID,
		IsGlobal:          brandID == nil,
		ExpiresAt:         time.Now().Add(24 * 3 * time.Hour), // Срок действия 3 дня
	}

	createdPromo, err := u.repo.CreatePromocode(ctx, promo)
	if err != nil {
		return domain.Promocode{}, errutil.Internal("failed to save wheel promocode", err)
	}

	_, err = u.repo.BindPromocodeToUser(ctx, userID, createdPromo.ID)
	if err != nil {
		return domain.Promocode{}, errutil.Internal("failed to bind wheel promocode to user", err)
	}

	return createdPromo, nil
}

func generateWheelPromoCode() (string, error) {
	b := make([]byte, 4) // 4 байта дадут 8 шестнадцатеричных символов
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return "WHL-" + strings.ToUpper(hex.EncodeToString(b)), nil
}

func computeDiscount(promo domain.Promocode, orderAmount int64) int64 {
	var discount int64
	switch {
	case promo.DiscountAmount != nil:
		discount = *promo.DiscountAmount
	case promo.DiscountPercent != nil:
		discount = orderAmount * int64(*promo.DiscountPercent) / 100
	}
	if discount > orderAmount {
		discount = orderAmount
	}
	return discount
}
