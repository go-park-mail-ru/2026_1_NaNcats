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

	restaurantDelivery "github.com/go-park-mail-ru/2026_1_NaNcats/services/restaurant/internal/delivery/grpc"
	restaurantPG "github.com/go-park-mail-ru/2026_1_NaNcats/services/restaurant/internal/repository/postgres"
	restaurantUseCase "github.com/go-park-mail-ru/2026_1_NaNcats/services/restaurant/internal/usecase"
	pb "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/restaurant"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
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
	appLogger.Info("Starting Restaurant microservice...")

	port := os.Getenv("GRPC_PORT")
	if port == "" {
		port = "50053"
	}

	ctx := context.Background()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		appLogger.Fatal("database connection string is missing", errors.New("DATABASE_URL env var is empty"))
	}

	config, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		appLogger.Fatal("config parsing failed", err)
	}

	config.ConnConfig.Tracer = postgres.NewDBTracer(appLogger)

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		appLogger.Fatal("database pool creation failed", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		appLogger.Fatal("could not ping the database", err)
	}
	appLogger.Info("Connected to PostgreSQL")

	defaultRestaurantLogo := os.Getenv("DEFAULT_RESTAURANT_LOGO_URL")
	defaultFoodLogo := os.Getenv("DEFAULT_FOOD_LOGO_URL")
	if defaultRestaurantLogo == "" || defaultFoodLogo == "" {
		appLogger.Warn("Default logo URLs are not set in environment variables")
	}

	brandRepo := restaurantPG.NewRestaurantBrandRepo(pool)
	dishRepo := restaurantPG.NewDishRepo(pool)

	brandUC := restaurantUseCase.NewRestaurantBrandUseCase(brandRepo, defaultRestaurantLogo)
	dishUC := restaurantUseCase.NewDishUseCase(dishRepo, defaultFoodLogo)

	restaurantHandler := restaurantDelivery.NewRestaurantHandler(brandUC, dishUC)

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptors.UnaryServerRecovery(appLogger),
			interceptors.UnaryServerLogging(appLogger),
		),
	)

	pb.RegisterRestaurantServiceServer(grpcServer, restaurantHandler)

	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		appLogger.Fatal("Failed to listen port", err)
	}

	go func() {
		appLogger.Info("Restaurant gRPC server is running", logger.String("port", port))
		if err := grpcServer.Serve(listener); err != nil {
			appLogger.Fatal("Server failed to start", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	appLogger.Info("Received shutdown signal, stopping gracefully...")
	grpcServer.GracefulStop()
	appLogger.Info("Restaurant microservice stopped")
}
