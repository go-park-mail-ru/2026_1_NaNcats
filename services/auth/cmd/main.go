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
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/metrics"
	"github.com/gomodule/redigo/redis"
	"github.com/joho/godotenv"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"

	authDelivery "github.com/go-park-mail-ru/2026_1_NaNcats/services/auth/internal/delivery/grpc"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/auth/internal/infrastructure/config"
	grpcClient "github.com/go-park-mail-ru/2026_1_NaNcats/services/auth/internal/infrastructure/grpc_client"
	redisRepo "github.com/go-park-mail-ru/2026_1_NaNcats/services/auth/internal/repository/redisrepo"
	authUsecase "github.com/go-park-mail-ru/2026_1_NaNcats/services/auth/internal/usecase"

	pbAuth "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/auth"
	pbUser "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/user"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

func main() {
	_ = godotenv.Load()
	cfg := config.Load()

	ctx := context.Background()

	// Контекст, который отменяется по сигналу ОС
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	rawLogger, err := logger.NewZapLogger(cfg.Logger.Level)
	if err != nil {
		log.Fatalf("Cannot start without logger: %v", err)
	}
	appLogger := infrastructureLogger.NewLoggerAdapter(rawLogger)
	appLogger.Info("Starting Auth microservice...")

	redisPool := &redis.Pool{
		MaxIdle:     cfg.Redis.MaxIdle,
		IdleTimeout: cfg.Redis.IdleTimeout,
		Dial: func() (redis.Conn, error) {
			return redis.DialURL(cfg.Redis.URL)
		},
	}
	defer redisPool.Close()

	conn := redisPool.Get()
	if _, err := conn.Do("PING"); err != nil {
		appLogger.Fatal("Failed to connect to Redis", err)
	}
	conn.Close()
	appLogger.Info("Connected to Redis")

	userConn, err := grpc.NewClient(
		cfg.UserServiceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		appLogger.Fatal("Failed to create User Service client", err)
	}
	defer userConn.Close()

	userGrpcClient := pbUser.NewUserServiceClient(userConn)
	userClient := grpcClient.NewUserClient(userGrpcClient)
	appLogger.Info("gRPC Client connected to User Service", logger.String("addr", cfg.UserServiceAddr))

	sessionRepo := redisRepo.NewSessionRepo(redisPool)
	sessionUC := authUsecase.NewSessionUseCase(sessionRepo, cfg.SessionTTL)
	authUC := authUsecase.NewAuthUseCase(userClient, sessionUC)
	authHandler := authDelivery.NewAuthHandler(authUC)

	cleanup, err := metrics.InitMetrics(ctx, cfg.OTEL.ServiceName, cfg.OTEL.CollectorAddr)
	if err != nil {
		appLogger.Fatal("failed to init metrics", err)
	}
	defer cleanup()

	grpcServer := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(
			interceptors.UnaryServerRecovery(appLogger),
			interceptors.UnaryServerLogging(appLogger),
		),
	)

	pbAuth.RegisterAuthServiceServer(grpcServer, authHandler)
	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", ":"+cfg.GRPC.Port)
	if err != nil {
		appLogger.Fatal("Failed to listen port", err)
	}

	go func() {
		appLogger.Info("Auth gRPC server is running", logger.String("port", cfg.GRPC.Port))
		if err := grpcServer.Serve(listener); err != nil {
			appLogger.Fatal("Server failed", err)
		}
	}()

	<-ctx.Done()
	appLogger.Info("Received shutdown signal, stopping gracefully...")
	grpcServer.GracefulStop()
	appLogger.Info("Auth microservice stopped")
}
