package metrics

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitMetrics(t *testing.T) {
	tests := []struct {
		name          string
		serviceName   string
		collectorAddr string
		expectErr     bool
	}{
		{
			name:          "Успешная инициализация метрик",
			serviceName:   "test-metrics-service",
			collectorAddr: "localhost:4317",
			expectErr:     false,
		},
		{
			name:          "Успешная инициализация с пустым адресом (проверка на отсутствие паники)",
			serviceName:   "test-metrics-service-empty",
			collectorAddr: "",
			expectErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			shutdown, err := InitMetrics(ctx, tt.serviceName, tt.collectorAddr)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, shutdown)
			} else {
				require.NoError(t, err)
				require.NotNil(t, shutdown)

				// Проверяем, что функция-замыкание корректно отрабатывает и не вызывает панику
				assert.NotPanics(t, shutdown)
			}
		})
	}
}

func TestInitTracing(t *testing.T) {
	tests := []struct {
		name          string
		serviceName   string
		collectorAddr string
		expectErr     bool
	}{
		{
			name:          "Успешная инициализация трейсинга",
			serviceName:   "test-tracing-service",
			collectorAddr: "localhost:4317",
			expectErr:     false,
		},
		{
			name:          "Успешная инициализация трейсинга с пустым адресом",
			serviceName:   "test-tracing-service-empty",
			collectorAddr: "",
			expectErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			shutdown, err := InitTracing(ctx, tt.serviceName, tt.collectorAddr)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, shutdown)
			} else {
				require.NoError(t, err)
				require.NotNil(t, shutdown)

				// Проверяем корректное завершение работы провайдера
				assert.NotPanics(t, shutdown)
			}
		})
	}
}
