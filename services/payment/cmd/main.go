package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/exaring/otelpgx"
	"github.com/gomodule/redigo/redis"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/api_clients/yookassa"
	infrastructureLogger "github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/common/logger"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/interceptors"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/metrics"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/postgres"

	paymentRabbitMQ "github.com/go-park-mail-ru/2026_1_NaNcats/services/payment/internal/delivery/rabbitmq"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/rabbitmq"

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
	_ = godotenv.Load()
	cfg := config.Load()

	rawLogger, err := logger.NewZapLogger(cfg.Logger.Level)
	if err != nil {
		log.Fatalf("Cannot start without logger: %v", err)
	}
	appLogger := infrastructureLogger.NewLoggerAdapter(rawLogger)
	appLogger.Info("Starting Payment microservice...")

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

	rabbitClient, err := rabbitmq.NewRabbitClient(cfg.RabbitMQURL, appLogger)
	if err != nil {
		appLogger.Fatal("Failed to connect to RabbitMQ", err)
	}
	defer rabbitClient.Close()

	paymentConsumer := paymentRabbitMQ.NewPaymentConsumer(rabbitClient, paymentUC, appLogger)
	if err := paymentConsumer.Start(ctx); err != nil {
		appLogger.Fatal("Failed to start Payment RabbitMQ consumer", err)
	}

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

	<-ctx.Done()
	appLogger.Info("Received shutdown signal, stopping gracefully...")
	grpcServer.GracefulStop()
	appLogger.Info("Payment microservice stopped")
}
