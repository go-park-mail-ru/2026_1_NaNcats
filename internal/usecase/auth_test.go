package usecase

import (
	"context"
	"testing"

	"github.com/go-park-mail-ru/2026_1_NaNcats/internal/domain"
	repoMocks "github.com/go-park-mail-ru/2026_1_NaNcats/internal/repository/mocks"
	ucMocks "github.com/go-park-mail-ru/2026_1_NaNcats/internal/usecase/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestAuthUseCase_Register(t *testing.T) {
	// Группируем моки в структуру для удобной передачи
	type mocks struct {
		userRepo  *repoMocks.MockUserRepository
		sessionUC *ucMocks.MockSessionUseCase
	}

	// Тип для инициализации моков (mockInit)
	type mockInit func(m mocks, input domain.User, resID int, userAgent string)

	mockUserID := 1
	mockSessionID := uuid.MustParse("99999999-9999-9999-9999-999999999999")

	tests := []struct {
		name      string
		input     domain.User
		mockInit  mockInit
		expectErr error
	}{
		{
			name: "Успешная регистрация",
			input: domain.User{
				Name:         "Ivan",
				Email:        "valid@mail.ru",
				PasswordHash: "valid_password_123",
			},
			mockInit: func(m mocks, input domain.User, resID int, userAgent string) {
				m.userRepo.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Return(resID, nil)
				m.sessionUC.EXPECT().
					Create(gomock.Any(), resID, userAgent).
					Return(domain.Session{ID: mockSessionID}, nil)
			},
			expectErr: nil,
		},
		{
			name: "Успех: допускается точка в названии почты",
			input: domain.User{
				Email:        "m.a.i.l@mail.ru",
				PasswordHash: "password123",
			},
			mockInit: func(m mocks, input domain.User, resID int, userAgent string) {
				m.userRepo.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Return(resID, nil)
				m.sessionUC.EXPECT().
					Create(gomock.Any(), resID, userAgent).
					Return(domain.Session{ID: mockSessionID}, nil)
			},
			expectErr: nil,
		},
		{
			name: "Ошибка: пользователь уже существует",
			input: domain.User{
				Email:        "exists@mail.ru",
				PasswordHash: "password123",
			},
			mockInit: func(m mocks, input domain.User, resID int, userAgent string) {
				m.userRepo.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Return(0, domain.ErrEmailAlreadyExists)
			},
			expectErr: domain.ErrEmailAlreadyExists,
		},
		{
			name: "Ошибка: спецсимволы в почте",
			input: domain.User{
				Email:        "()<>[]:;\\.,@mail.ru",
				PasswordHash: "password123",
			},
			mockInit:  nil, // Моки не должны вызываться
			expectErr: domain.ErrInvalidEmail,
		},
		{
			name: "Ошибка: две точки подряд",
			input: domain.User{
				Email:        "ma..il@mail.ru",
				PasswordHash: "password123",
			},
			mockInit:  nil, // Моки не должны вызываться
			expectErr: domain.ErrInvalidEmail,
		},
		{
			name: "Ошибка: эмодзи в почте",
			input: domain.User{
				Email:        "😂😂😂😂😂😂😂@mail.ru",
				PasswordHash: "password123",
			},
			mockInit:  nil,
			expectErr: domain.ErrInvalidEmail,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Инициализация контроллера и моков
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			m := mocks{
				userRepo:  repoMocks.NewMockUserRepository(ctrl),
				sessionUC: ucMocks.NewMockSessionUseCase(ctrl),
			}

			authUseCase := NewAuthUseCase(m.userRepo, m.sessionUC)

			userAgent := "Mozilla/5.0 (Test Agent)"

			// Если mockInit задан, настраиваем поведение моков
			if testCase.mockInit != nil {
				testCase.mockInit(m, testCase.input, mockUserID, userAgent)
			}

			// Выполнение тестируемого метода
			user, session, err := authUseCase.Register(context.Background(), testCase.input, userAgent)

			// Проверки результата
			if testCase.expectErr != nil {
				assert.ErrorIs(t, err, testCase.expectErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, mockUserID, user.ID)
				assert.Equal(t, mockSessionID, session.ID)
			}
		})
	}
}
