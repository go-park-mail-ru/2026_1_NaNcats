package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/auth/internal/domain"
	repoMocks "github.com/go-park-mail-ru/2026_1_NaNcats/services/auth/internal/repository/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestSessionUseCase_Create(t *testing.T) {
	type mockInit func(m *repoMocks.MockSessionRepository, ttl time.Duration)

	userID := int64(42)
	role := "user"
	userAgent := "Mozilla/5.0"
	ttl := 24 * time.Hour

	tests := []struct {
		name        string
		mockInit    mockInit
		expectedErr bool
	}{
		{
			name: "Успешное создание сессии",
			mockInit: func(m *repoMocks.MockSessionRepository, ttl time.Duration) {
				m.EXPECT().Create(gomock.Any(), gomock.Any(), ttl).Return(nil)
			},
			expectedErr: false,
		},
		{
			name: "Ошибка репозитория при создании",
			mockInit: func(m *repoMocks.MockSessionRepository, ttl time.Duration) {
				m.EXPECT().Create(gomock.Any(), gomock.Any(), ttl).Return(errors.New("db error"))
			},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := repoMocks.NewMockSessionRepository(ctrl)
			tt.mockInit(mockRepo, ttl)

			uc := NewSessionUseCase(mockRepo, ttl)
			session, err := uc.Create(context.Background(), userID, role, userAgent)

			if tt.expectedErr {
				assert.Error(t, err)
				assert.Equal(t, uuid.Nil, session.ID)
			} else {
				assert.NoError(t, err)
				assert.NotEqual(t, uuid.Nil, session.ID)
				assert.Equal(t, userID, session.UserID)
				assert.Equal(t, role, session.Role)
				assert.Equal(t, userAgent, session.UserAgent)
				assert.True(t, session.ExpiresAt.After(time.Now()))
			}
		})
	}
}

func TestSessionUseCase_Check(t *testing.T) {
	type mockInit func(m *repoMocks.MockSessionRepository, id uuid.UUID)

	sessionID := uuid.New()

	tests := []struct {
		name          string
		id            uuid.UUID
		mockInit      mockInit
		expectedErrIs error
	}{
		{
			name: "Успешная проверка валидной сессии",
			id:   sessionID,
			mockInit: func(m *repoMocks.MockSessionRepository, id uuid.UUID) {
				m.EXPECT().GetByID(gomock.Any(), id).Return(domain.Session{
					ID:        id,
					UserID:    1,
					ExpiresAt: time.Now().Add(time.Hour),
				}, nil)
			},
			expectedErrIs: nil,
		},
		{
			name: "Ошибка: сессия истекла",
			id:   sessionID,
			mockInit: func(m *repoMocks.MockSessionRepository, id uuid.UUID) {
				m.EXPECT().GetByID(gomock.Any(), id).Return(domain.Session{
					ID:        id,
					UserID:    1,
					ExpiresAt: time.Now().Add(-time.Hour),
				}, nil)
			},
			expectedErrIs: domain.ErrSessionExpired,
		},
		{
			name: "Ошибка: сессия не найдена в БД",
			id:   sessionID,
			mockInit: func(m *repoMocks.MockSessionRepository, id uuid.UUID) {
				m.EXPECT().GetByID(gomock.Any(), id).Return(domain.Session{}, domain.ErrSessionNotFound)
			},
			expectedErrIs: domain.ErrSessionNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := repoMocks.NewMockSessionRepository(ctrl)
			tt.mockInit(mockRepo, tt.id)

			uc := NewSessionUseCase(mockRepo, time.Hour)
			session, err := uc.Check(context.Background(), tt.id)

			if tt.expectedErrIs != nil {
				assert.ErrorIs(t, err, tt.expectedErrIs)
				assert.Empty(t, session.UserID)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.id, session.ID)
			}
		})
	}
}

func TestSessionUseCase_Destroy(t *testing.T) {
	type mockInit func(m *repoMocks.MockSessionRepository, id uuid.UUID)

	sessionID := uuid.New()

	tests := []struct {
		name        string
		id          uuid.UUID
		mockInit    mockInit
		expectedErr bool
	}{
		{
			name: "Успешное удаление",
			id:   sessionID,
			mockInit: func(m *repoMocks.MockSessionRepository, id uuid.UUID) {
				m.EXPECT().Delete(gomock.Any(), id).Return(nil)
			},
			expectedErr: false,
		},
		{
			name: "Ошибка при удалении",
			id:   sessionID,
			mockInit: func(m *repoMocks.MockSessionRepository, id uuid.UUID) {
				m.EXPECT().Delete(gomock.Any(), id).Return(errors.New("db fail"))
			},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := repoMocks.NewMockSessionRepository(ctrl)
			tt.mockInit(mockRepo, tt.id)

			uc := NewSessionUseCase(mockRepo, time.Hour)
			err := uc.Destroy(context.Background(), tt.id)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSessionUseCase_SetCSRF(t *testing.T) {
	type mockInit func(m *repoMocks.MockSessionRepository, id uuid.UUID)

	sessionID := uuid.New()

	tests := []struct {
		name        string
		id          uuid.UUID
		mockInit    mockInit
		expectedErr bool
	}{
		{
			name: "Успешная генерация и установка CSRF",
			id:   sessionID,
			mockInit: func(m *repoMocks.MockSessionRepository, id uuid.UUID) {
				m.EXPECT().SetCSRF(gomock.Any(), id, gomock.Any()).Return(nil)
			},
			expectedErr: false,
		},
		{
			name: "Ошибка при сохранении CSRF в репозиторий",
			id:   sessionID,
			mockInit: func(m *repoMocks.MockSessionRepository, id uuid.UUID) {
				m.EXPECT().SetCSRF(gomock.Any(), id, gomock.Any()).Return(errors.New("redis fail"))
			},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := repoMocks.NewMockSessionRepository(ctrl)
			tt.mockInit(mockRepo, tt.id)

			uc := NewSessionUseCase(mockRepo, time.Hour)
			token, err := uc.SetCSRF(context.Background(), tt.id)

			if tt.expectedErr {
				assert.Error(t, err)
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)
			}
		})
	}
}

func TestSessionUseCase_GetCSRF(t *testing.T) {
	type mockInit func(m *repoMocks.MockSessionRepository, id uuid.UUID)

	sessionID := uuid.New()

	tests := []struct {
		name          string
		id            uuid.UUID
		mockInit      mockInit
		expectedToken string
		expectedErr   bool
	}{
		{
			name: "Успешное получение существующего CSRF",
			id:   sessionID,
			mockInit: func(m *repoMocks.MockSessionRepository, id uuid.UUID) {
				m.EXPECT().GetCSRF(gomock.Any(), id).Return("existing-token-123", nil)
			},
			expectedToken: "existing-token-123",
			expectedErr:   false,
		},
		{
			name: "CSRF не найден (пустая строка), генерация нового",
			id:   sessionID,
			mockInit: func(m *repoMocks.MockSessionRepository, id uuid.UUID) {
				m.EXPECT().GetCSRF(gomock.Any(), id).Return("", nil)
				m.EXPECT().SetCSRF(gomock.Any(), id, gomock.Any()).Return(nil)
			},
			expectedToken: "",
			expectedErr:   false,
		},
		{
			name: "Ошибка при получении CSRF",
			id:   sessionID,
			mockInit: func(m *repoMocks.MockSessionRepository, id uuid.UUID) {
				m.EXPECT().GetCSRF(gomock.Any(), id).Return("", errors.New("db error"))
			},
			expectedToken: "",
			expectedErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := repoMocks.NewMockSessionRepository(ctrl)
			tt.mockInit(mockRepo, tt.id)

			uc := NewSessionUseCase(mockRepo, time.Hour)
			token, err := uc.GetCSRF(context.Background(), tt.id)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.expectedToken != "" {
					assert.Equal(t, tt.expectedToken, token)
				} else {
					assert.NotEmpty(t, token)
				}
			}
		})
	}
}
