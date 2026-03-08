package repository

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	"github.com/google/uuid"
)

// контракт репозитория пользователей
//
//go:generate mockgen -destination=mocks/user_mock.go -package=mocks github.com/go-park-mail-ru/2026_1_NaNcats/internal/repository UserRepository
type UserRepository interface {
	// метод создания юзера в репозитории, возвращает userID
	CreateUser(ctx context.Context, user domain.User) (uuid.UUID, error)
	// метод нахождения пользователя по email'у, возвращает юзера
	GetUserByEmail(ctx context.Context, email string) (domain.User, error)
	// метод нахождения пользователей по id, возвращает юзера
	GetUserByID(ctx context.Context, id uuid.UUID) (domain.User, error)
}
