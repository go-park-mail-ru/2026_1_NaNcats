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
	// Создаем структуру, чтобы не передавать много аргументов в mockInit
	type mocks struct {
		userRepo  *repoMocks.MockUserRepository
		sessionUC *ucMocks.MockSessionUseCase
	}

	type mockInit func(m mocks, input domain.User, resID uuid.UUID)

	// Заранее создаем UUID для тестов
	mockUserID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
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
			mockInit: func(m mocks, input domain.User, resID uuid.UUID) {
				m.userRepo.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Return(resID, nil)
				m.sessionUC.EXPECT().
					Create(gomock.Any(), resID).
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
			mockInit: func(m mocks, input domain.User, resID uuid.UUID) {
				m.userRepo.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Return(uuid.Nil, domain.ErrEmailAlreadyExists)
			},
			expectErr: domain.ErrEmailAlreadyExists,
		},
		{
			name:      "Ошибка: спецсимволы в почте",
			input:     domain.User{Email: "()<>[]:;\\.,@mail.ru"},
			mockInit:  nil, // Моки не вызываются, упадет на валидации
			expectErr: domain.ErrInvalidEmail,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// Группируем моки
			m := mocks{
				userRepo:  repoMocks.NewMockUserRepository(ctrl),
				sessionUC: ucMocks.NewMockSessionUseCase(ctrl),
			}

			authUseCase := NewAuthUseCase(m.userRepo, m.sessionUC)

			// Настройка моков через структуру
			if testCase.mockInit != nil {
				testCase.mockInit(m, testCase.input, mockUserID)
			}

			user, session, err := authUseCase.Register(context.Background(), testCase.input)

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
