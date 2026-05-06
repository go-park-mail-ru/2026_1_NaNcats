package common

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetUserID(t *testing.T) {
	tests := []struct {
		name       string
		setupCtx   func() context.Context
		expectedID int64
		expectedOk bool
	}{
		{
			name: "Успешное получение UserID",
			setupCtx: func() context.Context {
				return context.WithValue(context.Background(), UserIDKey, int64(123))
			},
			expectedID: 123,
			expectedOk: true,
		},
		{
			name: "UserID отсутствует в контексте",
			setupCtx: func() context.Context {
				return context.Background()
			},
			expectedID: 0,
			expectedOk: false,
		},
		{
			name: "UserID присутствует, но неверного типа (string вместо int64)",
			setupCtx: func() context.Context {
				return context.WithValue(context.Background(), UserIDKey, "123")
			},
			expectedID: 0,
			expectedOk: false,
		},
		{
			name: "UserID присутствует, но неверного типа (int вместо int64)",
			setupCtx: func() context.Context {
				return context.WithValue(context.Background(), UserIDKey, int(123))
			},
			expectedID: 0,
			expectedOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setupCtx()

			id, ok := GetUserID(ctx)

			assert.Equal(t, tt.expectedOk, ok)
			assert.Equal(t, tt.expectedID, id)
		})
	}
}
