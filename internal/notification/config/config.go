package config

import (
	"fmt"

	pkgconfig "github.com/PaPaSmUrFiK/OuchFix/pkg/config"
	httpserver "github.com/PaPaSmUrFiK/OuchFix/pkg/transport/http/server"
)

type Config struct {
	LogLevel string
	Server   httpserver.Config
	// DB — подключение под своим пользователем Postgres с правами только на схему notification.
	DB pkgconfig.Postgres
}

func Load() (*Config, error) {
	if err := pkgconfig.LoadEnv(); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	cfg := &Config{LogLevel: pkgconfig.String("LOG_LEVEL", "info")}
	var err error
	if cfg.Server, err = httpserver.LoadConfig("NOTIFICATION", 8085); err != nil {
		return nil, err
	}
	if cfg.DB, err = pkgconfig.LoadPostgres("NOTIFICATION"); err != nil {
		return nil, err
	}

	return cfg, nil
}
