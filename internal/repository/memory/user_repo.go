package memory

import (
	"sync"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/repository"
)

type userRepo struct {
	mu     sync.RWMutex           // защита от одновременного чтения из мапы
	users  map[string]domain.User // мапа юзеров, ключ - email
	nextID int                    // счетчик для автоинкремента id
}

func newUserRepo() repository.UserRepository {
	return &userRepo{
		users:  make(map[string]domain.User),
		nextID: 1,
	}
}

func (r *userRepo) CreateUser(user domain.User) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.users[user.Email]; exists {
		return 0, domain.ErrUserAlreadyExists
	}

	user.ID = r.nextID
	r.users[user.Email] = user
	r.nextID++

	return user.ID, nil
}

func (r *userRepo) GetUserByEmail(email string) (domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.users[email]
	if !exists {
		return domain.User{}, domain.ErrUserNotFound
	}

	return user, nil
}
