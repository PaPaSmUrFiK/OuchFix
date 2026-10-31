package logging

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
)

var (
	Global = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: &level}))
	level  slog.LevelVar
)

func Fatal(msg string, args ...any) {
	Global.Error(msg, args...)
	os.Exit(1)
}

func SetLevel(name string) error {
	parsed, err := ParseLevel(name)
	if err != nil {
		return err
	}
	level.Set(parsed)
	return nil
}

func ParseLevel(name string) (slog.Level, error) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "debug":
		return slog.LevelDebug, nil
	case "info", "":
		return slog.LevelInfo, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, fmt.Errorf("unknown log level %q, want one of debug/info/warn/error", name)
	}
}

func Info(msg string, args ...any)  { Global.Info(msg, args...) }
func Error(msg string, args ...any) { Global.Error(msg, args...) }
func Warn(msg string, args ...any)  { Global.Warn(msg, args...) }
func Debug(msg string, args ...any) { Global.Debug(msg, args...) }
