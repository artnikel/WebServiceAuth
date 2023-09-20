package config

import "github.com/caarlos0/env"

type Variables struct {
	PostgresConnWebAuth string `env:"POSTGRES_CONN_WEBAUTH"`
	RedisWebAddress     string `env:"REDIS_WEB_ADDRESS"`
	TokenSignature      string `env:"TOKEN_SIGNATURE"`
}

func New() (*Variables, error) {
	cfg := &Variables{}
	err := env.Parse(cfg)
	return cfg, err
}
