package config

import (
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type LoggerConfig struct {
	Level string `yaml:"level" env:"LOG_LEVEL" env-default:"info"`
}

type PostgresConfig struct {
	URL string `yaml:"url" env:"DATABASE_URL" env-required:"true"`
}

type GRPCServerConfig struct {
	Port    string `yaml:"port" env:"GRPC_PORT" env-default:"50051"`
	Timeout string `yaml:"timeout" env:"GRPC_TIMEOUT" env-default:"5s"`
}

type RedisConfig struct {
	URL string `yaml:"url" env:"REDIS_URL" env-default:"redis://localhost:6379/0"`
}

// Здесь храним дефолтные слаги
type ErrorConfig struct {
	InternalSlug string `yaml:"internal_slug" env:"ERR_INTERNAL_SLUG" env-default:"INTERNAL_SERVER_ERROR"`
}

// Универсальный загрузчик
func MustLoad(cfg interface{}) {
	// Путь к конфигу можно передать через переменную окружения CONFIG_PATH
	configPath := os.Getenv("CONFIG_PATH")

	if configPath != "" {
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			log.Fatalf("config file does not exist: %s", configPath)
		}
		if err := cleanenv.ReadConfig(configPath, cfg); err != nil {
			log.Fatalf("cannot read config: %v", err)
		}
	} else {
		// Если файла нет, пытаемся прочитать только из ENV
		if err := cleanenv.ReadEnv(cfg); err != nil {
			log.Fatalf("cannot read env: %v", err)
		}
	}
}
