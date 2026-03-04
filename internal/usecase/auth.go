package usecase

import (
	"context"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type AuthUseCase interface {
	Register(ctx context.Context, user domain.User) (domain.User, string, time.Time, error)
}

type authUseCase struct {
	userRepo  repository.UserRepository
	sessionUC SessionUseCase
}

func NewAuthUseCase(ur repository.UserRepository, suc SessionUseCase) AuthUseCase {
	return &authUseCase{
		userRepo:  ur,
		sessionUC: suc,
	}
}

func (u *authUseCase) Register(ctx context.Context, user domain.User) (domain.User, string, time.Time, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.PasswordHash), bcrypt.DefaultCost)
	if err != nil {
		return domain.User{}, "", time.Time{}, err
	}

	user.PasswordHash = string(hashedPassword)

	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	id, err := u.userRepo.CreateUser(user)
	if err != nil {
		return domain.User{}, "", time.Time{}, err
	}

	user.ID = id

	sessionID, expiresAt, err := u.sessionUC.Create(ctx, user.ID)
	if err != nil {
		return domain.User{}, "", time.Time{}, err
	}

	return user, sessionID, expiresAt, nil
}
