package config

import (
	"os"
	"path/filepath"

	logger "github.com/imsab23/platform-be/observability/logging"
	"github.com/joho/godotenv"
)

type Config struct {
	PostgresDB PostgresDB
	JWT        JWTConfig
	Server     ServerConfig
}

func Load() (*Config, error) {
	log, err := logger.NewLogger("backend-service")
	if err != nil {
		return nil, err
	}

	wd, _ := os.Getwd()
	log.Info("Current working directory", logger.F("path", wd))

	envPath := filepath.Join(wd, ".env")
	log.Info(".env path", logger.F("path", envPath))

	if err := godotenv.Load(envPath); err != nil {
		log.Error("Failed to load env", logger.F("err", err.Error()))
	}

	cfg := &Config{}

	cfg.postgresDbConfig()
	cfg.jwtConfig()
	cfg.serverConfig()

	return cfg, nil
}
