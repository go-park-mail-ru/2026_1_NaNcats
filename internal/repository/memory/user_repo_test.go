package memory

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	"github.com/stretchr/testify/require"
)

// Тестирует метод структуры userRepo, который создает юзера
func TestCreateUser(t *testing.T) {
	tests := []struct {
		name        string            // Название теста
		prepare     func(r *userRepo) // Состояние нашей БД перед запуском метода
		input       domain.User       // Входящие данные о потенциальном новом юзере
		expectedID  int               // Какой ID мы ожидаем получить для нового юзера
		expectedErr error             // Какую ошибку мы ожидаем получить при добавлении нового юзера в БД
	}{
		{
			name:        "Успешное создание первого пользователя",
			prepare:     func(r *userRepo) {}, // пустая мапа
			input:       domain.User{Email: "test@mail.ru", Name: "Ivan", Phone: "+74951239898", PasswordHash: "thisIsPasswordHash"},
			expectedID:  1,
			expectedErr: nil,
		},
		{
			name: "Успешное создание второго пользователя",
			prepare: func(r *userRepo) {
				// Создаем пользователя в БД заранее
				r.users["test@mail.ru"] = domain.User{Email: "test@mail.ru", ID: 1}
				r.nextID = 2
			},
			input:       domain.User{Email: "new_test@mail.ru", Name: "Ivan"},
			expectedID:  2,
			expectedErr: nil,
		},
		{
			name: "Ошибка: создание существующего пользователя",
			prepare: func(r *userRepo) {
				r.users["exists@mail.ru"] = domain.User{Email: "exists@mail.ru", ID: 1}
				r.nextID = 2
			},
			input:       domain.User{Email: "exists@mail.ru", Name: "Ivan"},
			expectedID:  0,
			expectedErr: domain.ErrUserAlreadyExists,
		},
		{
			name: "Ошибка: создание существующего пользователя (проверка на регистронезависимость @)",
			prepare: func(r *userRepo) {
				r.users["exists@mail.ru"] = domain.User{Email: "exists@mail.ru", ID: 1}
				r.nextID = 2
			},
			input:       domain.User{Email: "EXISTS@mail.ru", Name: "Ivan"},
			expectedID:  0,
			expectedErr: domain.ErrUserAlreadyExists,
		},
		{
			name: "Ошибка: создание существующего пользователя (проверка на регистронезависимость после @)",
			prepare: func(r *userRepo) {
				r.users["exists@mail.ru"] = domain.User{Email: "exists@mail.ru", ID: 1}
				r.nextID = 2
			},
			input:       domain.User{Email: "exists@MAIL.ru", Name: "Ivan"},
			expectedID:  0,
			expectedErr: domain.ErrUserAlreadyExists,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Создаем чистую БД
			// приводим интерфейс repository.UserRepository к конкретному типу *userRepo для доступа к приватным полям
			repo := NewUserRepo().(*userRepo)

			// Подготавливаем состояние
			tc.prepare(repo)

			ctx := context.Background()

			// Выполняем метод
			id, err := repo.CreateUser(ctx, tc.input)

			if tc.expectedErr != nil {
				require.ErrorIs(t, err, tc.expectedErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedID, id)

				mapKey := strings.ToLower(tc.input.Email)

				// Отдельно проверяем корректность записи некоторых полей пользователя
				require.Equal(t, tc.input.Email, repo.users[mapKey].Email)
				require.Equal(t, tc.input.Name, repo.users[mapKey].Name)
				require.Equal(t, tc.input.Phone, repo.users[mapKey].Phone)
				require.Equal(t, tc.input.PasswordHash, repo.users[mapKey].PasswordHash)
			}
		})
	}
}

// Тестируем конкурентный доступ в TestCreateUser
func TestCreateUser_concurrency(t *testing.T) {
	// Создаем общую чистую БД для реализации конкуретного доступа
	repo := NewUserRepo().(*userRepo)

	const numGoroutines = 100
	var wg sync.WaitGroup

	for i := range numGoroutines {
		wg.Add(1)

		go func(workerID int) {
			defer wg.Done()

			user := domain.User{
				Email: fmt.Sprintf("user_%d@mail.ru", workerID),
				Name:  "Ivan",
			}

			ctx := context.Background()

			_, err := repo.CreateUser(ctx, user)

			require.NoError(t, err)
		}(i)
	}

	wg.Wait()

	// Количество юзеров в БД должно быть равно количеству запущенных горутин
	require.Equal(t, numGoroutines, len(repo.users))
	// Счётчик должен верно инкрементироваться
	require.Equal(t, numGoroutines+1, repo.nextID)
}
