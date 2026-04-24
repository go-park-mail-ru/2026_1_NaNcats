package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	infrastructureLogger "github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/common/logger"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/interceptors"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/metrics"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/postgres"
	"github.com/joho/godotenv"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"

	restaurantDelivery "github.com/go-park-mail-ru/2026_1_NaNcats/services/restaurant/internal/delivery/grpc"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/restaurant/internal/infrastructure/config"
	restaurantPG "github.com/go-park-mail-ru/2026_1_NaNcats/services/restaurant/internal/repository/postgres"
	restaurantUseCase "github.com/go-park-mail-ru/2026_1_NaNcats/services/restaurant/internal/usecase"
	pb "github.com/go-park-mail-ru/2026_1_NaNcats/shared/proto/restaurant"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
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
	appLogger.Info("Starting Restaurant microservice...")

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

	brandRepo := restaurantPG.NewRestaurantBrandRepo(pool)
	dishRepo := restaurantPG.NewDishRepo(pool)

	brandUC := restaurantUseCase.NewRestaurantBrandUseCase(brandRepo, cfg.DefaultRestaurantLogoURL)
	dishUC := restaurantUseCase.NewDishUseCase(dishRepo, cfg.DefaultFoodLogoURL)

	restaurantHandler := restaurantDelivery.NewRestaurantHandler(brandUC, dishUC)

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

	pb.RegisterRestaurantServiceServer(grpcServer, restaurantHandler)
	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", ":"+cfg.GRPC.Port)
	if err != nil {
		appLogger.Fatal("Failed to listen port", err)
	}

	go func() {
		appLogger.Info("Restaurant gRPC server is running", logger.String("port", cfg.GRPC.Port))
		if err := grpcServer.Serve(listener); err != nil {
			appLogger.Fatal("Server failed", err)
		}
	}()

	<-ctx.Done()
	appLogger.Info("Received shutdown signal, stopping gracefully...")
	grpcServer.GracefulStop()
	appLogger.Info("Restaurant microservice stopped")
}
