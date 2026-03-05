package memory

import (
	"context"
	"sync"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/repository"
)

// структура репозитория сессий на основе мап
type sessionRepo struct {
	mu       sync.RWMutex              // защита от одновременного чтения из мапы
	sessions map[string]domain.Session // мапа сессий, ключ - sessionID
}

// функция-конструктор репозитория сессий
func NewSessionRepo() repository.SessionRepository {
	return &sessionRepo{
		sessions: make(map[string]domain.Session),
	}
}

func (r *sessionRepo) Create(ctx context.Context, session domain.Session) error {
	// добавить обработку ошибок
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sessions[session.ID] = session
	return nil
}

func (r *sessionRepo) GetByID(ctx context.Context, sessionID string) (domain.Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	session, exists := r.sessions[sessionID]
	if !exists {
		return domain.Session{}, domain.ErrSessionNotFound
	}

	return session, nil
}

func (r *sessionRepo) Delete(ctx context.Context, sessionID string) error {
	// добавить обработку ошибок
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.sessions, sessionID)
	return nil
}
