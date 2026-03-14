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
	// Тип для настройки поведения моков
	type mockBehavior func(r *repoMocks.MockUserRepository, s *ucMocks.MockSessionUseCase, input domain.User, resID uuid.UUID)

	// Заранее создаем UUID для тестов
	mockUserID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	mockSessionID := uuid.MustParse("99999999-9999-9999-9999-999999999999")

	tests := []struct {
		name      string
		input     domain.User
		prepare   mockBehavior
		expectErr error
	}{
		{
			name: "Успешная регистрация",
			input: domain.User{
				Name:         "Ivan",
				Email:        "valid@mail.ru",
				PasswordHash: "valid_password_123",
			},
			prepare: func(r *repoMocks.MockUserRepository, s *ucMocks.MockSessionUseCase, input domain.User, resID uuid.UUID) {
				r.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Return(resID, nil)
				s.EXPECT().
					Create(gomock.Any(), resID).
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
			prepare: func(r *repoMocks.MockUserRepository, s *ucMocks.MockSessionUseCase, input domain.User, resID uuid.UUID) {
				r.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(resID, nil)
				s.EXPECT().Create(gomock.Any(), resID).Return(domain.Session{ID: mockSessionID}, nil)
			},
			expectErr: nil,
		},
		{
			name: "Ошибка: пользователь уже существует",
			input: domain.User{
				Email:        "exists@mail.ru",
				PasswordHash: "password123",
			},
			prepare: func(r *repoMocks.MockUserRepository, s *ucMocks.MockSessionUseCase, input domain.User, resID uuid.UUID) {
				r.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Return(uuid.Nil, domain.ErrEmailAlreadyExists)
			},
			expectErr: domain.ErrEmailAlreadyExists,
		},
		{
			name: "Ошибка: спецсимволы в почте",
			input: domain.User{
				Email:        "()<>[]:;\\.,@mail.ru",
				PasswordHash: "password123",
			},
			prepare:   nil, // Моки не вызываются, упадет на валидации
			expectErr: domain.ErrInvalidEmail,
		},
		{
			name: "Ошибка: две точки подряд",
			input: domain.User{
				Email:        "ma..il@mail.ru",
				PasswordHash: "password123",
			},
			prepare:   nil,
			expectErr: domain.ErrInvalidEmail,
		},
		{
			name: "Ошибка: эмодзи в почте",
			input: domain.User{
				Email:        "😂😂😂😂😂😂😂@mail.ru",
				PasswordHash: "password123",
			},
			prepare:   nil,
			expectErr: domain.ErrInvalidEmail,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Инициализация
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUserRepo := repoMocks.NewMockUserRepository(ctrl)
			mockSessionUC := ucMocks.NewMockSessionUseCase(ctrl)
			authUseCase := NewAuthUseCase(mockUserRepo, mockSessionUC)

			// Настройка моков
			if testCase.prepare != nil {
				testCase.prepare(mockUserRepo, mockSessionUC, testCase.input, mockUserID)
			}

			// Выполнение
			user, session, err := authUseCase.Register(context.Background(), testCase.input)

			// Проверки
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
