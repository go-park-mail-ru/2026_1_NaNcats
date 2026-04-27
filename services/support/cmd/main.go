package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/exaring/otelpgx"
	infrastructureLogger "github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/common/logger"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/interceptors"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/metrics"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/postgres"

	pb "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/support"

	supportDelivery "github.com/go-park-mail-ru/2026_1_NaNcats/services/support/internal/delivery/grpc"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/support/internal/infrastructure/config"
	supportPG "github.com/go-park-mail-ru/2026_1_NaNcats/services/support/internal/repository/postgres"
	supportUseCase "github.com/go-park-mail-ru/2026_1_NaNcats/services/support/internal/usecase"
)

func main() {
	_ = godotenv.Load()
	cfg := config.Load()

	rawLogger, err := logger.NewZapLogger(cfg.Logger.Level)
	if err != nil {
		log.Fatalf("Cannot start without logger: %v", err)
	}
	appLogger := infrastructureLogger.NewLoggerAdapter(rawLogger)
	appLogger.Info("Starting Support microservice...")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pgConfig, err := pgxpool.ParseConfig(cfg.Postgres.URL)
	if err != nil {
		appLogger.Fatal("database config parsing failed", err)
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
	appLogger.Info("Connected to PostgreSQL")

	supportRepo := supportPG.NewSupportRepo(pool)
	supportUC := supportUseCase.NewSupportUseCase(supportRepo)
	supportHandler := supportDelivery.NewSupportHandler(supportUC)

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

	pb.RegisterSupportServiceServer(grpcServer, supportHandler)
	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", ":"+cfg.GRPC.Port)
	if err != nil {
		appLogger.Fatal("Failed to listen port", err)
	}

	go func() {
		appLogger.Info("Support gRPC server is running", logger.String("port", cfg.GRPC.Port))
		if err := grpcServer.Serve(listener); err != nil {
			appLogger.Fatal("Server failed", err)
		}
	}()

	<-ctx.Done()
	appLogger.Info("Received shutdown signal, stopping gracefully...")
	grpcServer.GracefulStop()
	appLogger.Info("Support microservice stopped")
}
