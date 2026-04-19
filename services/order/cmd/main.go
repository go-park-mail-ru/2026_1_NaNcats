package main

import (
	"context"
	"errors"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	infrastructureLogger "github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/common/logger"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/interceptors"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/postgres"

	addressPb "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/address"
	cartPb "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/cart"
	pb "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/order"
	paymentPb "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/payment"
	restaurantPb "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/restaurant"

	orderGrpcClient "github.com/go-park-mail-ru/2026_1_NaNcats/services/order/infrastructure/grpc_client"
	orderDelivery "github.com/go-park-mail-ru/2026_1_NaNcats/services/order/internal/delivery/grpc"
	orderPG "github.com/go-park-mail-ru/2026_1_NaNcats/services/order/internal/repository/postgres"
	orderUseCase "github.com/go-park-mail-ru/2026_1_NaNcats/services/order/internal/usecase"
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
	appLogger.Info("Starting Order microservice (Orchestrator)...")

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

	if err := pool.Ping(ctx); err != nil {
		appLogger.Fatal("could not ping the database", err)
	}
	appLogger.Info("Connected to PostgreSQL")

	addressAddr := getEnv("ADDRESS_SERVICE_ADDR", "localhost:50053")
	cartAddr := getEnv("CART_SERVICE_ADDR", "localhost:50055")
	paymentAddr := getEnv("PAYMENT_SERVICE_ADDR", "localhost:50056")
	restaurantAddr := getEnv("RESTAURANT_SERVICE_ADDR", "localhost:50052")

	addrConn := createGrpcConn(addressAddr, "Address", appLogger)
	defer addrConn.Close()
	cartConn := createGrpcConn(cartAddr, "Cart", appLogger)
	defer cartConn.Close()
	payConn := createGrpcConn(paymentAddr, "Payment", appLogger)
	defer payConn.Close()
	resConn := createGrpcConn(restaurantAddr, "Restaurant", appLogger)
	defer resConn.Close()

	addressGrpcClient := addressPb.NewAddressServiceClient(addrConn)
	addressClient := orderGrpcClient.NewAddressClient(addressGrpcClient)

	cartGrpcClient := cartPb.NewCartServiceClient(cartConn)
	cartClient := orderGrpcClient.NewCartClient(cartGrpcClient)

	paymentGrpcClient := paymentPb.NewPaymentServiceClient(payConn)
	paymentClient := orderGrpcClient.NewPaymentClient(paymentGrpcClient)

	restaurantGrpcClient := restaurantPb.NewRestaurantServiceClient(resConn)
	restaurantClient := orderGrpcClient.NewRestaurantClient(restaurantGrpcClient)

	defaultLogo := os.Getenv("DEFAULT_RESTAURANT_LOGO_URL")

	orderRepo := orderPG.NewOrderRepo(pool)
	orderUC := orderUseCase.NewOrderUseCase(
		orderRepo,
		addressClient,
		cartClient,
		paymentClient,
		restaurantClient,
		defaultLogo,
	)
	orderHandler := orderDelivery.NewOrderHandler(orderUC)

	port := getEnv("GRPC_PORT", "50057")
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptors.UnaryServerRecovery(appLogger),
			interceptors.UnaryServerLogging(appLogger),
		),
	)

	pb.RegisterOrderServiceServer(grpcServer, orderHandler)
	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		appLogger.Fatal("Failed to listen port", err)
	}

	go func() {
		appLogger.Info("Order gRPC server is running", logger.String("port", port))
		if err := grpcServer.Serve(listener); err != nil {
			appLogger.Fatal("Server failed", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	appLogger.Info("Received shutdown signal, stopping gracefully...")
	grpcServer.GracefulStop()
	appLogger.Info("Order microservice stopped")
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
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
