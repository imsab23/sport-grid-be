package config

import "github.com/imsab23/platform-be/pkg/util/env"

type ServerConfig struct {
	Host string
	Port string
}

func (s *ServerConfig) Addr() string {
	return s.Host + ":" + s.Port
}

func (c *Config) serverConfig() {
	c.Server.Host = env.GetOrDefault("SERVER_HOST", "localhost")
	c.Server.Port = env.GetOrDefault("SERVER_PORT", "9000")
}
