package usecase

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/user/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/user/internal/repository"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/errutil"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
)

//go:generate mockgen -destination=mocks/client_profile_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/services/user/internal/usecase ClientProfileUseCase
//go:generate gowrap gen -i ClientProfileUseCase -t ../../../../shared/templates/tracing.tmpl -o client_profile_tracing_mw.go -v TracerName=user-service
type ClientProfileUseCase interface {
	CreateProfile(ctx context.Context, accountID int64, idempotencyKey string) error
	GetByAccountID(ctx context.Context, accountID int64) (domain.ClientProfile, error)
	ActivateStreakFreeze(ctx context.Context, accountID int64) error
	IncrementStreak(ctx context.Context, accountID int64) error

	ClaimWheelSpin(ctx context.Context, accountID int64) error
	ResetWheelSpinCooldown(ctx context.Context, accountID int64) error
	GetWheelSectors(ctx context.Context) ([]WheelSector, error)
	SpinWheel(ctx context.Context, userID int64) (WheelSpinResult, error)
}

type OrderClient interface {
	CreateAndBindWheelPromo(
		ctx context.Context,
		userID int64,
		title string,
		discountAmount *int64,
		discountPercent *int,
		brandID *int64,
		minOrderAmount *int64,
	) (string, *string, error)

	GetTrendingBrands(ctx context.Context, windowDays, limit int32) ([]int64, error)
}

// Описывает результат прокрутки колеса
type WheelSpinResult struct {
	SectorID   int
	SectorName string
	Emoji      string
	PromoCode  *string
	ExpiresAt  *string
	Message    string
}

type WheelSector struct {
	ID     int
	Name   string
	Emoji  string
	Weight int
}

var sectors = []WheelSector{
	{ID: 1, Name: "Попробуй в следующий раз", Emoji: "🍂", Weight: 300},
	{ID: 2, Name: "Персональная скидка", Emoji: "💸", Weight: 150},
	{ID: 3, Name: "Скидка на популярный бренд", Emoji: "🔥", Weight: 150},
	{ID: 4, Name: "Заморозка стрика", Emoji: "❄️", Weight: 100},
	{ID: 5, Name: "Стрик-буст", Emoji: "🚀", Weight: 100},
	{ID: 6, Name: "Эксклюзивная ачивка", Emoji: "🎡", Weight: 20},
	{ID: 7, Name: "Супер-приз", Emoji: "🎁", Weight: 10},
	{ID: 8, Name: "Реролл", Emoji: "🌀", Weight: 120},
	{ID: 9, Name: "Бесплатная доставка", Emoji: "🛵", Weight: 50},
}

type clientProfileUseCase struct {
	repo          repository.ClientProfileRepository
	orderClient   OrderClient
	achievementUC AchievementUseCase
}

func NewClientProfileUseCase(
	r repository.ClientProfileRepository,
	oc OrderClient,
	auc AchievementUseCase,
) ClientProfileUseCase {
	return &clientProfileUseCase{
		repo:          r,
		orderClient:   oc,
		achievementUC: auc,
	}
}

func (u *clientProfileUseCase) SpinWheel(ctx context.Context, userID int64) (WheelSpinResult, error) {
	err := u.repo.ClaimWheelSpin(ctx, userID)
	if err != nil {
		return WheelSpinResult{}, err
	}

	// Разыгрываем случайный сектор
	totalWeight := 0
	for _, s := range sectors {
		totalWeight += s.Weight
	}

	randVal := rand.Intn(totalWeight)
	var wonSector WheelSector
	currentSum := 0

	for _, s := range sectors {
		currentSum += s.Weight
		if randVal < currentSum {
			wonSector = s
			break
		}
	}

	res := WheelSpinResult{
		SectorID:   wonSector.ID,
		SectorName: wonSector.Name,
		Emoji:      wonSector.Emoji,
	}

	switch wonSector.ID {
	case 1: // Попробуй в следующий раз
		res.Message = "Увы, в этот раз ничего не выпало. Попробуйте завтра!"

	case 2: // Персональная скидка (150 руб)
		discountAmount := int64(150_000_000)
		code, expiresAt, err := u.orderClient.CreateAndBindWheelPromo(ctx, userID, "Персональная скидка 150 рублей", &discountAmount, nil, nil, nil)
		if err != nil {
			return WheelSpinResult{}, err
		}
		res.PromoCode = &code
		res.ExpiresAt = expiresAt
		res.Message = "Поздравляем! Вы выиграли персональную скидку 150 рублей на любой заказ!"

	case 3: // Скидка на случайный популярный бренд (15%)
		brandIDs, err := u.orderClient.GetTrendingBrands(ctx, 7, 10)
		var selectedBrandID int64

		if err != nil || len(brandIDs) == 0 {
			selectedBrandID = 1
		} else {
			selectedBrandID = brandIDs[rand.Intn(len(brandIDs))]
		}

		discountPercent := 15
		code, expiresAt, err := u.orderClient.CreateAndBindWheelPromo(ctx, userID, "Скидка 15% на популярный бренд", nil, &discountPercent, &selectedBrandID, nil)
		if err != nil {
			return WheelSpinResult{}, err
		}
		res.PromoCode = &code
		res.ExpiresAt = expiresAt
		res.Message = "Вы выиграли скидку 15% на популярный ресторан!"

	case 4: // Заморозка стрика
		err := u.repo.UpdateStreakFreeze(ctx, userID, true)
		if err != nil {
			return WheelSpinResult{}, err
		}
		res.Message = "Заморозка серии активна! Теперь ваша серия заказов защищена, если вы пропустите неделю."

	case 5: // Стрик-буст
		err := u.repo.IncrementStreak(ctx, userID)
		if err != nil {
			return WheelSpinResult{}, err
		}
		res.Message = "Ура! Ваша серия заказов увеличена на +1 неделю!"

	case 6: // Эксклюзивная ачивка
		err := u.achievementUC.OnWheelSpin(ctx, userID, "lucky_wheel_winner")
		if err != nil {
			return WheelSpinResult{}, err
		}
		res.Message = "Невероятное везение! Вы получили редкую коллекционную ачивку «Любимчик Пиццули»!"

	case 7: // Супер-приз
		discountAmount := int64(500_000_000)
		minOrderAmount := int64(1_500_000_000)
		code, expiresAt, err := u.orderClient.CreateAndBindWheelPromo(ctx, userID, "Супер-приз: скидка 500 рублей", &discountAmount, nil, nil, &minOrderAmount)
		if err != nil {
			return WheelSpinResult{}, err
		}
		res.PromoCode = &code
		res.ExpiresAt = expiresAt
		res.Message = "Мега-приз! Вы выиграли скидку 500 рублей на заказ от 1500 рублей!"

	case 8: // Реролл
		err := u.repo.ResetWheelSpinCooldown(ctx, userID)
		if err != nil {
			return WheelSpinResult{}, err
		}
		res.Message = "Вам выпала еще одна попытка! Вращайте колесо прямо сейчас!"

	case 9: // Бесплатная доставка
		discountAmount := int64(360_000_000)
		code, expiresAt, err := u.orderClient.CreateAndBindWheelPromo(ctx, userID, "Бесплатная доставка (скидка 360 рублей)", &discountAmount, nil, nil, nil)
		if err != nil {
			return WheelSpinResult{}, err
		}
		res.PromoCode = &code
		res.ExpiresAt = expiresAt
		res.Message = "Поздравляем! Вы выиграли промокод на бесплатную доставку!"
	}

	if wonSector.ID != 6 {
		_ = u.achievementUC.OnWheelSpin(ctx, userID, "")
	}

	return res, nil
}

func (u *clientProfileUseCase) GetWheelSectors(ctx context.Context) ([]WheelSector, error) {
	return sectors, nil
}

func (u *clientProfileUseCase) CreateProfile(ctx context.Context, accountID int64, idempotencyKey string) (err error) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.Int64("user.id", accountID),
		attribute.String("idempotency_key", idempotencyKey),
	)

	err = u.repo.Create(ctx, accountID, idempotencyKey)
	if err != nil {
		return errutil.Internal("failed to create client profile in db", err)
	}

	return nil
}

func (u *clientProfileUseCase) GetByAccountID(ctx context.Context, accountID int64) (profile domain.ClientProfile, err error) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.Int64("user.id", accountID))

	profile, err = u.repo.GetByAccountID(ctx, accountID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return domain.ClientProfile{}, errutil.New("PROFILE_NOT_FOUND", "client profile not found", codes.NotFound)
		}
		return domain.ClientProfile{}, errutil.Internal("failed to get client profile from db", err)
	}

	if err := u.syncStreak(ctx, &profile); err != nil {
		return domain.ClientProfile{}, errutil.Internal("failed to sync user streak", err)
	}

	span.SetAttributes(
		attribute.Int64("profile.bonus_balance", profile.BonusBalance),
		attribute.Int("profile.streak_count", profile.StreakCount),
	)

	return profile, nil
}

func (u *clientProfileUseCase) ActivateStreakFreeze(ctx context.Context, accountID int64) error {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.Int64("user.id", accountID))

	profile, err := u.repo.GetByAccountID(ctx, accountID)
	if err != nil {
		return err
	}

	// Синхронизируем стрик перед активацией заморозки (чтобы не заморозить уже сгоревший)
	if err := u.syncStreak(ctx, &profile); err != nil {
		return err
	}

	err = u.repo.UpdateStreakFreeze(ctx, accountID, true)
	if err != nil {
		return errutil.Internal("failed to activate streak freeze", err)
	}

	return nil
}

func (u *clientProfileUseCase) IncrementStreak(ctx context.Context, accountID int64) error {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.Int64("user.id", accountID))

	profile, err := u.repo.GetByAccountID(ctx, accountID)
	if err != nil {
		return err
	}

	// Синхронизируем стрик перед бустом (чтобы не прибавить +1 к сгоревшей серии)
	if err := u.syncStreak(ctx, &profile); err != nil {
		return err
	}

	err = u.repo.IncrementStreak(ctx, accountID)
	if err != nil {
		return errutil.Internal("failed to increment streak", err)
	}

	return nil
}

func (u *clientProfileUseCase) syncStreak(ctx context.Context, p *domain.ClientProfile) error {
	if p.LastOrderDate == nil {
		return nil
	}

	currentMonday := startOfISOWeek(time.Now())
	lastOrderMonday := startOfISOWeek(*p.LastOrderDate)

	// Серия прервана, если текущий понедельник строго позже, чем понедельник последнего заказа + 1 неделя
	if currentMonday.After(lastOrderMonday.AddDate(0, 0, 7)) {
		if p.StreakFreezeActive {
			// Расходуем замороженную серию: в БД сбрасываем флаг в false
			if err := u.repo.UpdateStreakFreeze(ctx, p.AccountID, false); err != nil {
				return err
			}
			p.StreakFreezeActive = false

			// Виртуально продлеваем дату последнего заказа на текущую неделю, чтобы спасти серию
			now := time.Now()
			p.LastOrderDate = &now
		} else {
			// Заморозки нет - сбрасываем серию в 0
			if err := u.repo.ResetStreak(ctx, p.AccountID); err != nil {
				return err
			}
			p.StreakCount = 0
			p.LastOrderDate = nil
		}
	}
	return nil
}

func startOfISOWeek(t time.Time) time.Time {
	wd := int(t.Weekday())
	if wd == 0 {
		wd = 7 // воскресенье имеет индекс 0, переводим в 7
	}
	monday := t.AddDate(0, 0, -wd+1)
	return time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, t.Location())
}

func (u *clientProfileUseCase) ClaimWheelSpin(ctx context.Context, accountID int64) error {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.Int64("user.id", accountID))

	err := u.repo.ClaimWheelSpin(ctx, accountID)
	if err != nil {
		return err
	}

	return nil
}

func (u *clientProfileUseCase) ResetWheelSpinCooldown(ctx context.Context, accountID int64) error {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.Int64("user.id", accountID))

	err := u.repo.ResetWheelSpinCooldown(ctx, accountID)
	if err != nil {
		return err
	}

	return nil
}
