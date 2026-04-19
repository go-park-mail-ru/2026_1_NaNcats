package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	infrastructureLogger "github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/common/logger"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	"github.com/gomodule/redigo/redis"
	"github.com/joho/godotenv"

	grpcClient "github.com/go-park-mail-ru/2026_1_NaNcats/services/auth/infrastructure/grpc_client"
	authDelivery "github.com/go-park-mail-ru/2026_1_NaNcats/services/auth/internal/delivery/grpc"
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

	// 1. Инициализация логгера
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}
	rawLogger, err := logger.NewZapLogger(logLevel)
	if err != nil {
		log.Fatalf("Cannot start without logger: %v", err)
	}
	appLogger := infrastructureLogger.NewLoggerAdapter(rawLogger)
	appLogger.Info("Starting Auth microservice...")

	port := os.Getenv("GRPC_PORT")
	if port == "" {
		port = "50054"
	}

	userServiceAddr := os.Getenv("USER_SERVICE_ADDR")
	if userServiceAddr == "" {
		// дефолтный адрес сервиса юзеров для локалки
		userServiceAddr = "localhost:50052"
	}

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://user:@localhost:6379/0"
	}

	redisPool := &redis.Pool{
		MaxIdle:     10,
		IdleTimeout: 60 * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.DialURL(redisURL)
		},
	}
	defer redisPool.Close()

	conn := redisPool.Get()
	if _, err := conn.Do("PING"); err != nil {
		appLogger.Fatal("Failed to connect to Redis", err)
	}
	conn.Close()
	appLogger.Info("Connected to Redis")

	userConn, err := grpc.Dial(userServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		appLogger.Fatal("Failed to dial User Service", err)
	}
	defer userConn.Close()

	userGrpcClient := pbUser.NewUserServiceClient(userConn)
	appLogger.Info("gRPC Client connected to User Service", logger.String("addr", userServiceAddr))

	sessionRepo := redisRepo.NewSessionRepo(redisPool)

	userClient := grpcClient.NewUserClient(userGrpcClient)

	sessionTTL := 24 * time.Hour // TODO: вынеси в env или конфиг сделать
	sessionUC := authUsecase.NewSessionUseCase(sessionRepo, sessionTTL)
	authUC := authUsecase.NewAuthUseCase(userClient, sessionUC)

	authHandler := authDelivery.NewAuthHandler(authUC)

	grpcServer := grpc.NewServer()
	pbAuth.RegisterAuthServiceServer(grpcServer, authHandler)
	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		appLogger.Fatal("Failed to listen port", err)
	}

	go func() {
		appLogger.Info("Auth gRPC server is running", logger.String("port", port))
		if err := grpcServer.Serve(listener); err != nil {
			appLogger.Fatal("Server failed", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	appLogger.Info("Received shutdown signal, stopping gracefully...")
	grpcServer.GracefulStop()
	appLogger.Info("Auth microservice stopped")
}
