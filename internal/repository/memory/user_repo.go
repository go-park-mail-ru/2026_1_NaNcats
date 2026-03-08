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
	mu     sync.RWMutex           // защита от одновременного чтения из мапы
	users  map[string]domain.User // мапа юзеров, ключ - email
	nextID int                    // счетчик для автоинкремента id
}

// функция-конструктор userRepo
func NewUserRepo() repository.UserRepository {
	return &userRepo{
		users:  make(map[string]domain.User),
		nextID: 1,
	}
}

// метод создания юзера в репозитории
func (r *userRepo) CreateUser(ctx context.Context, user domain.User) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	emailLower := strings.ToLower(user.Email)

	// проверяем на существование email'а
	if _, exists := r.users[emailLower]; exists {
		return 0, domain.ErrUserAlreadyExists
	}

	user.ID = r.nextID
	r.users[user.Email] = user
	r.nextID++

	return user.ID, nil
}

// метод нахождения пользователя по email'у
func (r *userRepo) GetUserByEmail(ctx context.Context, email string) (domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// проверяем на существование email'а
	user, exists := r.users[strings.ToLower(email)]

	if !exists {
		return domain.User{}, domain.ErrUserNotFound
	}

	return user, nil
}

func (r *userRepo) GetUserByID(ctx context.Context, id int) (domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, user := range r.users {
		if user.ID == id {
			return user, nil
		}
	}

	return domain.User{}, domain.ErrUserNotFound
}
