package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"

	"github.com/go-park-mail-ru/2026_1_NaNcats/services/analytics/internal/infrastructure/config"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/analytics/internal/repository/clickhouse"
	infrastructureLogger "github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/common/logger"
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/logger"
)

func main() {
	_ = godotenv.Load()
	cfg := config.Load()

	rawLogger, err := logger.NewZapLogger(cfg.Logger.Level)
	if err != nil {
		log.Fatalf("cannot start without logger: %v", err)
	}
	appLogger := infrastructureLogger.NewLoggerAdapter(rawLogger)
	appLogger.Info("Starting Analytics Ingester service...")

	chConn, err := clickhouse.NewClickHouseClient(cfg.ClickHouse)
	if err != nil {
		appLogger.Fatal("failed to connect to ClickHouse", err)
	}
	defer chConn.Close()
	appLogger.Info("Connected to ClickHouse", logger.String("addr", cfg.ClickHouse.Host))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()
	appLogger.Info("Shutting down Analytics Ingester...")
}
