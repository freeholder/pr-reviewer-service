package config

import (
	"log"

	"github.com/caarlos0/env/v10"
)

type Config struct {
	HTTPPort string `env:"APP_HTTP_PORT" envDefault:"8080"`
	DBDSN    string `env:"APP_DB_DSN,required"`
}

func MustLoad() Config {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("failed to load config from env: %v", err)
	}
	return cfg
}
