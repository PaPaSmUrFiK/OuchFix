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
	Server   pkgconfig.Server `env-prefix:"GATEWAY_"`
	Services Services
	// TestUserID — ВРЕМЕННО до ЛР5 (JWT): от чьего имени выполняются
	// myAppointments и myNotifications.
	TestUserID int64 `env:"GATEWAY_TEST_USER_ID" env-default:"5"`
}

// Services — адреса внутренних REST API сервисов. Доступа к БД у gateway нет.
type Services struct {
	CatalogURL      string        `env:"CATALOG_URL" env-default:"http://localhost:8082"`
	AppointmentURL  string        `env:"APPOINTMENT_URL" env-default:"http://localhost:8083"`
	PaymentURL      string        `env:"PAYMENT_URL" env-default:"http://localhost:8084"`
	NotificationURL string        `env:"NOTIFICATION_URL" env-default:"http://localhost:8085"`
	Timeout         time.Duration `env:"GATEWAY_SERVICES_TIMEOUT" env-default:"5s"`
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

	if err := cfg.Server.Validate("GATEWAY_"); err != nil {
		return nil, err
	}
	if err := pkgconfig.ValidatePositiveDuration("GATEWAY_SERVICES_TIMEOUT", cfg.Services.Timeout); err != nil {
		return nil, err
	}
	if cfg.TestUserID <= 0 {
		return nil, fmt.Errorf("GATEWAY_TEST_USER_ID must be positive, got %d", cfg.TestUserID)
	}

	return cfg, nil
}
