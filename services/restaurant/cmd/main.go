package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/exaring/otelpgx"
	infrastructureLogger "github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/common/logger"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/interceptors"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/metrics"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/postgres"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/s3"
	"github.com/joho/godotenv"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"

	restaurantDelivery "github.com/go-park-mail-ru/2026_1_NaNcats/services/restaurant/internal/delivery/grpc"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/restaurant/internal/infrastructure/config"
	restaurantGrpcClient "github.com/go-park-mail-ru/2026_1_NaNcats/services/restaurant/internal/infrastructure/grpc_client"
	restaurantPG "github.com/go-park-mail-ru/2026_1_NaNcats/services/restaurant/internal/repository/postgres"
	restaurantUseCase "github.com/go-park-mail-ru/2026_1_NaNcats/services/restaurant/internal/usecase"
	orderPb "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/order"
	pb "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/restaurant"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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
	appLogger.Info("Starting Restaurant microservice...")

	ctx := context.Background()

	pgConfig, err := pgxpool.ParseConfig(cfg.Postgres.URL)
	if err != nil {
		appLogger.Fatal("database config parsing failed", err)
	}
	pgConfig.MaxConns = 15
	pgConfig.MinConns = 2
	pgConfig.MaxConnLifetime = time.Hour
	pgConfig.MaxConnIdleTime = 30 * time.Minute
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

	s3Repo, err := s3.NewS3Storage(
		ctx,
		cfg.S3.KeyID,
		cfg.S3.SecretKey,
		cfg.S3.BucketName,
		cfg.S3.Region,
	)
	if err != nil {
		appLogger.Fatal("Failed to init S3", err)
	}
	appLogger.Info("Connected to S3 Storage", logger.String("bucket", cfg.S3.BucketName))

	brandRepo := restaurantPG.NewRestaurantBrandRepo(pool)
	dishRepo := restaurantPG.NewDishRepo(pool)

	categoryUC := restaurantUseCase.NewCategoryUseCase(brandRepo)

	brandUC := restaurantUseCase.NewRestaurantBrandUseCase(brandRepo, cfg.DefaultRestaurantLogoURL, s3Repo, appLogger)
	tracedBrandUC := restaurantUseCase.NewRestaurantBrandUseCaseTracingMiddleware(brandUC)

	dishUC := restaurantUseCase.NewDishUseCase(dishRepo, cfg.DefaultFoodLogoURL, s3Repo, appLogger)
	tracedDishUC := restaurantUseCase.NewDishUseCaseTracingMiddleware(dishUC)

	orderConn, err := grpc.NewClient(
		cfg.OrderServiceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	if err != nil {
		appLogger.Fatal("Failed to dial order-service", err)
	}
	defer orderConn.Close()
	appLogger.Info("Connected to Order Service", logger.String("addr", cfg.OrderServiceAddr))
	orderHistoryClient := restaurantGrpcClient.NewOrderClient(orderPb.NewOrderServiceClient(orderConn))

	recoUC := restaurantUseCase.NewRecommendationsUseCase(brandRepo, dishRepo, orderHistoryClient, cfg.DefaultRestaurantLogoURL, cfg.DefaultFoodLogoURL, appLogger)

	restaurantHandler := restaurantDelivery.NewRestaurantHandler(categoryUC, tracedBrandUC, tracedDishUC, recoUC)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

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

	pb.RegisterRestaurantServiceServer(grpcServer, restaurantHandler)
	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", ":"+cfg.GRPC.Port)
	if err != nil {
		appLogger.Fatal("Failed to listen port", err)
	}

	go func() {
		appLogger.Info("Restaurant gRPC server is running", logger.String("port", cfg.GRPC.Port))
		if err := grpcServer.Serve(listener); err != nil {
			appLogger.Fatal("Server failed", err)
		}
	}()

	<-ctx.Done()
	appLogger.Info("Received shutdown signal, stopping gracefully...")
	grpcServer.GracefulStop()
	appLogger.Info("Restaurant microservice stopped")
}
