package rabbitmq

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAMQPHeadersCarrier_Get(t *testing.T) {
	tests := []struct {
		name     string
		carrier  AMQPHeadersCarrier
		key      string
		expected string
	}{
		{
			name: "Ключ существует и является строкой",
			carrier: AMQPHeadersCarrier{
				"traceparent": "00-12345-67890-01",
			},
			key:      "traceparent",
			expected: "00-12345-67890-01",
		},
		{
			name: "Ключ не существует",
			carrier: AMQPHeadersCarrier{
				"other_key": "value",
			},
			key:      "traceparent",
			expected: "",
		},
		{
			name: "Ключ существует, но не является строкой",
			carrier: AMQPHeadersCarrier{
				"traceparent": 12345, // int, а не string
			},
			key:      "traceparent",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val := tt.carrier.Get(tt.key)
			assert.Equal(t, tt.expected, val)
		})
	}
}

func TestAMQPHeadersCarrier_Set(t *testing.T) {
	t.Run("Успешная установка значения", func(t *testing.T) {
		carrier := make(AMQPHeadersCarrier)
		carrier.Set("my_key", "my_val")

		assert.Contains(t, carrier, "my_key")
		assert.Equal(t, "my_val", carrier["my_key"])
	})
}

func TestAMQPHeadersCarrier_Keys(t *testing.T) {
	t.Run("Получение списка ключей", func(t *testing.T) {
		carrier := AMQPHeadersCarrier{
			"key1": "val1",
			"key2": "val2",
		}
		keys := carrier.Keys()

		assert.Len(t, keys, 2)
		assert.Contains(t, keys, "key1")
		assert.Contains(t, keys, "key2")
	})
}

func TestInjectAndExtractContext(t *testing.T) {
	t.Run("Успешная инъекция и извлечение контекста", func(t *testing.T) {
		ctx := context.Background()
		headers := make(map[string]interface{})

		InjectContext(ctx, headers)
		extractedCtx := ExtractContext(ctx, headers)

		assert.NotNil(t, extractedCtx)
	})
}
