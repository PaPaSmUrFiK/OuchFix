package config

import (
	"fmt"
	"time"
)

// Postgres — подключение одного сервиса к своей БД. У каждого сервиса свой DSN,
// префикс переменных задаёт сервис: CATALOG_POSTGRES_DSN и т.д.
type Postgres struct {
	DSN          string
	MaxOpenConns int
	MaxIdleConns int
	MaxLifetime  time.Duration
	MaxIdleTime  time.Duration
}

func LoadPostgres(prefix string) (Postgres, error) {
	p := Postgres{DSN: String(prefix+"_POSTGRES_DSN", "")}

	var err error
	if p.MaxOpenConns, err = Int(prefix+"_POSTGRES_MAX_OPEN_CONNS", 10); err != nil {
		return p, err
	}
	if p.MaxIdleConns, err = Int(prefix+"_POSTGRES_MAX_IDLE_CONNS", 10); err != nil {
		return p, err
	}
	if p.MaxLifetime, err = Duration(prefix+"_POSTGRES_MAX_LIFETIME", 5*time.Minute); err != nil {
		return p, err
	}
	if p.MaxIdleTime, err = Duration(prefix+"_POSTGRES_MAX_IDLE_TIME", time.Minute); err != nil {
		return p, err
	}

	return p, p.validate(prefix)
}

func (p Postgres) validate(prefix string) error {
	if err := ValidateNotEmpty(prefix+"_POSTGRES_DSN", p.DSN); err != nil {
		return err
	}
	if err := ValidatePositiveInt(prefix+"_POSTGRES_MAX_OPEN_CONNS", p.MaxOpenConns); err != nil {
		return err
	}
	if err := ValidatePositiveInt(prefix+"_POSTGRES_MAX_IDLE_CONNS", p.MaxIdleConns); err != nil {
		return err
	}
	if p.MaxIdleConns > p.MaxOpenConns {
		return fmt.Errorf("%s_POSTGRES_MAX_IDLE_CONNS must be <= %s_POSTGRES_MAX_OPEN_CONNS, got %d > %d",
			prefix, prefix, p.MaxIdleConns, p.MaxOpenConns)
	}
	if err := ValidatePositiveDuration(prefix+"_POSTGRES_MAX_LIFETIME", p.MaxLifetime); err != nil {
		return err
	}

	return ValidatePositiveDuration(prefix+"_POSTGRES_MAX_IDLE_TIME", p.MaxIdleTime)
}
