package main

import (
	"context"
	"log"
	"net"

	"github.com/exaring/otelpgx"
	infrastructureLogger "github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/common/logger"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/interceptors"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/metrics"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/postgres"
	"github.com/joho/godotenv"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"

	addressDelivery "github.com/go-park-mail-ru/2026_1_NaNcats/services/address/internal/delivery/grpc"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/address/internal/infrastructure/config"
	addressPG "github.com/go-park-mail-ru/2026_1_NaNcats/services/address/internal/repository/postgres"
	addressUseCase "github.com/go-park-mail-ru/2026_1_NaNcats/services/address/internal/usecase"
	pb "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/address"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	_ = godotenv.Load()
	cfg := config.Load()

	rawLogger, err := logger.NewZapLogger(cfg.Logger.Level)
	if err != nil {
		log.Fatalf("Cannot start without logger: %v", err)
	}

	appLogger := infrastructureLogger.NewLoggerAdapter(rawLogger)
	appLogger.Info("Starting Address microservice...")

	ctx := context.Background()

	pgConfig, err := pgxpool.ParseConfig(cfg.Postgres.URL)
	if err != nil {
		appLogger.Fatal("config parsing failed", err)
	}
	consoleTracer := postgres.NewDBTracer(appLogger)
	otelOptions := []otelpgx.Option{
		otelpgx.WithTracerAttributes(semconv.ServiceNameKey.String(cfg.OTEL.ServiceName)),
	}
	if cfg.Logger.Level == "debug" {
		otelOptions = append(otelOptions, otelpgx.WithIncludeQueryParameters())
	}
	otelTracer := otelpgx.NewTracer(
		otelOptions...,
	)
	pgConfig.ConnConfig.Tracer = postgres.NewMultiTracer(consoleTracer, otelTracer)
	pool, err := pgxpool.NewWithConfig(ctx, pgConfig)
	if err != nil {
		appLogger.Fatal("database pool creation failed", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		appLogger.Fatal("could not ping the database", err)
	}
	appLogger.Info("Connected to PostgreSQL")

	addressRepo := addressPG.NewAddressRepo(pool)
	addressUC := addressUseCase.NewAddressUseCase(addressRepo)
	tracedAddressUC := addressUseCase.NewAddressUseCaseTracingMiddleware(addressUC)
	addressHandler := addressDelivery.NewAddressHandler(tracedAddressUC)

	cleanup, err := metrics.InitMetrics(ctx, cfg.OTEL.ServiceName, cfg.OTEL.CollectorAddr)
	if err != nil {
		appLogger.Fatal("failed to init metrics", err)
	}
	defer cleanup()

	cleanupTracing, err := metrics.InitTracing(ctx, cfg.OTEL.ServiceName, cfg.OTEL.CollectorAddr)
	if err != nil {
		appLogger.Fatal("failed to init tracing", err)
	}
	defer cleanupTracing()

	grpcServer := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(
			interceptors.UnaryServerRecovery(appLogger),
			interceptors.UnaryServerLogging(appLogger),
		),
	)

	pb.RegisterAddressServiceServer(grpcServer, addressHandler)
	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", ":"+cfg.GRPC.Port)
	if err != nil {
		appLogger.Fatal("Failed to listen port", err)
	}

	go func() {
		appLogger.Info("Address gRPC server is running", logger.String("port", cfg.GRPC.Port))
		if err := grpcServer.Serve(listener); err != nil {
			appLogger.Fatal("Server failed", err)
		}
	}()

	<-ctx.Done()
	appLogger.Info("Received shutdown signal, stopping gracefully...")
	grpcServer.GracefulStop()
	appLogger.Info("Address microservice stopped")
}
