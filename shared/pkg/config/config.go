package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type LoggerConfig struct {
	Level string `yaml:"level" env:"LOG_LEVEL" env-default:"info"`
}

type PostgresConfig struct {
	URL string `yaml:"url" env:"DATABASE_URL" env-required:"true"`
}

type GRPCServerConfig struct {
	Port    string `yaml:"port" env:"GRPC_PORT" env-required:"true"`
	Timeout string `yaml:"timeout" env:"GRPC_TIMEOUT" env-default:"5s"`
}

type RedisConfig struct {
	URL         string        `yaml:"url" env:"REDIS_URL" env-default:"redis://localhost:6379/0"`
	MaxIdle     int           `yaml:"max_idle" env:"REDIS_MAX_IDLE" env-default:"10"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env:"REDIS_IDLE_TIMEOUT" env-default:"60s"`
}

// Универсальный загрузчик
func MustLoad(cfg interface{}) {
	configPath := os.Getenv("CONFIG_PATH")

	if configPath != "" {
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			log.Fatalf("config file does not exist: %s", configPath)
		}
		if err := cleanenv.ReadConfig(configPath, cfg); err != nil {
			log.Fatalf("cannot read config: %v", err)
		}
	} else {
		if err := cleanenv.ReadEnv(cfg); err != nil {
			log.Fatalf("cannot read env: %v", err)
		}
	}
}
