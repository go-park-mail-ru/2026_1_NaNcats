package repository

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
)

type ClientProfileRepository interface {
	Create(ctx context.Context, accountID int64) error
	GetByAccountID(ctx context.Context, accountID int64) (domain.ClientProfile, error)
}
