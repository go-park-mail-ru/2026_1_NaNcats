package repository

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
)

// контракт репозитория пользователей
//
//go:generate mockgen -destination=mocks/user_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/internal/repository UserRepository
type UserRepository interface {
	// метод создания юзера в репозитории, возвращает userID
	CreateUser(ctx context.Context, user domain.User) (int, error)
	// метод нахождения пользователя по email'у, возвращает юзера
	GetUserByEmail(ctx context.Context, email string) (domain.User, error)
	// метод нахождения пользователей по id, возвращает юзера
	GetUserByID(ctx context.Context, id int) (domain.User, error)
	// метод для проверки существования юзера по ID
	CheckUserByID(ctx context.Context, userID int) (bool, error)
	// метод для обновления полей юзера
	UpdateProfile(ctx context.Context, userID int, name, email *string) error
}
