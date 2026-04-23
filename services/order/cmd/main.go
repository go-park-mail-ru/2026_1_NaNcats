package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	infrastructureLogger "github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/common/logger"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/interceptors"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/metrics"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/postgres"

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

	addrConn := createGrpcConn(cfg.AddressServiceAddr, "Address", appLogger)
	defer addrConn.Close()
	cartConn := createGrpcConn(cfg.CartServiceAddr, "Cart", appLogger)
	defer cartConn.Close()
	payConn := createGrpcConn(cfg.PaymentServiceAddr, "Payment", appLogger)
	defer payConn.Close()
	resConn := createGrpcConn(cfg.RestaurantServiceAddr, "Restaurant", appLogger)
	defer resConn.Close()

	addressGrpcClient := addressPb.NewAddressServiceClient(addrConn)
	addressClient := orderGrpcClient.NewAddressClient(addressGrpcClient)

	cartGrpcClient := cartPb.NewCartServiceClient(cartConn)
	cartClient := orderGrpcClient.NewCartClient(cartGrpcClient)

	restaurantGrpcClient := restaurantPb.NewRestaurantServiceClient(resConn)
	restaurantClient := orderGrpcClient.NewRestaurantClient(restaurantGrpcClient)

	orderRepo := orderPG.NewOrderRepo(pool)
	orderUC := orderUseCase.NewOrderUseCase(
		orderRepo,
		addressClient,
		cartClient,
		restaurantClient,
		cfg.DefaultRestaurantLogoURL,
	)
	orderHandler := orderDelivery.NewOrderHandler(orderUC)

	// Контекст, который отменяется по сигналу ОС
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

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

	pb.RegisterOrderServiceServer(grpcServer, orderHandler)
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
	)
	if err != nil {
		appLogger.Fatal("Failed to create "+serviceName+" Service client", err)
	}
	appLogger.Info("Connected to "+serviceName+" Service", logger.String("addr", addr))
	return conn
}
