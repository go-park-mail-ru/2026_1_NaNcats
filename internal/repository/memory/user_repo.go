package memory

import (
	"context"
	"strings"
	"sync"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/repository"
	"github.com/google/uuid"
)

// реализация контракта репозитория юзера на мапах (in-memory)
type userRepo struct {
	mu    sync.RWMutex              // защита от одновременного чтения из мапы
	users map[uuid.UUID]domain.User // мапа юзеров, ключ - id
}

// функция-конструктор userRepo
func NewUserRepo() repository.UserRepository {
	return &userRepo{
		users: make(map[uuid.UUID]domain.User),
	}
}

// метод создания юзера в репозитории
func (r *userRepo) CreateUser(ctx context.Context, user domain.User) (uuid.UUID, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	user.Email = strings.ToLower(strings.TrimSpace(user.Email))

	// проверяем на существование email'а
	for _, curUser := range r.users {
		if curUser.Email == user.Email {
			return uuid.Nil, domain.ErrEmailAlreadyExists
		}
	}

	user.ID = uuid.New()
	r.users[user.ID] = user

	return user.ID, nil
}

// метод нахождения пользователя по email'у
func (r *userRepo) GetUserByEmail(ctx context.Context, email string) (domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	emailLower := strings.ToLower(strings.TrimSpace(email))

	// проверяем на существование email'а
	for _, user := range r.users {
		if user.Email == emailLower {
			return user, nil
		}
	}

	return domain.User{}, domain.ErrUserNotFound
}

func (r *userRepo) GetUserByID(ctx context.Context, id uuid.UUID) (domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.users[id]
	if !exists {
		return domain.User{}, domain.ErrUserNotFound
	}

	return user, nil
}
