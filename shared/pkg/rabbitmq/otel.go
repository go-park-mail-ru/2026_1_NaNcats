package rabbitmq

import (
	"context"

	"go.opentelemetry.io/otel"
)

// Адаптирует заголовки RabbitMQ для OpenTelemetry
type AMQPHeadersCarrier map[string]interface{}

func (a AMQPHeadersCarrier) Get(key string) string {
	v, ok := a[key]
	if !ok {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		return ""
	}
	return s
}

func (a AMQPHeadersCarrier) Set(key string, value string) {
	a[key] = value
}

func (a AMQPHeadersCarrier) Keys() []string {
	keys := make([]string, 0, len(a))
	for k := range a {
		keys = append(keys, k)
	}
	return keys
}

// Записывает TraceID из context в заголовки RabbitMQ
func InjectContext(ctx context.Context, headers map[string]interface{}) {
	otel.GetTextMapPropagator().Inject(ctx, AMQPHeadersCarrier(headers))
}

// Извлекает TraceID из заголовков RabbitMQ и создает новый context
func ExtractContext(ctx context.Context, headers map[string]interface{}) context.Context {
	return otel.GetTextMapPropagator().Extract(ctx, AMQPHeadersCarrier(headers))
}
