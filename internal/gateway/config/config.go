package config

import (
	"fmt"
	"time"

	pkgconfig "github.com/PaPaSmUrFiK/OuchFix/pkg/config"
	httpserver "github.com/PaPaSmUrFiK/OuchFix/pkg/transport/http/server"
)

type Config struct {
	LogLevel string
	// Location — часовой пояс медцентра: в нём отдаётся время в ответах.
	Location *time.Location
	Server   httpserver.Config
	Services Services
	// TestUserID — ВРЕМЕННО до ЛР5 (JWT): от чьего имени выполняются
	// myAppointments и myNotifications.
	TestUserID int64
}

// Services — адреса внутренних REST API сервисов. Доступа к БД у gateway нет.
type Services struct {
	CatalogURL      string
	AppointmentURL  string
	PaymentURL      string
	NotificationURL string
	Timeout         time.Duration
}

func Load() (*Config, error) {
	if err := pkgconfig.LoadEnv(); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	cfg := &Config{LogLevel: pkgconfig.String("LOG_LEVEL", "info")}

	timezone := pkgconfig.String("APP_TIMEZONE", "Europe/Minsk")
	location, err := time.LoadLocation(timezone)
	if err != nil {
		return nil, fmt.Errorf("APP_TIMEZONE must be an IANA zone, got %q: %w", timezone, err)
	}
	cfg.Location = location

	if cfg.Server, err = httpserver.LoadConfig("GATEWAY", 8080); err != nil {
		return nil, err
	}
	if cfg.Services, err = loadServices(); err != nil {
		return nil, err
	}

	if cfg.TestUserID, err = pkgconfig.Int64("GATEWAY_TEST_USER_ID", 5); err != nil {
		return nil, err
	}
	if cfg.TestUserID <= 0 {
		return nil, fmt.Errorf("GATEWAY_TEST_USER_ID must be positive, got %d", cfg.TestUserID)
	}

	return cfg, nil
}

func loadServices() (Services, error) {
	s := Services{
		CatalogURL:      pkgconfig.String("CATALOG_URL", "http://localhost:8082"),
		AppointmentURL:  pkgconfig.String("APPOINTMENT_URL", "http://localhost:8083"),
		PaymentURL:      pkgconfig.String("PAYMENT_URL", "http://localhost:8084"),
		NotificationURL: pkgconfig.String("NOTIFICATION_URL", "http://localhost:8085"),
	}

	var err error
	if s.Timeout, err = pkgconfig.Duration("GATEWAY_SERVICES_TIMEOUT", 5*time.Second); err != nil {
		return s, err
	}

	return s, pkgconfig.ValidatePositiveDuration("GATEWAY_SERVICES_TIMEOUT", s.Timeout)
}
