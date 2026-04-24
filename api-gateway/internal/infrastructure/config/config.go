package config

import (
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/config"
)

type GRPCClientsConfig struct {
	AuthAddr       string `yaml:"auth_service_addr" env:"AUTH_SERVICE_ADDR"`
	UserAddr       string `yaml:"user_service_addr" env:"USER_SERVICE_ADDR"`
	RestaurantAddr string `yaml:"restaurant_service_addr" env:"RESTAURANT_SERVICE_ADDR"`
	CartAddr       string `yaml:"cart_service_addr" env:"CART_SERVICE_ADDR"`
	AddressAddr    string `yaml:"address_service_addr" env:"ADDRESS_SERVICE_ADDR"`
	PaymentAddr    string `yaml:"payment_service_addr" env:"PAYMENT_SERVICE_ADDR"`
	OrderAddr      string `yaml:"order_service_addr" env:"ORDER_SERVICE_ADDR"`
}

type Config struct {
	Logger      config.LoggerConfig `yaml:"logger"`
	HTTP        HTTPConfig          `yaml:"http"`
	GRPCClients GRPCClientsConfig   `yaml:"grpc_clients"`
	OTEL        config.OTELConfig   `yaml:"otel"`
}

type HTTPConfig struct {
	Port           string   `yaml:"port" env:"PORT" env-default:"8080"`
	AllowedOrigins []string `yaml:"allowed_origins"`
}

func Load() *Config {
	var cfg Config
	config.MustLoad(&cfg)
	return &cfg
}
