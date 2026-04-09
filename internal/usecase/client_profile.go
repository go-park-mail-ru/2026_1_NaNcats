package usecase

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/repository"
)

//go:generate mockgen -destination=mocks/client_profile_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/internal/usecase ClientProfileUseCase
type ClientProfileUseCase interface {
	CreateProfile(ctx context.Context, accountID int) error
}

type clientProfileUseCase struct {
	repo repository.ClientProfileRepository
}

func NewClientProfileUseCase(r repository.ClientProfileRepository) ClientProfileUseCase {
	return &clientProfileUseCase{repo: r}
}

func (u *clientProfileUseCase) CreateProfile(ctx context.Context, accountID int) error {
	return u.repo.Create(ctx, accountID)
}
