package config

import (
	"fmt"
	"time"
)

// Postgres — подключение одного сервиса к своей БД. Префикс задаёт сервис через
// env-prefix: `env-prefix:"CATALOG_"` даёт CATALOG_POSTGRES_DSN и т.д.
type Postgres struct {
	DSN          string        `env:"POSTGRES_DSN"`
	MaxOpenConns int           `env:"POSTGRES_MAX_OPEN_CONNS" env-default:"10"`
	MaxIdleConns int           `env:"POSTGRES_MAX_IDLE_CONNS" env-default:"10"`
	MaxLifetime  time.Duration `env:"POSTGRES_MAX_LIFETIME" env-default:"5m"`
	MaxIdleTime  time.Duration `env:"POSTGRES_MAX_IDLE_TIME" env-default:"1m"`
}

// Validate проверяет значения; prefix нужен, чтобы ошибка называла переменную из .env.
func (p Postgres) Validate(prefix string) error {
	if err := ValidateNotEmpty(prefix+"POSTGRES_DSN", p.DSN); err != nil {
		return err
	}
	if err := ValidatePositiveInt(prefix+"POSTGRES_MAX_OPEN_CONNS", p.MaxOpenConns); err != nil {
		return err
	}
	if err := ValidatePositiveInt(prefix+"POSTGRES_MAX_IDLE_CONNS", p.MaxIdleConns); err != nil {
		return err
	}
	if p.MaxIdleConns > p.MaxOpenConns {
		return fmt.Errorf("%sPOSTGRES_MAX_IDLE_CONNS must be <= %sPOSTGRES_MAX_OPEN_CONNS, got %d > %d",
			prefix, prefix, p.MaxIdleConns, p.MaxOpenConns)
	}
	if err := ValidatePositiveDuration(prefix+"POSTGRES_MAX_LIFETIME", p.MaxLifetime); err != nil {
		return err
	}

	return ValidatePositiveDuration(prefix+"POSTGRES_MAX_IDLE_TIME", p.MaxIdleTime)
}
