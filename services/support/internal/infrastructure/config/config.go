package config

import (
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/config"
)

type Config struct {
	Logger      config.LoggerConfig     `yaml:"logger"`
	Postgres    config.PostgresConfig   `yaml:"postgres"`
	GRPC        config.GRPCServerConfig `yaml:"grpc"`
	OTEL        config.OTELConfig       `yaml:"otel"`
	RabbitMQURL string                  `yaml:"rabbit_mq_url" env:"RABBITMQ_URL" env-required:"true"`
}

func Load() *Config {
	var cfg Config
	config.MustLoad(&cfg)
	return &cfg
}
