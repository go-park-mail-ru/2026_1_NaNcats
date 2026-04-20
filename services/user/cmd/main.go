package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	infrastructureLogger "github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/common/logger"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/interceptors"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/postgres"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/s3"

	userDelivery "github.com/go-park-mail-ru/2026_1_NaNcats/services/user/internal/delivery/grpc"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/user/internal/infrastructure/config"
	userPG "github.com/go-park-mail-ru/2026_1_NaNcats/services/user/internal/repository/postgres"
	userUsecase "github.com/go-park-mail-ru/2026_1_NaNcats/services/user/internal/usecase"
	pb "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/user"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
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
	pgConfig.ConnConfig.Tracer = postgres.NewDBTracer(appLogger)

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

	userUC := userUsecase.NewUserUseCase(userRepo, s3Repo, cfg.DefaultAvatarURL)
	clientProfileUC := userUsecase.NewClientProfileUseCase(clientProfileRepo)

	userHandler := userDelivery.NewUserHandler(userUC, clientProfileUC)

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptors.UnaryServerRecovery(appLogger),
			interceptors.UnaryServerLogging(appLogger),
		),
	)

	pb.RegisterUserServiceServer(grpcServer, userHandler)
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

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	appLogger.Info("Received shutdown signal, stopping gracefully...")
	grpcServer.GracefulStop()
	appLogger.Info("User microservice stopped")
}
