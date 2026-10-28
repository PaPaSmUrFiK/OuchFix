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
	Server   Server
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

	server, err := loadServer()
	if err != nil {
		return nil, err
	}
	cfg.Server = server

	return cfg, nil
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
