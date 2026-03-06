package usecase

import (
	"context"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

// контракт бизнес-логики авторизации
type AuthUseCase interface {
	Register(ctx context.Context, user domain.User) (domain.User, domain.Session, error)
	Login(ctx context.Context, user domain.User) (domain.User, domain.Session, error)
	Check(ctx context.Context, sessionID string) (domain.User, error)
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

// бизнес-логика регистрации
func (u *authUseCase) Register(ctx context.Context, user domain.User) (domain.User, domain.Session, error) {
	// генерируем хешированный пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.PasswordHash), bcrypt.DefaultCost)
	if err != nil {
		return domain.User{}, domain.Session{}, err
	}

	user.PasswordHash = string(hashedPassword)

	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	// вызов создания пользователя из репо
	id, err := u.userRepo.CreateUser(user)
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
	currUser, err := u.userRepo.GetUserByEmail(user.Email)
	if err != nil {
		return domain.User{}, domain.Session{}, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(currUser.PasswordHash))
	if err != nil {
		return domain.User{}, domain.Session{}, err
	}

	createdSession, err := u.sessionUC.Create(ctx, currUser.ID)
	if err != nil {
		return domain.User{}, domain.Session{}, err
	}

	return currUser, createdSession, nil
}

func (u *authUseCase) Check(ctx context.Context, sessionID string) (domain.User, error) {
	userID, err := u.sessionUC.Check(ctx, sessionID)
	if err != nil {
		return domain.User{}, err
	}

	user, err := u.userRepo.GetUserByID(userID)
	if err != nil {
		return domain.User{}, err
	}

	return user, nil
}
