package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/exaring/otelpgx"
	userRabbit "github.com/go-park-mail-ru/2026_1_NaNcats/services/user/internal/delivery/rabbitmq"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/user/internal/infrastructure/grpc_client"
	infrastructureLogger "github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/common/logger"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/interceptors"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/metrics"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/postgres"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/rabbitmq"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/s3"
	"github.com/joho/godotenv"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"

	userDelivery "github.com/go-park-mail-ru/2026_1_NaNcats/services/user/internal/delivery/grpc"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/user/internal/infrastructure/config"
	userPG "github.com/go-park-mail-ru/2026_1_NaNcats/services/user/internal/repository/postgres"
	userUsecase "github.com/go-park-mail-ru/2026_1_NaNcats/services/user/internal/usecase"
	pbOrder "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/order"
	pb "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/user"

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
	appLogger.Info("Starting User microservice...")

	ctx := context.Background()

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

	userRepo := userPG.NewUserRepo(pool)
	clientProfileRepo := userPG.NewClientProfileRepo(pool)
	achievementRepo := userPG.NewAchievementRepo(pool)
	wordleRepo := userPG.NewWordleRepo(pool)

	rabbitClient, err := rabbitmq.NewRabbitClient(cfg.RabbitMQURL, appLogger)
	if err != nil {
		appLogger.Fatal("Failed to connect to RabbitMQ", err)
	}
	defer rabbitClient.Close()

	orderConn, err := grpc.NewClient(
		cfg.OrderServiceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	if err != nil {
		appLogger.Fatal("Failed to connect to Order Service", err)
	}
	defer orderConn.Close()
	appLogger.Info("Connected to Order Service (gRPC client)", logger.String("addr", cfg.OrderServiceAddr))

	orderGrpcClient := pbOrder.NewOrderServiceClient(orderConn)
	promoGrpcClient := pbOrder.NewPromoServiceClient(orderConn)
	orderClient := grpc_client.NewOrderClient(orderGrpcClient, promoGrpcClient)

	userUC := userUsecase.NewUserUseCase(userRepo, s3Repo, cfg.DefaultAvatarURL, rabbitClient, appLogger)
	tracedUserUC := userUsecase.NewUserUseCaseTracingMiddleware(userUC)

	achievementUC := userUsecase.NewAchievementUseCase(achievementRepo, appLogger)

	clientProfileUC := userUsecase.NewClientProfileUseCase(clientProfileRepo, orderClient, achievementUC)
	tracedProfileUC := userUsecase.NewClientProfileUseCaseTracingMiddleware(clientProfileUC)

	wordleUC := userUsecase.NewWordleUseCase(wordleRepo, appLogger)

	userConsumer := userRabbit.NewUserConsumer(rabbitClient, achievementUC, appLogger)
	if err := userConsumer.Start(ctx); err != nil {
		appLogger.Fatal("Failed to start User RabbitMQ consumer", err)
	}

	userHandler := userDelivery.NewUserHandler(tracedUserUC, tracedProfileUC, achievementUC)
	gameHandler := userDelivery.NewGameHandler(wordleUC)

	// Контекст, который отменяется по сигналу ОС
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

	pb.RegisterUserServiceServer(grpcServer, userHandler)
	pb.RegisterWordleServiceServer(grpcServer, gameHandler)
	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", ":"+cfg.GRPC.Port)
	if err != nil {
		appLogger.Fatal("Failed to listen port", err)
	}

	go func() {
		appLogger.Info("User gRPC server is running", logger.String("port", cfg.GRPC.Port))
		if err := grpcServer.Serve(listener); err != nil {
			appLogger.Fatal("Server failed", err)
		}
	}()

	<-ctx.Done()
	appLogger.Info("Received shutdown signal, stopping gracefully...")
	grpcServer.GracefulStop()
	appLogger.Info("User microservice stopped")
}
