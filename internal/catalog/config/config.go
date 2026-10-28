package config

import (
	"fmt"

	pkgconfig "github.com/PaPaSmUrFiK/OuchFix/pkg/config"
)

type Config struct {
	LogLevel string           `env:"LOG_LEVEL" env-default:"info"`
	Server   pkgconfig.Server `env-prefix:"CATALOG_"`
	// DB — подключение под своим пользователем Postgres с правами только на схему catalog.
	DB pkgconfig.Postgres `env-prefix:"CATALOG_"`
}

func Load() (*Config, error) {
	cfg := &Config{}
	if err := pkgconfig.ReadEnv(cfg); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	if err := cfg.Server.Validate("CATALOG_"); err != nil {
		return nil, err
	}
	if err := cfg.DB.Validate("CATALOG_"); err != nil {
		return nil, err
	}

	return cfg, nil
}
