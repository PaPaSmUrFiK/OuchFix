package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

const envFile = ".env"

// LoadEnv подтягивает .env, если он есть. Уже выставленные переменные окружения
// не перезаписываются.
func LoadEnv() error {
	if _, err := os.Stat(envFile); err == nil {
		if err := godotenv.Load(envFile); err != nil {
			return fmt.Errorf("load %s: %w", envFile, err)
		}
	}

	return nil
}

func String(name, def string) string {
	if value, ok := os.LookupEnv(name); ok && strings.TrimSpace(value) != "" {
		return strings.TrimSpace(value)
	}

	return def
}

func Int(name string, def int) (int, error) {
	raw := String(name, "")
	if raw == "" {
		return def, nil
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer, got %q: %w", name, raw, err)
	}

	return value, nil
}

func Int64(name string, def int64) (int64, error) {
	raw := String(name, "")
	if raw == "" {
		return def, nil
	}

	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer, got %q: %w", name, raw, err)
	}

	return value, nil
}

func Duration(name string, def time.Duration) (time.Duration, error) {
	raw := String(name, "")
	if raw == "" {
		return def, nil
	}

	value, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("%s must be a duration, example: 15s, got %q: %w", name, raw, err)
	}

	return value, nil
}

func Bool(name string, def bool) (bool, error) {
	raw := String(name, "")
	if raw == "" {
		return def, nil
	}

	value, err := strconv.ParseBool(raw)
	if err != nil {
		return false, fmt.Errorf("%s must be a boolean, got %q: %w", name, raw, err)
	}

	return value, nil
}
