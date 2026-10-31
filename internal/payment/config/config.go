package config

import (
	"fmt"
	"time"

	pkgconfig "github.com/PaPaSmUrFiK/OuchFix/pkg/config"
)

type Config struct {
	LogLevel string `env:"LOG_LEVEL" env-default:"info"`
	// Timezone — часовой пояс медцентра: в нём отдаётся время в ответах.
	Timezone string           `env:"APP_TIMEZONE" env-default:"Europe/Minsk"`
	Location *time.Location   `env:"-"`
	Server   pkgconfig.Server `env-prefix:"PAYMENT_"`
	// DB — подключение под своим пользователем Postgres с правами только на схему payment.
	DB pkgconfig.Postgres `env-prefix:"PAYMENT_"`
}

func Load() (*Config, error) {
	cfg := &Config{}
	if err := pkgconfig.ReadEnv(cfg); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	location, err := time.LoadLocation(cfg.Timezone)
	if err != nil {
		return nil, fmt.Errorf("APP_TIMEZONE must be an IANA zone, got %q: %w", cfg.Timezone, err)
	}
	cfg.Location = location

	if err := cfg.Server.Validate("PAYMENT_"); err != nil {
		return nil, err
	}
	if err := cfg.DB.Validate("PAYMENT_"); err != nil {
		return nil, err
	}

	return cfg, nil
}
