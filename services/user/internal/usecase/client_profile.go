package usecase

import (
	"context"
	"errors"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/user/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/user/internal/repository"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/errutil"
	"google.golang.org/grpc/codes"
)

//go:generate mockgen -destination=mocks/client_profile_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/internal/usecase/user ClientProfileUseCase
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

func (u *clientProfileUseCase) CreateProfile(ctx context.Context, accountID int64, idempotencyKey string) error {
	err := u.repo.Create(ctx, accountID, idempotencyKey)
	if err != nil {
		return errutil.Wrap("failed to create client profile in db", err, codes.Internal)
	}

	return nil
}

func (u *clientProfileUseCase) GetByAccountID(ctx context.Context, accountID int64) (domain.ClientProfile, error) {
	profile, err := u.repo.GetByAccountID(ctx, accountID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return domain.ClientProfile{}, errutil.New("client profile not found", codes.NotFound)
		}
		return domain.ClientProfile{}, errutil.Wrap("failed to get client profile from db", err, codes.Internal)
	}

	return profile, nil
}
