package config

import (
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/config"
)

type YookassaConfig struct {
	ShopID    string `yaml:"shop_id" env:"YOOKASSA_SHOP_ID" env-required:"true"`
	SecretKey string `yaml:"secret_key" env:"YOOKASSA_SECRET_KEY" env-required:"true"`
	ReturnURL string `yaml:"return_url" env:"YOOKASSA_RETURN_URL" env-required:"true"`
}

type Config struct {
	Logger   config.LoggerConfig     `yaml:"logger"`
	Postgres config.PostgresConfig   `yaml:"postgres"`
	GRPC     config.GRPCServerConfig `yaml:"grpc"`
	Redis    config.RedisConfig      `yaml:"redis"`

	OrderServiceAddr string         `yaml:"order_service_addr" env:"ORDER_SERVICE_ADDR" env-default:"localhost:50057"`
	Yookassa         YookassaConfig `yaml:"yookassa"`
}

func Load() *Config {
	var cfg Config
	config.MustLoad(&cfg)
	return &cfg
}
