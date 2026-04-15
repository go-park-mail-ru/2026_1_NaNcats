package auth

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/repository"
	"github.com/go-park-mail-ru/2026_1_NaNcats/pkg/csrf"
)

//go:generate mockgen -destination=mocks/session_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/internal/usecase/auth SessionUseCase
type SessionUseCase interface {
	// бизнес-логика создания сессии для пользователя, вовзращает sessionID
	Create(ctx context.Context, userID int, userAgent string) (domain.Session, error)
	// проверяет, существует и не истек ли sessionID, возвращает айди юзера при успехе
	Check(ctx context.Context, id uuid.UUID) (domain.Session, error)
	// бизнес-логика для удаления сессии, просто вызывает удаление из repository.session
	Destroy(ctx context.Context, id uuid.UUID) error
	SetCSRF(ctx context.Context, sessionID uuid.UUID) (string, error)
	GetCSRF(ctx context.Context, sessionID uuid.UUID) (string, error)
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

func (u *sessionUseCase) Create(ctx context.Context, userID int, userAgent string) (domain.Session, error) {
	// бизнес-логика создания сессии
	// возвращает sessionID созданной сессии и момент времени, когда истекает

	// генерация уникальной криптостойкой строки
	sessionID := uuid.New()
	expiresAt := time.Now().Add(u.sessionTTL)

	// создаем новый объект сессии
	session := domain.Session{
		ID:        sessionID,
		UserID:    userID,
		UserAgent: userAgent,
		ExpiresAt: expiresAt,
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
		return domain.Session{}, domain.ErrSessionExpired
	}

	// возвращаем id юзера в случае успеха
	return session, nil
}

func (u *sessionUseCase) Destroy(ctx context.Context, id uuid.UUID) error {
	// просто передаем команду удаления куки в репо
	return u.sessionRepo.Delete(ctx, id)
}

func (u *sessionUseCase) SetCSRF(ctx context.Context, sessionID uuid.UUID) (string, error) {
	csrfToken, err := csrf.GenerateToken()
	if err != nil {
		return "", err
	}

	err = u.sessionRepo.SetCSRF(ctx, sessionID, csrfToken)
	if err != nil {
		return "", err
	}

	return csrfToken, nil
}

func (u *sessionUseCase) GetCSRF(ctx context.Context, sessionID uuid.UUID) (string, error) {
	token, err := u.sessionRepo.GetCSRF(ctx, sessionID)
	if err != nil {
		return "", err
	}

	if token == "" {
		token, err = u.SetCSRF(ctx, sessionID)
		if err != nil {
			return "", err
		}
	}

	return token, nil
}
