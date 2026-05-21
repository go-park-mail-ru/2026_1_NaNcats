package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"

	analyticsConsumer "github.com/go-park-mail-ru/2026_1_NaNcats/services/analytics/internal/delivery/rabbitmq"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/analytics/internal/infrastructure/config"
	clickhouseRepo "github.com/go-park-mail-ru/2026_1_NaNcats/services/analytics/internal/repository/clickhouse"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/analytics/internal/usecase"
	infrastructureLogger "github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/common/logger"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/metrics"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/rabbitmq"
)

func main() {
	_ = godotenv.Load()
	cfg := config.Load()

	rawLogger, err := logger.NewZapLogger(cfg.Logger.Level)
	if err != nil {
		log.Fatalf("Cannot start without logger: %v", err)
	}
	appLogger := infrastructureLogger.NewLoggerAdapter(rawLogger)
	appLogger.Info("Starting Analytics Ingester microservice...")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cleanupMetrics, err := metrics.InitMetrics(ctx, cfg.OTEL.ServiceName, cfg.OTEL.CollectorAddr)
	if err != nil {
		appLogger.Fatal("Failed to init metrics", err)
	}
	defer cleanupMetrics()

	cleanupTracing, err := metrics.InitTracing(ctx, cfg.OTEL.ServiceName, cfg.OTEL.CollectorAddr)
	if err != nil {
		appLogger.Fatal("Failed to init tracing", err)
	}
	defer cleanupTracing()

	// Подключение к ClickHouse по Native Protocol
	chConn, err := clickhouseRepo.NewClickHouseClient(cfg.ClickHouse)
	if err != nil {
		appLogger.Fatal("Failed to connect to ClickHouse", err)
	}
	defer chConn.Close()
	appLogger.Info("Connected to ClickHouse", logger.String("addr", cfg.ClickHouse.Host))

	repo := clickhouseRepo.NewAnalyticsRepository(
		chConn,
		appLogger,
		cfg.Ingester.BatchSize,
		cfg.Ingester.FlushInterval,
	)

	uc := usecase.NewAnalyticsUseCase(repo, appLogger)

	rabbitClient, err := rabbitmq.NewRabbitClient(cfg.RabbitMQURL, appLogger)
	if err != nil {
		appLogger.Fatal("Failed to connect to RabbitMQ", err)
	}
	defer rabbitClient.Close()
	appLogger.Info("Connected to RabbitMQ")

	consumer := analyticsConsumer.NewAnalyticsConsumer(rabbitClient, uc, appLogger)
	if err := consumer.Start(ctx); err != nil {
		appLogger.Fatal("Failed to start RabbitMQ consumer", err)
	}

	appLogger.Info("Analytics Ingester is fully running. Waiting for events...")

	<-ctx.Done()
	appLogger.Info("Received shutdown signal, stopping Analytics Ingester gracefully...")

	if err := repo.Close(); err != nil {
		appLogger.Error("failed to gracefully close analytics repository", err)
	} else {
		appLogger.Info("Analytics repository gracefully stopped, all buffered data flushed")
	}
}
