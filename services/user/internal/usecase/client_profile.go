package usecase

import (
	"context"
	"errors"

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

	span.SetAttributes(attribute.Int64("profile.bonus_balance", profile.BonusBalance))

	return profile, nil
}
