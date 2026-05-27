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
		// Настройки для высокой производительности и надежности
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
			// Включаем асинхронные вставки на стороне сервера
			"async_insert": 1,

			// Ждем, пока ClickHouse подтвердит, что данные записаны в его буфер
			// RabbitMQ Consumer сделает Ack только когда CH принял данные
			"wait_for_async_insert": 1,

			// Timeout сброса буфера в CH
			"async_insert_busy_timeout_ms": 5000,

			// Лимит по количеству запросов
			"async_insert_max_query_number": 1000,

			// Лимит на объем данных
			"async_insert_max_data_size": 10485760,
		},
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},
		ConnMaxLifetime: time.Duration(10) * time.Minute,
		MaxOpenConns:    10,
		MaxIdleConns:    5,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to open clickhouse connection: %w", err)
	}

	// Проверка связи
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := conn.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping clickhouse: %w", err)
	}

	return conn, nil
}
