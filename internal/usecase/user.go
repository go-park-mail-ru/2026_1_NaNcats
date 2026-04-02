package usecase

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/repository"
)

//go:generate mockgen -destination=mocks/user_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/internal/usecase UserUseCase
type UserUseCase interface {
	Create(ctx context.Context, user domain.User) (int, error)
	GetByID(ctx context.Context, userID int) (domain.User, error)
	GetByEmail(ctx context.Context, email string) (domain.User, error)
	Check(ctx context.Context, userID int) (bool, error)
	UpdateProfile(ctx context.Context, userID int, name, email *string) error
}

type userUseCase struct {
	userRepo repository.UserRepository
}

func NewUserUseCase(ur repository.UserRepository) UserUseCase {
	return &userUseCase{
		userRepo: ur,
	}
}

// создаем юзера
func (u *userUseCase) Create(ctx context.Context, user domain.User) (int, error) {
	id, err := u.userRepo.CreateUser(ctx, user)
	if err != nil {
		return 0, err
	}

	return id, nil
}

// возвращает юзера по переданному userID
func (u *userUseCase) GetByID(ctx context.Context, userID int) (domain.User, error) {
	user, err := u.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return domain.User{}, err
	}

	return user, nil
}

// возвращает юзера по переданной почте
func (u *userUseCase) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	user, err := u.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}

	return user, nil
}

// проверяет существует ли юзер
func (u *userUseCase) Check(ctx context.Context, userID int) (bool, error) {
	isExists, err := u.userRepo.CheckUserByID(ctx, userID)
	if err != nil {
		return false, err
	}

	return isExists, nil
}

// обновляет поля юзера
func (u *userUseCase) UpdateProfile(ctx context.Context, userID int, name, email *string) error {
	return u.userRepo.UpdateProfile(ctx, userID, name, email)
}
