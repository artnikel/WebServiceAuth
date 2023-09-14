package config

import "github.com/caarlos0/env"

type Variables struct {
	PostgresConnWebAuth string `env:"POSTGRES_CONN_WEBAUTH"`
}

func New() (*Variables, error) {
	cfg := &Variables{}
	err := env.Parse(cfg)
	return cfg, err
}