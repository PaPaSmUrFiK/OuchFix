package unit

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor"

	pkgerrors "github.com/PaPaSmUrFiK/OuchFix/pkg/errors"
	httpclient "github.com/PaPaSmUrFiK/OuchFix/pkg/transport/http/client"
	httpserver "github.com/PaPaSmUrFiK/OuchFix/pkg/transport/http/server"
)

// Ошибка сервиса проходит через REST с тем же текстом и тем же доменным видом,
// а внутренние ошибки наружу не утекают.
func TestHTTPClientKeepsDomainErrors(t *testing.T) {
	app := fiber.New()
	app.Get("/nf", func(c fiber.Ctx) error {
		return httpserver.WriteError(c, pkgerrors.ErrNotFound)
	})
	app.Get("/bad/:id", func(c fiber.Ctx) error {
		_, err := httpserver.ParamID(c, "id")
		return httpserver.WriteError(c, err)
	})
	app.Get("/boom", func(c fiber.Ctx) error {
		return httpserver.WriteError(c, errors.New("pq: secret sql"))
	})
	app.Get("/ok", func(c fiber.Ctx) error {
		return httpserver.Reply(c, []int{1, 2}, nil)
	})
	srv := httptest.NewServer(adaptorHandler(app))
	defer srv.Close()

	cl := httpclient.New(srv.URL, time.Second)
	ctx := context.Background()
	var out []int

	err := cl.Get(ctx, "/nf", nil, &out)
	if !errors.Is(err, pkgerrors.ErrNotFound) || err.Error() != pkgerrors.ErrNotFound.Error() {
		t.Fatalf("not found: %v", err)
	}
	err = cl.Get(ctx, "/bad/x", nil, &out)
	if !errors.Is(err, pkgerrors.ErrInvalidArgument) || err.Error() != `некорректный аргумент: id должен быть числом, получено "x"` {
		t.Fatalf("bad input: %v", err)
	}
	err = cl.Get(ctx, "/boom", nil, &out)
	if err == nil || errors.Is(err, pkgerrors.ErrNotFound) || errors.Is(err, pkgerrors.ErrInvalidArgument) || strings.Contains(err.Error(), "secret") {
		t.Fatalf("internal: %v", err)
	}
	if err := cl.Get(ctx, "/ok", nil, &out); err != nil || len(out) != 2 {
		t.Fatalf("ok: %v %v", out, err)
	}
}

// POST передаёт тело как JSON, а 409 и 422 сервиса становятся
// ErrConflict и ErrFailedPrecondition с исходным текстом.
func TestHTTPClientPostAndBusinessErrors(t *testing.T) {
	app := fiber.New()
	app.Post("/echo", func(c fiber.Ctx) error {
		var in struct {
			Name string `json:"name"`
		}
		if err := httpserver.BindJSON(c, &in); err != nil {
			return httpserver.WriteError(c, err)
		}
		return httpserver.Reply(c, in, nil)
	})
	app.Post("/conflict", func(c fiber.Ctx) error {
		return httpserver.WriteError(c, fmt.Errorf("%w: слот занят", pkgerrors.ErrConflict))
	})
	app.Post("/state", func(c fiber.Ctx) error {
		return httpserver.WriteError(c, fmt.Errorf("%w: приём завершён", pkgerrors.ErrFailedPrecondition))
	})
	srv := httptest.NewServer(adaptorHandler(app))
	defer srv.Close()

	cl := httpclient.New(srv.URL, time.Second)
	ctx := context.Background()

	var out struct {
		Name string `json:"name"`
	}
	if err := cl.Post(ctx, "/echo", map[string]string{"name": "x"}, &out); err != nil || out.Name != "x" {
		t.Fatalf("echo: %+v %v", out, err)
	}
	err := cl.Post(ctx, "/conflict", struct{}{}, &out)
	if !errors.Is(err, pkgerrors.ErrConflict) || !strings.Contains(err.Error(), "слот занят") {
		t.Fatalf("conflict: %v", err)
	}
	err = cl.Post(ctx, "/state", struct{}{}, &out)
	if !errors.Is(err, pkgerrors.ErrFailedPrecondition) {
		t.Fatalf("precondition: %v", err)
	}
}

func adaptorHandler(app *fiber.App) http.Handler {
	return adaptor.FiberApp(app)
}
