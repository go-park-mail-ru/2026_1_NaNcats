package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/exaring/otelpgx"
	infrastructureLogger "github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/common/logger"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/interceptors"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/metrics"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/postgres"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/rabbitmq/events"
	"github.com/joho/godotenv"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"

	cartDelivery "github.com/go-park-mail-ru/2026_1_NaNcats/services/cart/internal/delivery/grpc"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/cart/internal/infrastructure/config"
	cartGrpcClient "github.com/go-park-mail-ru/2026_1_NaNcats/services/cart/internal/infrastructure/grpc_client"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/cart/internal/infrastructure/outbox"
	cartPG "github.com/go-park-mail-ru/2026_1_NaNcats/services/cart/internal/repository/postgres"
	cartUseCase "github.com/go-park-mail-ru/2026_1_NaNcats/services/cart/internal/usecase"

	cartRabbitMQ "github.com/go-park-mail-ru/2026_1_NaNcats/services/cart/internal/delivery/rabbitmq"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/rabbitmq"

	pb "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/cart"
	restaurantPb "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/restaurant"

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
	appLogger.Info("Starting Cart microservice...")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pgConfig, err := pgxpool.ParseConfig(cfg.Postgres.URL)
	if err != nil {
		appLogger.Fatal("config parsing failed", err)
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

	resConn, err := grpc.NewClient(
		cfg.RestaurantServiceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		appLogger.Fatal("Failed to create Restaurant Service client", err)
	}
	defer resConn.Close()

	restaurantGrpcClient := restaurantPb.NewRestaurantServiceClient(resConn)
	appLogger.Info("Connected to Restaurant Service", logger.String("addr", cfg.RestaurantServiceAddr))

	restaurantClient := cartGrpcClient.NewRestaurantClient(restaurantGrpcClient)

	cartRepo := cartPG.NewCartRepo(pool)
	cartUC := cartUseCase.NewCartUseCase(cartRepo, restaurantClient, cfg.DefaultFoodLogoURL)
	tracedCartUC := cartUseCase.NewCartUseCaseTracingMiddleware(cartUC)
	cartHandler := cartDelivery.NewCartHandler(tracedCartUC)

	rabbitClient, err := rabbitmq.NewRabbitClient(cfg.RabbitMQURL, appLogger)
	if err != nil {
		appLogger.Fatal("Failed to connect to RabbitMQ", err)
	}
	defer rabbitClient.Close()

	cartRelay := outbox.NewRelay(pool, rabbitClient, appLogger, events.QueueGatewayEvents)
	go cartRelay.Run(ctx)

	cartConsumer := cartRabbitMQ.NewCartConsumer(rabbitClient, cartUC, appLogger)
	if err := cartConsumer.Start(ctx); err != nil {
		appLogger.Fatal("Failed to start Cart RabbitMQ consumer", err)
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

	pb.RegisterCartServiceServer(grpcServer, cartHandler)
	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", ":"+cfg.GRPC.Port)
	if err != nil {
		appLogger.Fatal("Failed to listen port", err)
	}

	go func() {
		appLogger.Info("Cart gRPC server is running", logger.String("port", cfg.GRPC.Port))
		if err := grpcServer.Serve(listener); err != nil {
			appLogger.Fatal("Server failed", err)
		}
	}()

	<-ctx.Done()
	appLogger.Info("Received shutdown signal, stopping gracefully...")
	grpcServer.GracefulStop()
	appLogger.Info("Cart microservice stopped")
}
