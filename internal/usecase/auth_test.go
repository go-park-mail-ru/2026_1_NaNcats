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

func TestAuthUseCase_Regirster_Validation(t *testing.T) {
	// Тип для функции, которая настраивает моки под конкретный кейс
	type mockBehavior func(r *repoMocks.MockUserRepository, s *ucMocks.MockSessionUseCase)

	tests := []struct {
		name      string
		input     domain.User
		prepare   mockBehavior // здесь настраиваем моки
		expectErr error
	}{
		{
			name: "Успех: допускается использование одинарной точки в названии почты",
			input: domain.User{
				Name:         "Ivan",
				Email:        "m.a.i.l@mail.ru",
				PasswordHash: "valid_password_123",
			},
			prepare: func(r *repoMocks.MockUserRepository, s *ucMocks.MockSessionUseCase) {
				// Для успешной почты мы ждем вызовов базы и сессий
				r.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(1, nil)
				s.EXPECT().Create(gomock.Any(), 1).Return(domain.Session{ID: "ok"}, nil)
			},
			expectErr: nil,
		},
		{
			name: "Ошибка: спецсимволы в почте",
			input: domain.User{
				Name:         "Ivan",
				Email:        "()<>[]:;\\.,@mail.ru",
				PasswordHash: "valid_password_123",
			},
			expectErr: domain.ErrInvalidEmail,
		},
		{
			name: "Ошибка: точка в начале и конце",
			input: domain.User{
				Name:         "Ivan",
				Email:        ".mail.@mail.ru.",
				PasswordHash: "valid_password_123",
			},
			expectErr: domain.ErrInvalidEmail,
		},
		{
			name: "Ошибка: две точки подряд",
			input: domain.User{
				Name:         "Ivan",
				Email:        "ma..il@mail.ru",
				PasswordHash: "valid_password_123",
			},
			expectErr: domain.ErrInvalidEmail,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUserRepo := repoMocks.NewMockUserRepository(ctrl)
			mockSessionUC := ucMocks.NewMockSessionUseCase(ctrl)

			authUseCase := NewAuthUseCase(mockUserRepo, mockSessionUC)

			// Вызываем настройку моков из таблицы, если были указаны
			if testCase.prepare != nil {
				testCase.prepare(mockUserRepo, mockSessionUC)
			}

			_, _, err := authUseCase.Register(context.Background(), testCase.input)

			assert.ErrorIs(t, err, testCase.expectErr)
		})
	}
}

func TestAuthUseCase_Register(t *testing.T) {
	// Тип для функции, которая настраивает моки под конкретный кейс
	type mockBehavior func(r *repoMocks.MockUserRepository, s *ucMocks.MockSessionUseCase, user domain.User)

	tests := []struct {
		name      string
		input     domain.User
		prepare   mockBehavior // настройки для моков
		expectErr error
	}{
		{
			name: "Успешная регистрация",
			input: domain.User{
				Name:         "Ivan",
				Email:        "valid@mail.ru",
				PasswordHash: "valid_password_123",
			},
			prepare: func(r *repoMocks.MockUserRepository, s *ucMocks.MockSessionUseCase, user domain.User) {
				// Ожидаем создание юзера, возвращаем ID 1
				r.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(1, nil)
				// Ожидаем создание сессии для юзера 1
				s.EXPECT().Create(gomock.Any(), 1).Return(domain.Session{ID: "session_id"}, nil)
			},
			expectErr: nil,
		},
		{
			name: "Ошибка: пользователь уже существует",
			input: domain.User{
				Email:        "exists@mail.ru",
				PasswordHash: "valid_password_123",
			},
			prepare: func(r *repoMocks.MockUserRepository, s *ucMocks.MockSessionUseCase, user domain.User) {
				r.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(0, domain.ErrUserAlreadyExists)
				// Сессия не должна создаваться, если юзер не создан
			},
			expectErr: domain.ErrUserAlreadyExists,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			r := repoMocks.NewMockUserRepository(ctrl)
			s := ucMocks.NewMockSessionUseCase(ctrl)
			uc := NewAuthUseCase(r, s)

			if testCase.prepare != nil {
				testCase.prepare(r, s, testCase.input)
			}

			_, _, err := uc.Register(context.Background(), testCase.input)
			assert.ErrorIs(t, err, testCase.expectErr)
		})
	}
}
