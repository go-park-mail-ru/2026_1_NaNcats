package repository

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
)

type SessionRepository interface {
	Create(ctx context.Context, session domain.Session) error
	GetByID(ctx context.Context, sessionId string) (domain.Session, error)
	Delete(ctx context.Context, sessionId string) error
}
