package memory

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
)

func TestCreateUser(t *testing.T) {
	tests := []struct {
		name        string
		prepare     func(r *userRepo)
		input       domain.User
		expectedErr error
	}{
		{
			name:        "Успешное создание первого пользователя",
			prepare:     func(r *userRepo) {},
			input:       domain.User{Email: "test@mail.ru", Name: "Ivan", Phone: "+74951239898", PasswordHash: "hash"},
			expectedErr: nil,
		},
		{
			name: "Ошибка: создание существующего пользователя",
			prepare: func(r *userRepo) {
				id := uuid.New()
				r.users[id] = domain.User{ID: id, Email: "exists@mail.ru", Name: "Ivan"}
			},
			input:       domain.User{Email: "exists@mail.ru", Name: "Ivan"},
			expectedErr: domain.ErrEmailAlreadyExists,
		},
		{
			name: "Ошибка: регистронезависимость email",
			prepare: func(r *userRepo) {
				id := uuid.New()
				r.users[id] = domain.User{ID: id, Email: "exists@mail.ru"}
			},
			input:       domain.User{Email: "EXISTS@mail.ru"},
			expectedErr: domain.ErrEmailAlreadyExists,
		},
		{
			name:        "Ошибка: спецсимволы в почте",
			prepare:     func(r *userRepo) {},
			input:       domain.User{Email: "()<>[]:;\\.,@mail.ru"},
			expectedErr: domain.ErrInvalidEmail,
		},
		{
			name:        "Ошибка: две точки подряд",
			prepare:     func(r *userRepo) {},
			input:       domain.User{Email: "ma..il@mail.ru"},
			expectedErr: domain.ErrInvalidEmail,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Создаем чистый репозиторий
			repo := NewUserRepo().(*userRepo)
			tc.prepare(repo)

			ctx := context.Background()
			id, err := repo.CreateUser(ctx, tc.input)

			if tc.expectedErr != nil {
				require.ErrorIs(t, err, tc.expectedErr)
				require.Equal(t, uuid.Nil, id) // При ошибке ID должен быть нулевым
			} else {
				require.NoError(t, err)
				require.NotEqual(t, uuid.Nil, id) // Проверяем, что UUID сгенерирован

				// Проверяем, что пользователь появился в мапе по этому ID
				repo.mu.RLock()
				userInMap, exists := repo.users[id]
				repo.mu.RUnlock()

				require.True(t, exists, "User should exist in map by returned ID")
				require.Equal(t, tc.input.Name, userInMap.Name)
				// Email проверяем через ToLower, так как репозиторий должен его нормализовать
				require.Equal(t, strings.ToLower(tc.input.Email), strings.ToLower(userInMap.Email))
			}
		})
	}
}

func TestCreateUser_concurrency(t *testing.T) {
	repo := NewUserRepo().(*userRepo)
	const numGoroutines = 100

	g := new(errgroup.Group)

	for i := 0; i < numGoroutines; i++ {
		workerID := i
		g.Go(func() error {
			user := domain.User{
				Email: fmt.Sprintf("user_%d@mail.ru", workerID),
				Name:  "Ivan",
			}
			_, err := repo.CreateUser(context.Background(), user)
			return err
		})
	}

	// Wait возвращает ошибку, если хоть одна горутина вернула ошибку
	err := g.Wait()
	require.NoError(t, err)

	// Проверяем итоговое состояние
	require.Equal(t, numGoroutines, len(repo.users))
}
