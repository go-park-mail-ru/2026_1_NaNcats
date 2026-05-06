package config

import (
	"time"

	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/config"
)

type Config struct {
	Logger config.LoggerConfig     `yaml:"logger"`
	GRPC   config.GRPCServerConfig `yaml:"grpc"`
	Redis  config.RedisConfig      `yaml:"redis"`
	OTEL   config.OTELConfig       `yaml:"otel"`

	// Адрес gRPC сервера пользователей
	UserServiceAddr string `yaml:"user_service_addr" env:"USER_SERVICE_ADDR" env-default:"localhost:50052"`
	// Время жизни сессии
	SessionTTL time.Duration `yaml:"session_ttl" env:"SESSION_TTL" env-default:"24h"`
}

func Load() *Config {
	var cfg Config
	config.MustLoad(&cfg)
	return &cfg
}
