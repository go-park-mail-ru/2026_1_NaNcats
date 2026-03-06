package repository

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
)

// контракт репозитория пользователей
type UserRepository interface {
	// метод создания юзера в репозитории
	CreateUser(ctx context.Context, user domain.User) (int, error)
	// метод нахождения пользователя по email'у
	GetUserByEmail(ctx context.Context, email string) (domain.User, error)
	// метод нахождения пользователей по id
	GetUserByID(ctx context.Context, id int) (domain.User, error)
}
