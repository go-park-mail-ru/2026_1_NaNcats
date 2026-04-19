package main

import (
	"context"
	"errors"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	infrastructureLogger "github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/common/logger"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/interceptors"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/postgres"
	"github.com/joho/godotenv"

	cartDelivery "github.com/go-park-mail-ru/2026_1_NaNcats/services/cart/internal/delivery/grpc"
	cartPG "github.com/go-park-mail-ru/2026_1_NaNcats/services/cart/internal/repository/postgres"
	cartUseCase "github.com/go-park-mail-ru/2026_1_NaNcats/services/cart/internal/usecase"

	pb "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/cart"
	restaurantPb "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/restaurant"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
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
	appLogger.Info("Starting Cart microservice...")

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

	restaurantAddr := os.Getenv("RESTAURANT_SERVICE_ADDR")
	if restaurantAddr == "" {
		restaurantAddr = "localhost:50052"
	}

	resConn, err := grpc.NewClient(
		restaurantAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		appLogger.Fatal("Failed to create Restaurant Service client", err)
	}
	defer resConn.Close()

	restaurantClient := restaurantPb.NewRestaurantServiceClient(resConn)
	appLogger.Info("Connected to Restaurant Service", logger.String("addr", restaurantAddr))

	defaultFoodLogo := os.Getenv("DEFAULT_FOOD_LOGO_URL")

	cartRepo := cartPG.NewCartRepo(pool)
	cartUC := cartUseCase.NewCartUseCase(cartRepo, restaurantClient, defaultFoodLogo)
	cartHandler := cartDelivery.NewCartHandler(cartUC)

	port := os.Getenv("GRPC_PORT")
	if port == "" {
		port = "50055"
	}

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptors.UnaryServerRecovery(appLogger),
			interceptors.UnaryServerLogging(appLogger),
		),
	)

	pb.RegisterCartServiceServer(grpcServer, cartHandler)
	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		appLogger.Fatal("Failed to listen port", err)
	}

	go func() {
		appLogger.Info("Cart gRPC server is running", logger.String("port", port))
		if err := grpcServer.Serve(listener); err != nil {
			appLogger.Fatal("Server failed", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	appLogger.Info("Received shutdown signal, stopping gracefully...")
	grpcServer.GracefulStop()
	appLogger.Info("Cart microservice stopped")
}
