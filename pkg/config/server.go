package config

import (
	"strconv"
	"time"

	httpserver "github.com/PaPaSmUrFiK/OuchFix/pkg/transport/http/server"
)

// Server — HTTP-сервер сервиса. Префикс задаёт сервис через env-prefix:
// `env-prefix:"GATEWAY_"` даёт GATEWAY_PORT, GATEWAY_READ_TIMEOUT и т.д.
type Server struct {
	Port            int           `env:"PORT"`
	ReadTimeout     time.Duration `env:"READ_TIMEOUT" env-default:"15s"`
	WriteTimeout    time.Duration `env:"WRITE_TIMEOUT" env-default:"30s"`
	IdleTimeout     time.Duration `env:"IDLE_TIMEOUT" env-default:"60s"`
	ShutdownTimeout time.Duration `env:"SHUTDOWN_TIMEOUT" env-default:"10s"`
}

func (s Server) Validate(prefix string) error {
	if err := ValidatePort(prefix+"PORT", s.Port); err != nil {
		return err
	}
	for _, t := range []struct {
		name  string
		value time.Duration
	}{
		{prefix + "READ_TIMEOUT", s.ReadTimeout},
		{prefix + "WRITE_TIMEOUT", s.WriteTimeout},
		{prefix + "IDLE_TIMEOUT", s.IdleTimeout},
		{prefix + "SHUTDOWN_TIMEOUT", s.ShutdownTimeout},
	} {
		if err := ValidatePositiveDuration(t.name, t.value); err != nil {
			return err
		}
	}
	return nil
}

func (s Server) HTTP() httpserver.Config {
	return httpserver.Config{
		Addr:            ":" + strconv.Itoa(s.Port),
		ReadTimeout:     s.ReadTimeout,
		WriteTimeout:    s.WriteTimeout,
		IdleTimeout:     s.IdleTimeout,
		ShutdownTimeout: s.ShutdownTimeout,
	}
}
