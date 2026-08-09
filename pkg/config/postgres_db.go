package config

import "github.com/imsab23/platform-be/pkg/util/env"

type PostgresDB struct {
	Host     string
	Port     int
	Database string
	Username string
	Password string
	SSLMode  string
}

func (c *Config) postgresDbConfig() {
	c.PostgresDB.Host = env.GetOrDefault("POSTGRES_HOST", "localhost")
	c.PostgresDB.Port = env.GetInt("POSTGRES_PORT", 5432)
	c.PostgresDB.Database = env.GetOrDefault("POSTGRES_DB", "postgres")
	c.PostgresDB.Username = env.GetOrDefault("POSTGRES_USER", "postgres")
	c.PostgresDB.Password = env.GetOrDefault("POSTGRES_PASSWORD", "password")
	c.PostgresDB.SSLMode = env.GetOrDefault("POSTGRES_SSL_MODE", "disable")
}
