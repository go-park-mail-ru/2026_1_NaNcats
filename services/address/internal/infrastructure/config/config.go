package config

import (
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/config"
)

type Config struct {
	Logger   config.LoggerConfig     `yaml:"logger"`
	Postgres config.PostgresConfig   `yaml:"postgres"`
	GRPC     config.GRPCServerConfig `yaml:"grpc"`
	OTEL     config.OTELConfig       `yaml:"otel"`

	DefaultFoodLogoURL string `yaml:"default_food_logo_url" env:"DEFAULT_FOOD_LOGO_URL" env-required:"true"`
}

func Load() *Config {
	var cfg Config

	config.MustLoad(&cfg)

	return &cfg
}
