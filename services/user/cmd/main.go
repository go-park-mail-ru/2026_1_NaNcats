package main

import (
	"context"
	"errors"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	infrastructureLogger "github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/common/logger"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/postgres"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/s3"
	"github.com/joho/godotenv"

	userDelivery "github.com/go-park-mail-ru/2026_1_NaNcats/services/user/internal/delivery/grpc"
	userPG "github.com/go-park-mail-ru/2026_1_NaNcats/services/user/internal/repository/postgres"
	userUsecase "github.com/go-park-mail-ru/2026_1_NaNcats/services/user/internal/usecase"
	pb "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/user"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	_ = godotenv.Load()

	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}

	rawLogger, err := logger.NewZapLogger(logLevel)
	if err != nil {
		log.Fatalf("Cannot start without logger: %v", err)
	}
	appLogger := infrastructureLogger.NewLoggerAdapter(rawLogger)
	appLogger.Info("Starting User microservice...")

	port := os.Getenv("GRPC_PORT")
	if port == "" {
		port = "50052"
	}

	ctx := context.Background()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		appLogger.Fatal("database connection string is missing", errors.New("DATABASE_URL env var is empty"))
	}

	config, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		appLogger.Fatal("config parsing failed", err)
	}

	config.ConnConfig.Tracer = postgres.NewDBTracer(appLogger)

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		appLogger.Fatal("database pool creation failed", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		appLogger.Fatal("could not ping the database", err)
	}
	appLogger.Info("Connected to PostgreSQL")

	keyID := os.Getenv("S3_KEY_ID")
	s3SecretKey := os.Getenv("S3_SECRET_KEY")
	bucketName := "nancats-bucket"

	s3Repo, err := s3.NewS3Storage(ctx, keyID, s3SecretKey, bucketName, "ru-central1")
	if err != nil {
		appLogger.Fatal("Failed to init S3", err)
	}
	appLogger.Info("Connected to S3 Storage")

	defaultAvatarURL := os.Getenv("DEFAULT_AVATAR_URL")
	if defaultAvatarURL == "" {
		appLogger.Warn("DEFAULT_AVATAR_URL is empty, frontend might break when requesting default avatar")
	}

	userRepo := userPG.NewUserRepo(pool)
	clientProfileRepo := userPG.NewClientProfileRepo(pool)

	userUC := userUsecase.NewUserUseCase(userRepo, s3Repo, defaultAvatarURL)
	clientProfileUC := userUsecase.NewClientProfileUseCase(clientProfileRepo)

	userHandler := userDelivery.NewUserHandler(userUC, clientProfileUC)

	// TODO: добавить gRPC Interceptor для логирования всех запросов
	grpcServer := grpc.NewServer()

	pb.RegisterUserServiceServer(grpcServer, userHandler)
	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		appLogger.Fatal("Failed to listen port", err)
	}

	go func() {
		appLogger.Info("User gRPC server is running", logger.String("port", port))
		if err := grpcServer.Serve(listener); err != nil {
			appLogger.Fatal("Server failed", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	appLogger.Info("Received shutdown signal, stopping gracefully...")
	grpcServer.GracefulStop()
	appLogger.Info("User microservice stopped")
}
