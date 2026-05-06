package usecase

import (
	"context"
	"strings"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/auth/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/errutil"
	passUtil "github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/password"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
)

// контракт бизнес-логики авторизации
//
//go:generate mockgen -destination=mocks/auth_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/services/auth/internal/usecase AuthUseCase,UserClient
//go:generate gowrap gen -i AuthUseCase -t ../../../../shared/templates/tracing.tmpl -o auth_tracing_mw.go -v TracerName=auth-service

type UserClient interface {
	GetUserByEmail(ctx context.Context, email string) (domain.User, error)
}

type AuthUseCase interface {
	IssueSession(ctx context.Context, userID int64, role, userAgent string) (domain.Session, error)
	Login(ctx context.Context, email, password, userAgent string) (domain.Session, error)
	Logout(ctx context.Context, sessionID uuid.UUID) error
	CheckUserSession(ctx context.Context, sessionID uuid.UUID) (int64, string, error)
	SetCSRFForUser(ctx context.Context, sessionID uuid.UUID) (string, error)
	GetCSRFBySessionID(ctx context.Context, sessionID uuid.UUID) (string, error)
}

// реализация контракта
type authUseCase struct {
	userClient UserClient
	sessionUC  SessionUseCase
}

// функция-конструктор бизнес-логики авторизации
func NewAuthUseCase(uc UserClient, suc SessionUseCase) AuthUseCase {
	return &authUseCase{
		userClient: uc,
		sessionUC:  suc,
	}
}

func (u *authUseCase) IssueSession(ctx context.Context, userID int64, role, userAgent string) (domain.Session, error) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.Int64("user.id", userID),
		attribute.String("user.role", role),
	)

	createdSession, err := u.sessionUC.Create(ctx, userID, role, userAgent)
	if err != nil {
		return domain.Session{}, errutil.Internal("failed to issue session", err)
	}

	return createdSession, nil
}

func (u *authUseCase) Login(ctx context.Context, email, password, userAgent string) (domain.Session, error) {
	span := trace.SpanFromContext(ctx)
	email = strings.ToLower(strings.TrimSpace(email))

	currUser, err := u.userClient.GetUserByEmail(ctx, email)
	if err != nil {
		return domain.Session{}, errutil.New("INVALID_CREDENTIALS", "invalid email or password", codes.Unauthenticated)
	}
	span.SetAttributes(attribute.Int64("user.id", currUser.ID))

	isValid, err := passUtil.VerifyPassword(password, currUser.PasswordHash)
	if err != nil || !isValid {
		return domain.Session{}, errutil.New("INVALID_CREDENTIALS", "invalid email or password", codes.Unauthenticated)
	}

	createdSession, err := u.sessionUC.Create(ctx, currUser.ID, currUser.Role, userAgent)
	if err != nil {
		return domain.Session{}, errutil.Internal("failed to create session", err)
	}

	return createdSession, nil
}

func (u *authUseCase) Logout(ctx context.Context, sessionID uuid.UUID) error {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.String("session.id", sessionID.String()))

	err := u.sessionUC.Destroy(ctx, sessionID)
	if err != nil {
		return errutil.Internal("failed to destroy session", err)
	}

	return nil
}

// возвращает пользователя сессии, проверяя, существует ли сессия и пользователь сессии
func (u *authUseCase) CheckUserSession(ctx context.Context, sessionID uuid.UUID) (int64, string, error) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.String("session.id", sessionID.String()))

	session, err := u.sessionUC.Check(ctx, sessionID)
	if err != nil {
		return 0, "", err
	}

	span.SetAttributes(
		attribute.Int64("user.id", session.UserID),
		attribute.String("user.role", session.Role),
	)

	return session.UserID, session.Role, nil
}

func (u *authUseCase) SetCSRFForUser(ctx context.Context, sessionID uuid.UUID) (string, error) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.String("session.id", sessionID.String()))

	return u.sessionUC.SetCSRF(ctx, sessionID)
}

func (u *authUseCase) GetCSRFBySessionID(ctx context.Context, sessionID uuid.UUID) (string, error) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.String("session.id", sessionID.String()))

	return u.sessionUC.GetCSRF(ctx, sessionID)
}
