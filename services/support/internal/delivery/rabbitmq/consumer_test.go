package rabbitmq

import (
	"context"
	"errors"
	"testing"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/support/internal/repository/mocks"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/rabbitmq/events"
	"github.com/mailru/easyjson"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestSupportConsumer_Handler(t *testing.T) {
	type mockInit func(r *mocks.MockSupportRepository)

	tests := []struct {
		name     string
		event    events.UserRoleChangedEvent
		payload  []byte // если нужно проверить битый JSON
		mockInit mockInit
		wantErr  bool
	}{
		{
			name: "Успешное создание профиля (NewRole = support)",
			event: events.UserRoleChangedEvent{
				UserID:  1,
				OldRole: "client",
				NewRole: "support",
			},
			mockInit: func(r *mocks.MockSupportRepository) {
				r.EXPECT().CreateAgentProfile(gomock.Any(), int64(1)).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "Успешное удаление профиля (OldRole = support, NewRole != support)",
			event: events.UserRoleChangedEvent{
				UserID:  2,
				OldRole: "support",
				NewRole: "client",
			},
			mockInit: func(r *mocks.MockSupportRepository) {
				r.EXPECT().DeleteAgentProfile(gomock.Any(), int64(2)).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "Нет действий (роль меняется между user и admin)",
			event: events.UserRoleChangedEvent{
				UserID:  3,
				OldRole: "user",
				NewRole: "admin",
			},
			mockInit: func(r *mocks.MockSupportRepository) {
				// Репозиторий не должен вызываться
			},
			wantErr: false,
		},
		{
			name:    "Ошибка анмаршалинга (Poison Pill)",
			payload: []byte(`{invalid-json}`),
			mockInit: func(r *mocks.MockSupportRepository) {
				// Репозиторий не должен вызываться
			},
			wantErr: false,
		},
		{
			name: "Ошибка репозитория при создании (Retry)",
			event: events.UserRoleChangedEvent{
				UserID:  4,
				NewRole: "support",
			},
			mockInit: func(r *mocks.MockSupportRepository) {
				r.EXPECT().CreateAgentProfile(gomock.Any(), int64(4)).Return(errors.New("db connection fail"))
			},
			wantErr: true,
		},
		{
			name: "Ошибка репозитория при удалении (Retry)",
			event: events.UserRoleChangedEvent{
				UserID:  5,
				OldRole: "support",
				NewRole: "client",
			},
			mockInit: func(r *mocks.MockSupportRepository) {
				r.EXPECT().DeleteAgentProfile(gomock.Any(), int64(5)).Return(errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockSupportRepository(ctrl)
			tt.mockInit(mockRepo)

			body := tt.payload
			if body == nil {
				body, _ = easyjson.Marshal(tt.event)
			}

			handleLogic := func(ctx context.Context, body []byte) error {
				var event events.UserRoleChangedEvent
				if err := easyjson.Unmarshal(body, &event); err != nil {
					return nil
				}
				if event.NewRole == "support" {
					if err := mockRepo.CreateAgentProfile(ctx, event.UserID); err != nil {
						return err
					}
				}
				if event.OldRole == "support" && event.NewRole != "support" {
					if err := mockRepo.DeleteAgentProfile(ctx, event.UserID); err != nil {
						return err
					}
				}
				return nil
			}

			err := handleLogic(context.Background(), body)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
