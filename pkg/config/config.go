package config

import (
	"fmt"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

const envFile = ".env"

// ReadEnv подтягивает .env, если он есть (уже выставленные переменные окружения
// не перезаписываются), и заполняет cfg по тегам env/env-default.
func ReadEnv(cfg any) error {
	if _, err := os.Stat(envFile); err == nil {
		if err := godotenv.Load(envFile); err != nil {
			return fmt.Errorf("load %s: %w", envFile, err)
		}
	}

	if err := cleanenv.ReadEnv(cfg); err != nil {
		return fmt.Errorf("read env: %w", err)
	}

	return nil
}
