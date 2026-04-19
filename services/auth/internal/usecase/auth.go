package usecase

import (
	"context"
	"strings"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/auth/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/errutil"
	passUtil "github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/password"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
)

// контракт бизнес-логики авторизации
//
//go:generate mockgen -destination=../mocks/auth_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/internal/usecase/auth AuthUseCase
type AuthUseCase interface {
	IssueSession(ctx context.Context, userID int64, role, userAgent string) (domain.Session, error)
	Login(ctx context.Context, email, password, userAgent string) (domain.Session, error)
	Logout(ctx context.Context, sessionID uuid.UUID) error
	CheckUserSession(ctx context.Context, sessionID uuid.UUID) (int64, error)
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
	createdSession, err := u.sessionUC.Create(ctx, userID, role, userAgent)
	if err != nil {
		return domain.Session{}, errutil.Wrap("failed to issue session", err, codes.Internal)
	}

	return createdSession, nil
}

func (u *authUseCase) Login(ctx context.Context, email, password, userAgent string) (domain.Session, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	currUser, err := u.userClient.GetUserByEmail(ctx, email)
	if err != nil {
		return domain.Session{}, errutil.New("invalid email or password", codes.Unauthenticated)
	}

	isValid, err := passUtil.VerifyPassword(password, currUser.PasswordHash)
	if err != nil || !isValid {
		return domain.Session{}, errutil.New("invalid email or password", codes.Unauthenticated)
	}

	createdSession, err := u.sessionUC.Create(ctx, currUser.ID, currUser.Role, userAgent)
	if err != nil {
		return domain.Session{}, errutil.Wrap("failed to create session", err, codes.Internal)
	}

	return createdSession, nil
}

func (u *authUseCase) Logout(ctx context.Context, sessionID uuid.UUID) error {
	err := u.sessionUC.Destroy(ctx, sessionID)
	if err != nil {
		return errutil.Wrap("failed to destroy session", err, codes.Internal)
	}

	return nil
}

// возвращает пользователя сессии, проверяя, существует ли сессия и пользователь сессии
func (u *authUseCase) CheckUserSession(ctx context.Context, sessionID uuid.UUID) (int64, error) {
	session, err := u.sessionUC.Check(ctx, sessionID)
	if err != nil {
		// Ошибка уже обернута в SessionUseCase
		return 0, err
	}

	return session.UserID, nil
}

func (u *authUseCase) SetCSRFForUser(ctx context.Context, sessionID uuid.UUID) (string, error) {
	return u.sessionUC.SetCSRF(ctx, sessionID)
}

func (u *authUseCase) GetCSRFBySessionID(ctx context.Context, sessionID uuid.UUID) (string, error) {
	return u.sessionUC.GetCSRF(ctx, sessionID)
}
