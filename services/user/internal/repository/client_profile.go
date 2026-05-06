package repository

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/user/internal/domain"
)

//go:generate mockgen -destination=mocks/client_profile_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/services/user/internal/repository ClientProfileRepository
type ClientProfileRepository interface {
	Create(ctx context.Context, accountID int64, idempotencyKey string) error
	GetByAccountID(ctx context.Context, accountID int64) (domain.ClientProfile, error)
}
