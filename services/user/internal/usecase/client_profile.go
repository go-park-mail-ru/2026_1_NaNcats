package user

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/user/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/user/internal/repository"
)

//go:generate mockgen -destination=mocks/client_profile_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/internal/usecase/user ClientProfileUseCase
type ClientProfileUseCase interface {
	CreateProfile(ctx context.Context, accountID int64) error
	GetByAccountID(ctx context.Context, accountID int64) (domain.ClientProfile, error)
}

type clientProfileUseCase struct {
	repo repository.ClientProfileRepository
}

func NewClientProfileUseCase(r repository.ClientProfileRepository) ClientProfileUseCase {
	return &clientProfileUseCase{repo: r}
}

func (u *clientProfileUseCase) CreateProfile(ctx context.Context, accountID int64) error {
	return u.repo.Create(ctx, accountID)
}

func (u *clientProfileUseCase) GetByAccountID(ctx context.Context, accountID int64) (domain.ClientProfile, error) {
	return u.repo.GetByAccountID(ctx, accountID)
}
