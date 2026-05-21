package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	analyticsDelivery "github.com/go-park-mail-ru/2026_1_NaNcats/services/analytics/internal/delivery/grpc"
	analyticsConsumer "github.com/go-park-mail-ru/2026_1_NaNcats/services/analytics/internal/delivery/rabbitmq"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/analytics/internal/infrastructure/config"
	clickhouseRepo "github.com/go-park-mail-ru/2026_1_NaNcats/services/analytics/internal/repository/clickhouse"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/analytics/internal/usecase"
	infrastructureLogger "github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/common/logger"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/interceptors"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/metrics"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/rabbitmq"
	pb "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/analytics"
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

	grpcServer := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(
			interceptors.UnaryServerRecovery(appLogger),
			interceptors.UnaryServerLogging(appLogger),
			interceptors.UnaryServerUserIDKey(),
		),
	)

	analyticsHandler := analyticsDelivery.NewAnalyticsHandler(uc)
	pb.RegisterAnalyticsServiceServer(grpcServer, analyticsHandler)
	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", ":"+cfg.GRPC.Port)
	if err != nil {
		appLogger.Fatal("Failed to listen port", err)
	}

	go func() {
		appLogger.Info("Analytics gRPC query server is running", logger.String("port", cfg.GRPC.Port))
		if err := grpcServer.Serve(listener); err != nil {
			appLogger.Fatal("gRPC server failed to start", err)
		}
	}()

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

	grpcServer.GracefulStop()
	appLogger.Info("gRPC server stopped")

	rabbitClient.Close()
	appLogger.Info("RabbitMQ consumer stopped")

	if err := repo.Close(); err != nil {
		appLogger.Error("failed to gracefully close analytics repository", err)
	} else {
		appLogger.Info("Analytics repository gracefully stopped, all buffered data flushed")
	}
}
