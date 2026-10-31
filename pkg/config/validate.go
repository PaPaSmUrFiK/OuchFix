package config

import (
	"fmt"
	"strings"
	"time"
)

// Общие проверки конфигов. Имя переменной окружения передаётся явно, чтобы
// ошибка называла то, что правят в .env, а не поле структуры.

func ValidatePort(name string, value int) error {
	if value < 1 || value > 65535 {
		return fmt.Errorf("%s must be in range 1..65535, got %d", name, value)
	}
	return nil
}

func ValidatePositiveDuration(name string, value time.Duration) error {
	if value <= 0 {
		return fmt.Errorf("%s must be a positive duration, example: 15s, got %s", name, value)
	}
	return nil
}

func ValidatePositiveInt(name string, value int) error {
	if value <= 0 {
		return fmt.Errorf("%s must be a positive integer, got %d", name, value)
	}
	return nil
}

func ValidateNotEmpty(name, value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s must not be empty", name)
	}
	return nil
}
