package config

import (
	logger "github.com/imsab23/platform-be/observability/logging"
	"github.com/joho/godotenv"
)

type Config struct {
	PostgresDB PostgresDB
}

func Load() (*Config, error) {
	log, err := logger.NewLogger("backend-service")
	if err != nil {
		return nil, err
	}

	cfg := &Config{}

	err = godotenv.Load()
	if err != nil {
		log.Error("Failed to load env", logger.F("err", err.Error()))
	}

	cfg.postgresDbConfig()

	return cfg, nil
}
