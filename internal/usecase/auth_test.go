package usecase

import (
	"context"
	"testing"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	repoMocks "github.com/go-park-mail-ru/2026_1_NaNcats/internal/repository/mocks"
	ucMocks "github.com/go-park-mail-ru/2026_1_NaNcats/internal/usecase/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestAuthUseCase_Register_Success(t *testing.T) {
	// Создаем котролеер моков
	ctrl := gomock.NewController(t)
	// Проверяет, все ли методы, которые мы ожидали, были вызваны
	defer ctrl.Finish()

	// Экземпляр мока репозитория юзера
	mockUserRepo := repoMocks.NewMockUserRepository(ctrl)

	// Мок юзкейса сессии
	mockSessionUC := ucMocks.NewMockSessionUseCase(ctrl)

	// Создаем настоящий usecase, подсовывая ему моки
	authUC := NewAuthUseCase(mockUserRepo, mockSessionUC)

	ctx := context.Background()

	// данные юзера, поступившие запросом
	requestedUserData := domain.User{
		Email:        "aboba@gmail.com",
		PasswordHash: "secret_password",
	}

	// expect на срабатывание CreateUser. Передать можно any, возвращаем userID 1 и err nil, повторяем 1 раз
	mockUserRepo.EXPECT().
		CreateUser(gomock.Any(), gomock.Any()).
		Return(1, nil).
		Times(1)

	// expect на срабатывание Create сессии. Передать можно any и только userID 1, возвращаем мок сессию с id 123456 и err nil
	mockSessionUC.EXPECT().
		Create(gomock.Any(), 1).
		Return(domain.Session{ID: "123456"}, nil).
		Times(1)

	createdUser, createdSession, err := authUC.Register(ctx, requestedUserData)

	assert.NoError(t, err)
	assert.Equal(t, 1, createdUser.ID)
	assert.Equal(t, "123456", createdSession.ID)
}
