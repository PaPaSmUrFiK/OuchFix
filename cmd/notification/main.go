package main

import (
	"fmt"
	_ "time/tzdata" // зоны для APP_TIMEZONE и на Windows без системной базы

	"github.com/PaPaSmUrFiK/OuchFix/internal/notification/app"
	"github.com/PaPaSmUrFiK/OuchFix/internal/notification/config"
	"github.com/PaPaSmUrFiK/OuchFix/internal/notification/container"
	"github.com/PaPaSmUrFiK/OuchFix/pkg/logging"
	httpserver "github.com/PaPaSmUrFiK/OuchFix/pkg/transport/http/server"
)

func main() {
	if err := run(); err != nil {
		logging.Fatal("notification stopped with error", "error", err)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	if err := logging.SetLevel(cfg.LogLevel); err != nil {
		return fmt.Errorf("apply log level: %w", err)
	}

	ctx, stop := httpserver.SignalContext()
	defer stop()

	ctr, err := container.New(ctx, cfg)
	if err != nil {
		return fmt.Errorf("build container: %w", err)
	}
	defer func() {
		if err := ctr.Close(); err != nil {
			logging.Error("close container failed", "error", err)
		}
	}()

	return app.Run(ctx, ctr)
}
