package config

import (
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/config"
)

type ClickHouseConfig struct {
	Host     string `yaml:"host" env:"CLICKHOUSE_HOST" env-default:"localhost"`
	Port     string `yaml:"port" env:"CLICKHOUSE_PORT" env-default:"9000"`
	Database string `yaml:"database" env:"CLICKHOUSE_DB" env-default:"analytics"`
	User     string `yaml:"user" env:"CLICKHOUSE_USER" env-default:"admin"`
	Password string `yaml:"password" env:"CLICKHOUSE_PASSWORD"`
}

type IngesterConfig struct {
	BatchSize     int           `yaml:"batch_size" env:"INGESTER_BATCH_SIZE" env-default:"1000"`
	FlushInterval time.Duration `yaml:"flush_interval" env:"INGESTER_FLUSH_INTERVAL" env-default:"5s"`
}

type Config struct {
	Logger      config.LoggerConfig     `yaml:"logger"`
	ClickHouse  ClickHouseConfig        `yaml:"clickhouse"`
	Ingester    IngesterConfig          `yaml:"ingester"`
	OTEL        config.OTELConfig       `yaml:"otel"`
	RabbitMQURL string                  `yaml:"rabbit_mq_url" env:"RABBITMQ_URL" env-required:"true"`
	GRPC        config.GRPCServerConfig `yaml:"grpc"`
}

func Load() *Config {
	var cfg Config
	config.MustLoad(&cfg)
	return &cfg
}
