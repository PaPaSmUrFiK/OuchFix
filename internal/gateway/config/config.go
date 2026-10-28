package config

import (
	"fmt"
	"strconv"
	"time"

	pkgconfig "github.com/PaPaSmUrFiK/OuchFix/pkg/config"
	httpserver "github.com/PaPaSmUrFiK/OuchFix/pkg/transport/http/server"
)

type Config struct {
	LogLevel string
	// Location — часовой пояс медцентра: в нём считаются границы дня (freeSlots).
	Location *time.Location
	Server   Server
	DB       Databases
}

// Databases — у каждого сервиса своя БД. В ЛР3 все сервисы живут внутри
// процесса gateway, но подключения уже раздельные, как будет после ЛР6.
type Databases struct {
	IAM          pkgconfig.Postgres
	Catalog      pkgconfig.Postgres
	Appointment  pkgconfig.Postgres
	Payment      pkgconfig.Postgres
	Notification pkgconfig.Postgres
}

type Server struct {
	Port            int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
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

	server, err := loadServer()
	if err != nil {
		return nil, err
	}
	cfg.Server = server

	if cfg.DB, err = loadDatabases(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func loadDatabases() (Databases, error) {
	var (
		d   Databases
		err error
	)
	if d.IAM, err = pkgconfig.LoadPostgres("IAM"); err != nil {
		return d, err
	}
	if d.Catalog, err = pkgconfig.LoadPostgres("CATALOG"); err != nil {
		return d, err
	}
	if d.Appointment, err = pkgconfig.LoadPostgres("APPOINTMENT"); err != nil {
		return d, err
	}
	if d.Payment, err = pkgconfig.LoadPostgres("PAYMENT"); err != nil {
		return d, err
	}
	if d.Notification, err = pkgconfig.LoadPostgres("NOTIFICATION"); err != nil {
		return d, err
	}
	return d, nil
}

func loadServer() (Server, error) {
	var (
		s   Server
		err error
	)
	if s.Port, err = pkgconfig.Int("GATEWAY_PORT", 8080); err != nil {
		return s, err
	}
	if s.ReadTimeout, err = pkgconfig.Duration("GATEWAY_READ_TIMEOUT", 15*time.Second); err != nil {
		return s, err
	}
	if s.WriteTimeout, err = pkgconfig.Duration("GATEWAY_WRITE_TIMEOUT", 30*time.Second); err != nil {
		return s, err
	}
	if s.IdleTimeout, err = pkgconfig.Duration("GATEWAY_IDLE_TIMEOUT", 60*time.Second); err != nil {
		return s, err
	}
	if s.ShutdownTimeout, err = pkgconfig.Duration("GATEWAY_SHUTDOWN_TIMEOUT", 10*time.Second); err != nil {
		return s, err
	}

	return s, s.validate()
}

func (s Server) validate() error {
	if err := pkgconfig.ValidatePort("GATEWAY_PORT", s.Port); err != nil {
		return err
	}
	for _, t := range []struct {
		name  string
		value time.Duration
	}{
		{"GATEWAY_READ_TIMEOUT", s.ReadTimeout},
		{"GATEWAY_WRITE_TIMEOUT", s.WriteTimeout},
		{"GATEWAY_IDLE_TIMEOUT", s.IdleTimeout},
		{"GATEWAY_SHUTDOWN_TIMEOUT", s.ShutdownTimeout},
	} {
		if err := pkgconfig.ValidatePositiveDuration(t.name, t.value); err != nil {
			return err
		}
	}
	return nil
}

func (c *Config) ServerConfig() httpserver.Config {
	return httpserver.Config{
		Addr:            ":" + strconv.Itoa(c.Server.Port),
		ReadTimeout:     c.Server.ReadTimeout,
		WriteTimeout:    c.Server.WriteTimeout,
		IdleTimeout:     c.Server.IdleTimeout,
		ShutdownTimeout: c.Server.ShutdownTimeout,
	}
}
