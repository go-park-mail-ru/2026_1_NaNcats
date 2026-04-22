package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/gomodule/redigo/redis"
	"github.com/jackc/pgx/v5/pgxpool"
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
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/payment/internal/infrastructure/config"
	paymentGrpcClient "github.com/go-park-mail-ru/2026_1_NaNcats/services/payment/internal/infrastructure/grpc_client"
	paymentPG "github.com/go-park-mail-ru/2026_1_NaNcats/services/payment/internal/repository/postgres"
	paymentRedis "github.com/go-park-mail-ru/2026_1_NaNcats/services/payment/internal/repository/redisrepo"
	paymentUseCase "github.com/go-park-mail-ru/2026_1_NaNcats/services/payment/internal/usecase"
)

func main() {
	cfg := config.Load()

	rawLogger, err := logger.NewZapLogger(cfg.Logger.Level)
	if err != nil {
		log.Fatalf("Cannot start without logger: %v", err)
	}
	appLogger := infrastructureLogger.NewLoggerAdapter(rawLogger)
	appLogger.Info("Starting Payment microservice...")

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
	appLogger.Info("Connected to PostgreSQL")

	redisPool := &redis.Pool{
		MaxIdle:     cfg.Redis.MaxIdle,
		IdleTimeout: cfg.Redis.IdleTimeout,
		Dial: func() (redis.Conn, error) {
			return redis.DialURL(cfg.Redis.URL)
		},
	}
	defer redisPool.Close()

	rConn := redisPool.Get()
	if _, err := rConn.Do("PING"); err != nil {
		appLogger.Fatal("Failed to connect to Redis", err)
	}
	rConn.Close()
	appLogger.Info("Connected to Redis")

	orderConn, err := grpc.NewClient(
		cfg.OrderServiceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		appLogger.Fatal("Failed to create Order Service client", err)
	}
	defer orderConn.Close()
	appLogger.Info("Connected to Order Service", logger.String("addr", cfg.OrderServiceAddr))

	orderGrpcClient := orderPb.NewOrderServiceClient(orderConn)
	orderClient := paymentGrpcClient.NewOrderClient(orderGrpcClient)

	yookassaClient := yookassa.NewClient(cfg.Yookassa.ShopID, cfg.Yookassa.SecretKey)

	paymentRepo := paymentPG.NewPaymentRepo(pool)
	cacheRepo := paymentRedis.NewPaymentCacheRepo(redisPool)

	paymentUC := paymentUseCase.NewPaymentUseCase(
		paymentRepo,
		cacheRepo,
		orderClient,
		yookassaClient,
		cfg.Yookassa.ReturnURL,
		appLogger,
	)
	paymentHandler := paymentDelivery.NewPaymentHandler(paymentUC)

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptors.UnaryServerRecovery(appLogger),
			interceptors.UnaryServerLogging(appLogger),
		),
	)

	pb.RegisterPaymentServiceServer(grpcServer, paymentHandler)
	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", ":"+cfg.GRPC.Port)
	if err != nil {
		appLogger.Fatal("Failed to listen port", err)
	}

	go func() {
		appLogger.Info("Payment gRPC server is running", logger.String("port", cfg.GRPC.Port))
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
