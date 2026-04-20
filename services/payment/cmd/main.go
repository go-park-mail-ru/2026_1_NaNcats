package main

import (
	"context"
	"errors"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/api_clients/yookassa"
	infrastructureLogger "github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/common/logger"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/interceptors"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/postgres"

	orderPb "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/order"
	pb "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/payment"

	paymentDelivery "github.com/go-park-mail-ru/2026_1_NaNcats/services/payment/internal/delivery/grpc"
	paymentGrpcClient "github.com/go-park-mail-ru/2026_1_NaNcats/services/payment/internal/infrastructure/grpc_client"
	paymentPG "github.com/go-park-mail-ru/2026_1_NaNcats/services/payment/internal/repository/postgres"
	paymentRedis "github.com/go-park-mail-ru/2026_1_NaNcats/services/payment/internal/repository/redisrepo"
	paymentUseCase "github.com/go-park-mail-ru/2026_1_NaNcats/services/payment/internal/usecase"
)

func main() {
	_ = godotenv.Load()

	logLevel := getEnv("LOG_LEVEL", "info")
	rawLogger, err := logger.NewZapLogger(logLevel)
	if err != nil {
		log.Fatalf("Cannot start without logger: %v", err)
	}
	appLogger := infrastructureLogger.NewLoggerAdapter(rawLogger)
	appLogger.Info("Starting Payment microservice...")

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		appLogger.Fatal("database connection string is missing", errors.New("DATABASE_URL env var is empty"))
	}
	ctx := context.Background()
	pgConfig, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		appLogger.Fatal("config parsing failed", err)
	}
	pgConfig.ConnConfig.Tracer = postgres.NewDBTracer(appLogger)
	pool, err := pgxpool.NewWithConfig(ctx, pgConfig)
	if err != nil {
		appLogger.Fatal("database pool creation failed", err)
	}
	defer pool.Close()
	appLogger.Info("Connected to PostgreSQL")

	redisURL := getEnv("REDIS_URL", "redis://localhost:6379/0")
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

	orderAddr := getEnv("ORDER_SERVICE_ADDR", "localhost:50057")
	orderConn, err := grpc.NewClient(
		orderAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		appLogger.Fatal("Failed to create Order Service client", err)
	}
	defer orderConn.Close()
	appLogger.Info("Connected to Order Service", logger.String("addr", orderAddr))

	orderGrpcClient := orderPb.NewOrderServiceClient(orderConn)
	orderClient := paymentGrpcClient.NewOrderClient(orderGrpcClient)

	shopID := os.Getenv("YOOKASSA_SHOP_ID")
	secretKey := os.Getenv("YOOKASSA_SECRET_KEY")
	returnURL := os.Getenv("YOOKASSA_RETURN_URL")
	if shopID == "" || secretKey == "" || returnURL == "" {
		appLogger.Fatal("Yookassa config missing", errors.New("check YOOKASSA_SHOP_ID, YOOKASSA_SECRET_KEY, YOOKASSA_RETURN_URL"))
	}
	yookassaClient := yookassa.NewClient(shopID, secretKey)

	paymentRepo := paymentPG.NewPaymentRepo(pool)
	cacheRepo := paymentRedis.NewPaymentCacheRepo(redisPool)

	paymentUC := paymentUseCase.NewPaymentUseCase(
		paymentRepo,
		cacheRepo,
		orderClient,
		yookassaClient,
		returnURL,
	)
	paymentHandler := paymentDelivery.NewPaymentHandler(paymentUC)

	port := getEnv("GRPC_PORT", "50056")
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptors.UnaryServerRecovery(appLogger),
			interceptors.UnaryServerLogging(appLogger),
		),
	)

	pb.RegisterPaymentServiceServer(grpcServer, paymentHandler)
	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		appLogger.Fatal("Failed to listen port", err)
	}

	go func() {
		appLogger.Info("Payment gRPC server is running", logger.String("port", port))
		if err := grpcServer.Serve(listener); err != nil {
			appLogger.Fatal("Server failed", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	appLogger.Info("Received shutdown signal, stopping gracefully...")
	grpcServer.GracefulStop()
	appLogger.Info("Payment microservice stopped")
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
