package usecase

import (
	"context"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/repository"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// контракт бизнес-логики авторизации
//
//go:generate mockgen -destination=mocks/auth_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/internal/usecase AuthUseCase
type AuthUseCase interface {
	Register(ctx context.Context, user domain.User) (domain.User, domain.Session, error)
	Login(ctx context.Context, user domain.User) (domain.User, domain.Session, error)
	Logout(ctx context.Context, sessionID uuid.UUID) error
	Check(ctx context.Context, sessionID uuid.UUID) (domain.User, error)
	GetProfile(ctx context.Context, userID uuid.UUID) (domain.User, error)
}

// реализация контракта
type authUseCase struct {
	userRepo  repository.UserRepository
	sessionUC SessionUseCase
}

// функция-конструктор бизнес-логики авторизации
func NewAuthUseCase(ur repository.UserRepository, suc SessionUseCase) AuthUseCase {
	return &authUseCase{
		userRepo:  ur,
		sessionUC: suc,
	}
}

func isValidEmail(email string) bool {
	if len(email) < 4 || len(email) > 254 {
		return false
	}

	// Парсинг RFC 5322
	addr, err := mail.ParseAddress(email)
	if err != nil {
		return false
	}

	if strings.Contains(email, "..") {
		return false
	}

	// mail.ParseAddress позволяет вводить "Name <test@test.com>"
	// Нам же нужно, чтобы введенная строка была только email-ом
	if addr.Address != email {
		return false
	}

	return true
}

// бизнес-логика регистрации
func (u *authUseCase) Register(ctx context.Context, user domain.User) (domain.User, domain.Session, error) {
	user.Email = strings.ToLower(strings.TrimSpace(user.Email))

	if !isValidEmail(user.Email) {
		return domain.User{}, domain.Session{}, domain.ErrInvalidEmail
	}

	if len(user.PasswordHash) < 8 {
		return domain.User{}, domain.Session{}, domain.ErrInvalidPassword
	}

	// генерируем хешированный пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.PasswordHash), bcrypt.DefaultCost)
	if err != nil {
		return domain.User{}, domain.Session{}, fmt.Errorf("bcrypt failed: %w", err)
	}

	user.PasswordHash = string(hashedPassword)

	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	// вызов создания пользователя из репо
	id, err := u.userRepo.CreateUser(ctx, user)
	if err != nil {
		return domain.User{}, domain.Session{}, err
	}

	user.ID = id

	// вызов бизнес-логики по созданию сессии
	createdSession, err := u.sessionUC.Create(ctx, user.ID)
	if err != nil {
		return domain.User{}, domain.Session{}, err
	}

	return user, createdSession, nil
}

func (u *authUseCase) Login(ctx context.Context, user domain.User) (domain.User, domain.Session, error) {
	user.Email = strings.ToLower(strings.TrimSpace(user.Email))

	currUser, err := u.userRepo.GetUserByEmail(ctx, user.Email)
	if err != nil {
		return domain.User{}, domain.Session{}, domain.ErrInvalidCredentials
	}

	err = bcrypt.CompareHashAndPassword([]byte(currUser.PasswordHash), []byte(user.PasswordHash))
	if err != nil {
		return domain.User{}, domain.Session{}, domain.ErrInvalidCredentials
	}

	createdSession, err := u.sessionUC.Create(ctx, currUser.ID)
	if err != nil {
		return domain.User{}, domain.Session{}, fmt.Errorf("failed to create session: %w", err)
	}

	return currUser, createdSession, nil
}

func (u *authUseCase) Logout(ctx context.Context, sessionID uuid.UUID) error {
	err := u.sessionUC.Destroy(ctx, sessionID)
	if err != nil {
		return err
	}

	return nil
}

// возвращает пользователя сессии, проверяя, существует ли сессия и пользователь сессии
func (u *authUseCase) Check(ctx context.Context, sessionID uuid.UUID) (domain.User, error) {
	userID, err := u.sessionUC.Check(ctx, sessionID)
	if err != nil {
		return domain.User{}, err
	}

	user, err := u.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return domain.User{}, err
	}

	return user, nil
}

// возвращает юзера по переданному userID
func (u *authUseCase) GetProfile(ctx context.Context, userID uuid.UUID) (domain.User, error) {
	user, err := u.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return domain.User{}, err
	}

	return user, nil
}
