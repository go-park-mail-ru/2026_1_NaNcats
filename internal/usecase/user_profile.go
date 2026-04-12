package usecase

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
)

//go:generate mockgen -destination=mocks/user_profile_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/internal/usecase UserProfileUseCase
type UserProfileUseCase interface {
	GetUserProfile(ctx context.Context, userID int) (domain.User, error)
}

type userProfileUseCase struct {
	userUC UserUseCase
}

func NewUserProfileUseCase(uuc UserUseCase) UserProfileUseCase {
	return &userProfileUseCase{
		userUC: uuc,
	}
}

func (up *userProfileUseCase) GetUserProfile(ctx context.Context, userID int) (domain.User, error) {
	// этот метод будет масштабироваться, сюда ещё добавлю Get платежных методов, локаций и элементов геймификации
	user, err := up.userUC.GetByID(ctx, userID)
	if err != nil {
		return domain.User{}, err
	}

	return user, nil
}
