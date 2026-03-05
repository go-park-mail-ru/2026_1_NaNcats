package repository

import "github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"

// контракт репозитория пользователей
type UserRepository interface {
	// метод создания юзера в репозитории
	CreateUser(user domain.User) (int, error)
	// метод нахождения пользователя по email'у
	GetUserByEmail(email string) (domain.User, error)
	// метод нахождения пользователей по id
	GetUserByID(id int) (domain.User, error)
}
