package clickhouse

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/go-park-mail-ru/2026_1_NaNcats/services/analytics/internal/infrastructure/config"
)

// Создает новое соединение с ClickHouse по нативному протоколу
func NewClickHouseClient(cfg config.ClickHouseConfig) (driver.Conn, error) {
	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)

	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{addr},
		Auth: clickhouse.Auth{
			Database: cfg.Database,
			Username: cfg.User,
			Password: cfg.Password,
		},
		// Настройки подключения
		DialContext: func(ctx context.Context, addr string) (net.Conn, error) {
			var d net.Dialer
			return d.DialContext(ctx, "tcp", addr)
		},
		// Включаем сжатие LZ4 для экономии трафика при больших батчах
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},
		// Таймаут
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
		},
		ConnMaxLifetime: time.Duration(10) * time.Minute,
		MaxOpenConns:    10,
		MaxIdleConns:    5,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to open clickhouse connection: %w", err)
	}

	// Проверяем доступность базы сразу при старте
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := conn.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping clickhouse: %w", err)
	}

	return conn, nil
}
