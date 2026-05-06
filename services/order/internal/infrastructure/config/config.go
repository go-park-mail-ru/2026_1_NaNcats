package config

import (
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/config"
)

type Config struct {
	Logger   config.LoggerConfig     `yaml:"logger"`
	Postgres config.PostgresConfig   `yaml:"postgres"`
	GRPC     config.GRPCServerConfig `yaml:"grpc"`
	OTEL     config.OTELConfig       `yaml:"otel"`

	AddressServiceAddr    string `yaml:"address_service_addr" env:"ADDRESS_SERVICE_ADDR" env-default:"localhost:50053"`
	CartServiceAddr       string `yaml:"cart_service_addr" env:"CART_SERVICE_ADDR" env-default:"localhost:50055"`
	PaymentServiceAddr    string `yaml:"payment_service_addr" env:"PAYMENT_SERVICE_ADDR" env-default:"localhost:50056"`
	RestaurantServiceAddr string `yaml:"restaurant_service_addr" env:"RESTAURANT_SERVICE_ADDR" env-default:"localhost:50052"`

	RabbitMQURL              string `yaml:"rabbit_mq_url" env:"RABBITMQ_URL" env-required:"true"`
	DefaultRestaurantLogoURL string `yaml:"default_restaurant_logo_url" env:"DEFAULT_RESTAURANT_LOGO_URL" env-required:"true"`
}

func Load() *Config {
	var cfg Config
	config.MustLoad(&cfg)
	return &cfg
}
