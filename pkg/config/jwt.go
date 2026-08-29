package config

import "github.com/imsab23/platform-be/pkg/util/env"

type JWTConfig struct {
	PrivateKeyPath string
	PublicKeyPath  string
	KID            string
	Issuer         string
	Audience       string
}

func (c *Config) jwtConfig() {
	c.JWT.PrivateKeyPath = env.GetOrDefault("JWT_PRIVATE_KEY_PATH", "/home/naeem/Documents/Sam/sport-grid-be/keys/jwt_private.pem")
	c.JWT.PublicKeyPath = env.GetOrDefault("JWT_PUBLIC_KEY_PATH", "/home/naeem/Documents/Sam/sport-grid-be/keys/jwt_public.pem")
	c.JWT.KID = env.GetOrDefault("JWT_KID", "sport-grid")
	c.JWT.Issuer = env.GetOrDefault("JWT_ISSUER", "sport-grid")
	c.JWT.Audience = env.GetOrDefault("JWT_AUDIENCE", "sport-grid")
}
