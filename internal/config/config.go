package config

import (
	"log"

	"github.com/caarlos0/env/v10"
	"github.com/joho/godotenv"
)

type Config struct {
	HTTPPort string `env:"APP_HTTP_PORT" envDefault:"8080"`
	DBDSN    string `env:"APP_DB_DSN,required"`
}

func MustLoad() Config {
	var cfg Config
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("failed to load config from env: %v", err)
	}
	return cfg
}
