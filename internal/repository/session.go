package repository

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	"github.com/google/uuid"
)

// контракт репозитория сессий
//
//go:generate mockgen -destination=mocks/session_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/internal/repository SessionRepository
type SessionRepository interface {
	// метод создания сесии в репозитории
	Create(ctx context.Context, session domain.Session) error
	// метод получения сессии из репозитория по sessionID
	GetByID(ctx context.Context, sessionID uuid.UUID) (domain.Session, error)
	// метод удаления сессии из репозитория по sessionID
	Delete(ctx context.Context, sessionID uuid.UUID) error
}
