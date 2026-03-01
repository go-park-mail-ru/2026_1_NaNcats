package usecase

import (
	"context"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type AuthUseCase interface {
	Register(ctx context.Context, user domain.User) (domain.User, error)
}

type authUseCase struct {
	userRepo repository.UserRepository
}

func NewAuthUseCase(ur repository.UserRepository) AuthUseCase {
	return &authUseCase{
		userRepo: ur,
	}
}

func (u *authUseCase) Register(ctx context.Context, user domain.User) (domain.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.PasswordHash), bcrypt.DefaultCost)
	if err != nil {
		return domain.User{}, err
	}

	user.PasswordHash = string(hashedPassword)

	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	id, err := u.userRepo.CreateUser(user)
	if err != nil {
		return domain.User{}, err
	}

	user.ID = id
	return user, nil
}
