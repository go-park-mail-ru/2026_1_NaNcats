package memory

import (
	"context"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
)

func TestSessionRepo_CRUD(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name string
		run  func(t *testing.T, repo repository.SessionRepository)
	}{
		{
			name: "Успешное создание и получение",
			run: func(t *testing.T, repo repository.SessionRepository) {
				// Подготовка данных
				sessID := uuid.New()
				userID := 1
				session := domain.Session{
					ID:        sessID,
					UserID:    userID,
					ExpiresAt: time.Now().Add(time.Hour),
				}

				// Create
				err := repo.Create(ctx, session)
				require.NoError(t, err)

				// Get
				savedSess, err := repo.GetByID(ctx, sessID)
				require.NoError(t, err)
				require.Equal(t, userID, savedSess.UserID)
				require.Equal(t, sessID, savedSess.ID)
			},
		},
		{
			name: "Ошибка: сессия не найдена",
			run: func(t *testing.T, repo repository.SessionRepository) {
				_, err := repo.GetByID(ctx, uuid.New()) // Случайный ID
				require.ErrorIs(t, err, domain.ErrSessionNotFound)
			},
		},
		{
			name: "Успешное удаление",
			run: func(t *testing.T, repo repository.SessionRepository) {
				sessID := uuid.New()
				session := domain.Session{
					ID:        sessID,
					UserID:    5,
					ExpiresAt: time.Now().Add(time.Hour),
				}

				// Сначала создаем
				err := repo.Create(ctx, session)
				require.NoError(t, err)

				// Удаляем
				err = repo.Delete(ctx, sessID)
				require.NoError(t, err)

				// Проверяем, что её больше нет
				_, err = repo.GetByID(ctx, sessID)
				require.ErrorIs(t, err, domain.ErrSessionNotFound)
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			repo := NewSessionRepo()
			testCase.run(t, repo)
		})
	}
}

func TestSessionRepo_Concurrency(t *testing.T) {
	repo := NewSessionRepo()
	ctx := context.Background()
	g := new(errgroup.Group)

	const iterations = 100

	// Одновременно запускаем 100 записей
	for i := 0; i < iterations; i++ {
		g.Go(func() error {
			s := domain.Session{
				ID:     uuid.New(),
				UserID: i + 1,
			}
			return repo.Create(ctx, s)
		})
	}

	// Ждем завершения без ошибок
	err := g.Wait()
	require.NoError(t, err)

	// Проверяем, что в мапе ровно 100 записей
	// Приводим интерфейс к структуре, чтобы заглянуть в приватную мапу
	r := repo.(*sessionRepo)
	require.Equal(t, iterations, len(r.sessions))
}
