package metrics

import (
	"context"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
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

func InitTracing(ctx context.Context, serviceName string, collectorAddr string) (func(), error) {
	traceExporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(collectorAddr),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	res, _ := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(serviceName),
		),
	)

	tp := trace.NewTracerProvider(
		trace.WithBatcher(traceExporter),
		trace.WithResource(res),
		trace.WithSampler(trace.AlwaysSample()),
	)

	otel.SetTracerProvider(tp)

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return func() { _ = tp.Shutdown(context.Background()) }, nil
}
