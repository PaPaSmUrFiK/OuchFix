package config

import (
	"fmt"
	"time"

	pkgconfig "github.com/PaPaSmUrFiK/OuchFix/pkg/config"
)

type Config struct {
	LogLevel string           `env:"LOG_LEVEL" env-default:"info"`
	Server   pkgconfig.Server `env-prefix:"GATEWAY_"`
	Services Services
	// TestUserID — ВРЕМЕННО до ЛР5 (JWT): от чьего имени выполняются
	// myAppointments и myNotifications.
	TestUserID int64 `env:"GATEWAY_TEST_USER_ID" env-default:"5"`
}

// Services — базовые адреса внутренних REST API сервисов, куда gateway
// проксирует запросы. Доступа к БД у gateway нет.
type Services struct {
	IAMBaseURL          string        `env:"IAM_BASE_URL" env-default:"http://localhost:8081"`
	CatalogBaseURL      string        `env:"CATALOG_BASE_URL" env-default:"http://localhost:8082"`
	AppointmentBaseURL  string        `env:"APPOINTMENT_BASE_URL" env-default:"http://localhost:8083"`
	PaymentBaseURL      string        `env:"PAYMENT_BASE_URL" env-default:"http://localhost:8084"`
	NotificationBaseURL string        `env:"NOTIFICATION_BASE_URL" env-default:"http://localhost:8085"`
	Timeout             time.Duration `env:"GATEWAY_SERVICES_TIMEOUT" env-default:"5s"`
}

func Load() (*Config, error) {
	cfg := &Config{}
	if err := pkgconfig.ReadEnv(cfg); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

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
