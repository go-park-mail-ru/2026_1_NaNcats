package memory

import (
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
func (r *userRepo) CreateUser(user domain.User) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// проверяем на существование email'а
	if _, exists := r.users[user.Email]; exists {
		return 0, domain.ErrUserAlreadyExists
	}

	user.ID = r.nextID
	r.users[user.Email] = user
	r.nextID++

	return user.ID, nil
}

// метод нахождения пользователя по email'у
func (r *userRepo) GetUserByEmail(email string) (domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// проверяем на существование email'а
	user, exists := r.users[email]
	if !exists {
		return domain.User{}, domain.ErrUserNotFound
	}

	return user, nil
}
