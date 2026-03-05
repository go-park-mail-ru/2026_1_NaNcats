package repository

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
)

// контракт репозитория сессий
type SessionRepository interface {
	// метод создания сесии в репозитории
	Create(ctx context.Context, session domain.Session) error
	// метод получения сессии из репозитория по sessionID
	GetByID(ctx context.Context, sessionId string) (domain.Session, error)
	// метод удаления сессии из репозитория по sessionID
	Delete(ctx context.Context, sessionId string) error
}
