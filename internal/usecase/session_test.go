package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	repoMocks "github.com/go-park-mail-ru/2026_1_NaNcats/internal/repository/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestSessionUseCase_Create(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repoMocks.NewMockSessionRepository(ctrl)
	ttl := 24 * time.Hour
	uc := NewSessionUseCase(mockRepo, ttl)

	t.Run("Успешное создание сессии", func(t *testing.T) {
		ctx := context.Background()

		// Ожидаем вызов Create в репозитории.
		// Используем gomock.Any(), так как ID генерируется внутри метода Create.
		mockRepo.EXPECT().
			Create(gomock.Any(), gomock.Any()).
			Return(nil)

		sess, err := uc.Create(ctx, 1)

		assert.NoError(t, err)
		assert.Equal(t, 1, sess.UserID)
		assert.NotEqual(t, uuid.Nil, sess.ID)
		// Проверяем, что время истечения примерно соответствует Now + TTL
		assert.WithinDuration(t, time.Now().Add(ttl), sess.ExpiresAt, time.Second)
	})
}

func TestSessionUseCase_Check(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repoMocks.NewMockSessionRepository(ctrl)
	uc := NewSessionUseCase(mockRepo, 24*time.Hour)
	ctx := context.Background()

	t.Run("Успешная проверка валидной сессии", func(t *testing.T) {
		sessID := uuid.New()
		userID := 1

		mockRepo.EXPECT().
			GetByID(gomock.Any(), sessID).
			Return(domain.Session{
				ID:        sessID,
				UserID:    1,
				ExpiresAt: time.Now().Add(time.Hour), // валидна еще час
			}, nil)

		resUserID, err := uc.Check(ctx, sessID)

		assert.NoError(t, err)
		assert.Equal(t, userID, resUserID)
	})

	t.Run("Ошибка: сессия истекла", func(t *testing.T) {
		sessID := uuid.New()

		mockRepo.EXPECT().
			GetByID(gomock.Any(), sessID).
			Return(domain.Session{
				ID:        sessID,
				ExpiresAt: time.Now().Add(-time.Hour), // истекла час назад
			}, nil)

		// Ждем, что UseCase сам инициирует удаление протухшей сессии
		mockRepo.EXPECT().
			Delete(gomock.Any(), sessID).
			Return(nil)

		uid, err := uc.Check(ctx, sessID)

		assert.ErrorIs(t, err, domain.ErrSessionExpired)
		assert.Equal(t, 0, uid)
	})

	t.Run("Ошибка: сессия не найдена в репо", func(t *testing.T) {
		sessID := uuid.New()
		mockRepo.EXPECT().
			GetByID(gomock.Any(), sessID).
			Return(domain.Session{}, domain.ErrSessionNotFound)

		uid, err := uc.Check(ctx, sessID)

		assert.ErrorIs(t, err, domain.ErrSessionNotFound)
		assert.Equal(t, 0, uid)
	})
}

func TestSessionUseCase_Destroy(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repoMocks.NewMockSessionRepository(ctrl)
	uc := NewSessionUseCase(mockRepo, 24*time.Hour)

	t.Run("Успешное удаление", func(t *testing.T) {
		sessID := uuid.New()
		mockRepo.EXPECT().
			Delete(gomock.Any(), sessID).
			Return(nil)

		err := uc.Destroy(context.Background(), sessID)
		assert.NoError(t, err)
	})
}
