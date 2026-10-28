package config

import (
	"fmt"
	"time"

	pkgconfig "github.com/PaPaSmUrFiK/OuchFix/pkg/config"
	httpserver "github.com/PaPaSmUrFiK/OuchFix/pkg/transport/http/server"
)

type Config struct {
	LogLevel string
	// Location — часовой пояс медцентра: в нём считаются границы дня (freeSlots).
	Location *time.Location
	Server   httpserver.Config
	// DB — подключение под своим пользователем Postgres с правами только на схему appointment.
	DB pkgconfig.Postgres
}

func Load() (*Config, error) {
	if err := pkgconfig.LoadEnv(); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	cfg := &Config{LogLevel: pkgconfig.String("LOG_LEVEL", "info")}
	var err error

	timezone := pkgconfig.String("APP_TIMEZONE", "Europe/Minsk")
	if cfg.Location, err = time.LoadLocation(timezone); err != nil {
		return nil, fmt.Errorf("APP_TIMEZONE must be an IANA zone, got %q: %w", timezone, err)
	}
	if cfg.Server, err = httpserver.LoadConfig("APPOINTMENT", 8083); err != nil {
		return nil, err
	}
	if cfg.DB, err = pkgconfig.LoadPostgres("APPOINTMENT"); err != nil {
		return nil, err
	}

	return cfg, nil
}
