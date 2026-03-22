package memory

import (
	"context"
	"strings"
	"sync"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/repository"
)

// реализация контракта репозитория юзера на мапах (in-memory)
type userRepo struct {
	mu     sync.RWMutex        // защита от записи во время чтения из мапы
	users  map[int]domain.User // мапа юзеров, ключ - id
	nextID int
}

// функция-конструктор userRepo
func NewUserRepo() repository.UserRepository {
	return &userRepo{
		users:  make(map[int]domain.User),
		nextID: 1,
	}
}

// метод создания юзера в репозитории
func (r *userRepo) CreateUser(ctx context.Context, user domain.User) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	user.Email = strings.ToLower(strings.TrimSpace(user.Email))

	// проверяем на существование email'а
	for _, curUser := range r.users {
		if curUser.Email == user.Email {
			return 0, domain.ErrEmailAlreadyExists
		}
	}

	user.ID = r.nextID
	r.users[user.ID] = user
	r.nextID++

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

func (r *userRepo) GetUserByID(ctx context.Context, id int) (domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.users[id]
	if !exists {
		return domain.User{}, domain.ErrUserNotFound
	}

	return user, nil
}
