package repository

import "github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"

type UserRepository interface {
	CreateUser(user domain.User) (int, error)
	GetUserByEmail(email string) (domain.User, error)
}
