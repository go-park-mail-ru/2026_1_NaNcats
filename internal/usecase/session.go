package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/repository"
)

//go:generate mockgen -destination=mocks/session_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/internal/usecase SessionUseCase
type SessionUseCase interface {
	// бизнес-логика создания сессии для пользователя, вовзращает sessionID
	Create(ctx context.Context, user domain.User) (domain.Session, error)
	// проверяет, существует и не истек ли sessionID, возвращает айди юзера при успехе
	Check(ctx context.Context, id uuid.UUID) (domain.Session, error)
	// бизнес-логика для удаления сессии, просто вызывает удаление из repository.session
	Destroy(ctx context.Context, id uuid.UUID) error
}

// структура usecase сессий на основе мап
type sessionUseCase struct {
	sessionRepo repository.SessionRepository
	sessionTTL  time.Duration
}

func NewSessionUseCase(sr repository.SessionRepository, ttl time.Duration) SessionUseCase {
	return &sessionUseCase{
		sessionRepo: sr,
		sessionTTL:  ttl,
	}
}

func (u *sessionUseCase) Create(ctx context.Context, user domain.User) (domain.Session, error) {
	// бизнес-логика создания сессии
	// возвращает sessionID созданной сессии и момент времени, когда истекает

	// генерация уникальной криптостойкой строки
	sessionID := uuid.New()

	// создаем новый объект сессии
	session := domain.Session{
		ID:     sessionID,
		UserID: user.ID,
	}

	// вызов создания сессии в репо
	err := u.sessionRepo.Create(ctx, session, u.sessionTTL)
	if err != nil {
		return domain.Session{}, err
	}

	return session, nil
}

// проверяет, существует ли сессия, если да - возвращаем id пользователя сессии
func (u *sessionUseCase) Check(ctx context.Context, id uuid.UUID) (domain.Session, error) {
	// просим репозиторий найти сессию
	session, err := u.sessionRepo.GetByID(ctx, id)
	if err != nil {
		// сессия не найдена
		return domain.Session{}, err
	}

	if time.Now().After(session.ExpiresAt) {
		return domain.Session{}, fmt.Errorf("session expired")
	}

	// возвращаем id юзера в случае успеха
	return session, nil
}

func (u *sessionUseCase) Destroy(ctx context.Context, id uuid.UUID) error {
	// просто передаем команду удаления куки в репо
	return u.sessionRepo.Delete(ctx, id)
}
