package httpserver

import (
	"strconv"
	"time"

	pkgconfig "github.com/PaPaSmUrFiK/OuchFix/pkg/config"
)

// LoadConfig читает настройки HTTP-сервера сервиса с префиксом: GATEWAY_PORT,
// CATALOG_READ_TIMEOUT и т.д.
func LoadConfig(prefix string, defPort int) (Config, error) {
	var (
		c    Config
		port int
		err  error
	)
	if port, err = pkgconfig.Int(prefix+"_PORT", defPort); err != nil {
		return c, err
	}
	if err = pkgconfig.ValidatePort(prefix+"_PORT", port); err != nil {
		return c, err
	}
	c.Addr = ":" + strconv.Itoa(port)

	for _, d := range []struct {
		target *time.Duration
		name   string
		def    time.Duration
	}{
		{&c.ReadTimeout, prefix + "_READ_TIMEOUT", 15 * time.Second},
		{&c.WriteTimeout, prefix + "_WRITE_TIMEOUT", 30 * time.Second},
		{&c.IdleTimeout, prefix + "_IDLE_TIMEOUT", 60 * time.Second},
		{&c.ShutdownTimeout, prefix + "_SHUTDOWN_TIMEOUT", 10 * time.Second},
	} {
		if *d.target, err = pkgconfig.Duration(d.name, d.def); err != nil {
			return c, err
		}
		if err = pkgconfig.ValidatePositiveDuration(d.name, *d.target); err != nil {
			return c, err
		}
	}

	return c, nil
}
