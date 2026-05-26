package usecase

import (
	"context"
	"errors"
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
}

type clientProfileUseCase struct {
	repo repository.ClientProfileRepository
}

func NewClientProfileUseCase(r repository.ClientProfileRepository) ClientProfileUseCase {
	return &clientProfileUseCase{repo: r}
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
