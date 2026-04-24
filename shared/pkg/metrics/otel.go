package metrics

import (
	"context"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.9.0"
)

// Настройка пути, по которому метрики из кода попадут в Sidecar
func InitMetrics(ctx context.Context, serviceName string, collectorAddr string) (func(), error) {
	// Экспортер, который шлет данные в Sidecar
	exporter, err := otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithEndpoint(collectorAddr),
		otlpmetricgrpc.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	// Resouce, который прикреплен к метрикам каждого сервиса
	res, _ := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(serviceName),
		),
	)

	// Управляет созданием всех счетчиков
	provider := metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(metric.NewPeriodicReader(exporter)),
	)

	otel.SetMeterProvider(provider)

	err = runtime.Start(runtime.WithMinimumReadMemStatsInterval(time.Second))
	if err != nil {
		return nil, err
	}

	// Возвращаем функцию-замыкание для корректного закрытия
	return func() { _ = provider.Shutdown(context.Background()) }, nil
}
