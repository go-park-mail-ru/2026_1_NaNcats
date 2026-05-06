package logger

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewZapLogger(t *testing.T) {
	tests := []struct {
		name      string
		levelStr  string
		expectErr bool
	}{
		{
			name:      "Успешное создание с уровнем debug (DevelopmentConfig)",
			levelStr:  "debug",
			expectErr: false,
		},
		{
			name:      "Успешное создание с уровнем info (ProductionConfig)",
			levelStr:  "info",
			expectErr: false,
		},
		{
			name:      "Успешное создание с неизвестным уровнем (fallback на info)",
			levelStr:  "unknown_level",
			expectErr: false, // Ошибки не будет, произойдет fallback
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := NewZapLogger(tt.levelStr)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, logger)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, logger)
				assert.NotNil(t, logger.GetInternal())
			}
		})
	}
}

func TestZapLogger_Methods(t *testing.T) {
	// Создаем реальный инстанс для проверки паник и корректности маппинга
	zLog, err := NewZapLogger("debug")
	assert.NoError(t, err)

	tests := []struct {
		name   string
		action func()
	}{
		{
			name: "Успешный вызов Info",
			action: func() {
				zLog.Info("test info", map[string]any{"key": "val"})
			},
		},
		{
			name: "Успешный вызов Error",
			action: func() {
				zLog.Error("test error", errors.New("db error"), map[string]any{"user_id": 1})
			},
		},
		{
			name: "Успешный вызов With",
			action: func() {
				newZLog := zLog.With(zap.String("req_id", "123"))
				assert.NotNil(t, newZLog)
				assert.NotEqual(t, zLog, newZLog) // Должен вернуться новый инстанс
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotPanics(t, tt.action)
		})
	}
}
