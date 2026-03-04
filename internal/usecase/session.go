package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/repository"
)

type SessionUseCase interface {
	// бизнес-логика создания сессии для пользователя, вовзращает sessionID
	Create(ctx context.Context, userID int) (string, time.Time, error)
	// проверяет, существует и не истек ли sessionID, возвращает айди юзера при успехе
	Check(ctx context.Context, sessionID string) (int, error)
	// бизнес-логика для удаления сессии, просто вызывает удаление из repository.session
	Destroy(ctx context.Context, sessionId string) error
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

func (u *sessionUseCase) Create(ctx context.Context, userID int) (string, time.Time, error) {
	// бизнес-логика создания сессии
	// возвращает sessionID созданной сессии и момент времени, когда истекает

	// генерация уникальной криптостойкой строки
	sessionID := uuid.New().String()

	// сессия живет sessionTTL времени
	expiresAt := time.Now().Add(u.sessionTTL)

	session := domain.Session{
		ID:        sessionID,
		UserID:    userID,
		ExpiresAt: expiresAt,
	}

	err := u.sessionRepo.Create(ctx, session)
	if err != nil {
		return "", time.Time{}, err
	}

	return sessionID, time.Time{}, nil
}

func (u *sessionUseCase) Check(ctx context.Context, sessionID string) (int, error) {
	// просим репозиторий найти сессию
	session, err := u.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		// сессия не найдена
		return 0, err
	}

	// проверяем срок годности сессии
	if time.Now().After(session.ExpiresAt) {
		// если сессия протухла, удаляем ее из БД
		// игнорируем ошибку удаления, так как для пользователя главное получить ответ, что сессия невалидна
		_ = u.sessionRepo.Delete(ctx, sessionID)

		return 0, domain.ErrSessionExpired
	}

	// возвращаем id юзера в случае успеха
	return session.UserID, nil
}

func (u *sessionUseCase) Destroy(ctx context.Context, sessionID string) error {
	// просто передаем команду удаления куки в репо
	return u.sessionRepo.Delete(ctx, sessionID)
}
