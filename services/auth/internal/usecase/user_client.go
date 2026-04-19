package usecase

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/auth/internal/domain"
)

type UserClient interface {
	GetUserByEmail(ctx context.Context, email string) (domain.User, error)
}
