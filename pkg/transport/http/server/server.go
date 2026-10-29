package httpserver

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/requestid"

	"github.com/PaPaSmUrFiK/OuchFix/pkg/logging"
)

func SignalContext() (context.Context, context.CancelFunc) {
	return signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
}

type Config struct {
	Addr            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

type Server struct {
	app *fiber.App
	cfg Config
}

func New(cfg Config, routes func(fiber.Router)) *Server {
	app := fiber.New(fiber.Config{
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	})
	app.Use(requestid.New(), recover.New())
	routes(app)

	return &Server{app: app, cfg: cfg}
}

// Run слушает адрес до отмены ctx, затем мягко гасит сервер.
func (s *Server) Run(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		errCh <- s.app.Listen(s.cfg.Addr, fiber.ListenConfig{DisableStartupMessage: true})
	}()
	logging.Info("http server started", "addr", s.cfg.Addr)

	select {
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("listen on %s: %w", s.cfg.Addr, err)
		}
		return nil
	case <-ctx.Done():
	}

	if err := s.app.ShutdownWithTimeout(s.cfg.ShutdownTimeout); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}

	return <-errCh
}

func Health(c fiber.Ctx) error {
	return c.JSON(fiber.Map{"status": "ok"})
}
