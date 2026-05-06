package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/auth/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/auth/internal/repository"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/csrf"
)

//go:generate mockgen -destination=mocks/session_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/services/auth/internal/usecase SessionUseCase
//go:generate gowrap gen -i SessionUseCase -t ../../../../shared/templates/tracing.tmpl -o session_tracing_mw.go -v TracerName=auth-service
type SessionUseCase interface {
	// бизнес-логика создания сессии для пользователя, вовзращает sessionID
	Create(ctx context.Context, userID int64, role, userAgent string) (domain.Session, error)
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

func (u *sessionUseCase) Create(ctx context.Context, userID int64, role, userAgent string) (domain.Session, error) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.Int64("user.id", userID),
		attribute.String("user.role", role),
	)

	sessionID := uuid.New()
	expiresAt := time.Now().Add(u.sessionTTL)

	span.SetAttributes(attribute.String("session.id", sessionID.String()))

	session := domain.Session{
		ID:        sessionID,
		UserID:    userID,
		UserAgent: userAgent,
		Role:      role,
		ExpiresAt: expiresAt,
	}

	err := u.sessionRepo.Create(ctx, session, u.sessionTTL)
	if err != nil {
		return domain.Session{}, err
	}

	return session, nil
}

// проверяет, существует ли сессия, если да - возвращаем id пользователя сессии
func (u *sessionUseCase) Check(ctx context.Context, id uuid.UUID) (domain.Session, error) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.String("session.id", id.String()))

	session, err := u.sessionRepo.GetByID(ctx, id)
	if err != nil {
		return domain.Session{}, err
	}

	span.SetAttributes(
		attribute.Int64("user.id", session.UserID),
		attribute.String("user.role", session.Role),
	)

	if time.Now().After(session.ExpiresAt) {
		span.AddEvent("session_expired")
		return domain.Session{}, domain.ErrSessionExpired
	}

	return session, nil
}

func (u *sessionUseCase) Destroy(ctx context.Context, id uuid.UUID) error {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.String("session.id", id.String()))

	return u.sessionRepo.Delete(ctx, id)
}

func (u *sessionUseCase) SetCSRF(ctx context.Context, sessionID uuid.UUID) (string, error) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.String("session.id", sessionID.String()))

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
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.String("session.id", sessionID.String()))

	token, err := u.sessionRepo.GetCSRF(ctx, sessionID)
	if err != nil {
		return "", err
	}

	if token == "" {
		span.AddEvent("csrf_token_not_found_generating_new")
		token, err = u.SetCSRF(ctx, sessionID)
		if err != nil {
			return "", err
		}
	}

	return token, nil
}
