package config

import (
	"github.com/go-park-mail-ru/2026_1_NaNcats/shared/pkg/config"
)

type S3Config struct {
	KeyID      string `yaml:"key_id" env:"S3_KEY_ID" env-required:"true"`
	SecretKey  string `yaml:"secret_key" env:"S3_SECRET_KEY" env-required:"true"`
	BucketName string `yaml:"bucket_name" env:"S3_BUCKET_NAME" env-default:"nancats-bucket"`
	Region     string `yaml:"region" env:"S3_REGION" env-default:"ru-central1"`
}

type Config struct {
	Logger   config.LoggerConfig     `yaml:"logger"`
	Postgres config.PostgresConfig   `yaml:"postgres"`
	GRPC     config.GRPCServerConfig `yaml:"grpc"`
	OTEL     config.OTELConfig       `yaml:"otel"`
	S3       S3Config                `yaml:"s3"`

	DefaultRestaurantLogoURL string `yaml:"default_restaurant_logo_url" env:"DEFAULT_RESTAURANT_LOGO_URL" env-required:"true"`
	DefaultFoodLogoURL       string `yaml:"default_food_logo_url" env:"DEFAULT_FOOD_LOGO_URL" env-required:"true"`

	OrderServiceAddr string `yaml:"order_service_addr" env:"ORDER_SERVICE_ADDR" env-default:"localhost:50057"`
}

func Load() *Config {
	var cfg Config
	config.MustLoad(&cfg)
	return &cfg
}
