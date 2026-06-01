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
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	infrastructureLogger "github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/common/logger"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/interceptors"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/metrics"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/postgres"

	orderRabbitMQ "github.com/go-park-mail-ru/2026_1_NaNcats/services/order/internal/delivery/rabbitmq"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/order/internal/infrastructure/autoadvance"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/rabbitmq"

	addressPb "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/address"
	cartPb "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/cart"
	pb "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/order"
	restaurantPb "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/restaurant"

	orderDelivery "github.com/go-park-mail-ru/2026_1_NaNcats/services/order/internal/delivery/grpc"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/order/internal/infrastructure/config"
	orderGrpcClient "github.com/go-park-mail-ru/2026_1_NaNcats/services/order/internal/infrastructure/grpc_client"
	orderPG "github.com/go-park-mail-ru/2026_1_NaNcats/services/order/internal/repository/postgres"
	orderUseCase "github.com/go-park-mail-ru/2026_1_NaNcats/services/order/internal/usecase"
)

func main() {
	_ = godotenv.Load()
	cfg := config.Load()

	rawLogger, err := logger.NewZapLogger(cfg.Logger.Level)
	if err != nil {
		log.Fatalf("Cannot start without logger: %v", err)
	}
	appLogger := infrastructureLogger.NewLoggerAdapter(rawLogger)
	appLogger.Info("Starting Order microservice (Orchestrator)...")

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

	addrConn := createGrpcConn(cfg.AddressServiceAddr, "Address", appLogger)
	defer addrConn.Close()
	cartConn := createGrpcConn(cfg.CartServiceAddr, "Cart", appLogger)
	defer cartConn.Close()
	resConn := createGrpcConn(cfg.RestaurantServiceAddr, "Restaurant", appLogger)
	defer resConn.Close()

	addressGrpcClient := addressPb.NewAddressServiceClient(addrConn)
	addressClient := orderGrpcClient.NewAddressClient(addressGrpcClient)

	cartGrpcClient := cartPb.NewCartServiceClient(cartConn)
	cartClient := orderGrpcClient.NewCartClient(cartGrpcClient)

	restaurantGrpcClient := restaurantPb.NewRestaurantServiceClient(resConn)
	restaurantClient := orderGrpcClient.NewRestaurantClient(restaurantGrpcClient)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	rabbitClient, err := rabbitmq.NewRabbitClient(cfg.RabbitMQURL, appLogger)
	if err != nil {
		appLogger.Fatal("Failed to connect to RabbitMQ", err)
	}
	defer rabbitClient.Close()

	orderRepo := orderPG.NewOrderRepo(pool)
	orderUC := orderUseCase.NewOrderUseCase(
		orderRepo,
		addressClient,
		cartClient,
		restaurantClient,
		rabbitClient,
		cfg.DefaultRestaurantLogoURL,
		appLogger,
	)
	tracedOrderUC := orderUseCase.NewOrderUseCaseTracingMiddleware(orderUC)
	orderHandler := orderDelivery.NewOrderHandler(tracedOrderUC)

	promoRepo := orderPG.NewPromoRepo(pool)
	promoUC := orderUseCase.NewPromoUseCase(promoRepo)
	promoHandler := orderDelivery.NewPromoHandler(promoUC)

	orderConsumer := orderRabbitMQ.NewOrderConsumer(rabbitClient, orderUC, appLogger)
	if err := orderConsumer.Start(ctx); err != nil {
		appLogger.Fatal("Failed to start RabbitMQ consumer", err)
	}

	advancer := autoadvance.New(orderRepo, rabbitClient, 15*time.Second, appLogger, tracedOrderUC)
	go advancer.Run(ctx)

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

	pb.RegisterOrderServiceServer(grpcServer, orderHandler)
	pb.RegisterPromoServiceServer(grpcServer, promoHandler)
	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", ":"+cfg.GRPC.Port)
	if err != nil {
		appLogger.Fatal("Failed to listen port", err)
	}

	go func() {
		appLogger.Info("Order gRPC server is running", logger.String("port", cfg.GRPC.Port))
		if err := grpcServer.Serve(listener); err != nil {
			appLogger.Fatal("Server failed", err)
		}
	}()

	<-ctx.Done()
	appLogger.Info("Received shutdown signal, stopping gracefully...")
	grpcServer.GracefulStop()
	appLogger.Info("Order microservice stopped")
}

func createGrpcConn(addr, serviceName string, appLogger logger.Logger) *grpc.ClientConn {
	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	if err != nil {
		appLogger.Fatal("Failed to create "+serviceName+" Service client", err)
	}
	appLogger.Info("Connected to "+serviceName+" Service", logger.String("addr", addr))
	return conn
}
