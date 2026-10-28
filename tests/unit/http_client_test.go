package unit

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"

	pkgerrors "github.com/PaPaSmUrFiK/OuchFix/pkg/errors"
	httpclient "github.com/PaPaSmUrFiK/OuchFix/pkg/transport/http/client"
	httpserver "github.com/PaPaSmUrFiK/OuchFix/pkg/transport/http/server"
)

// Ошибка сервиса проходит через REST с тем же текстом и тем же доменным видом,
// а внутренние ошибки наружу не утекают.
func TestHTTPClientKeepsDomainErrors(t *testing.T) {
	app := fiber.New()
	app.Get("/nf", func(c *fiber.Ctx) error {
		return httpserver.WriteError(c, pkgerrors.ErrNotFound)
	})
	app.Get("/bad/:id", func(c *fiber.Ctx) error {
		_, err := httpserver.ParamID(c, "id")
		return httpserver.WriteError(c, err)
	})
	app.Get("/boom", func(c *fiber.Ctx) error {
		return httpserver.WriteError(c, errors.New("pq: secret sql"))
	})
	app.Get("/ok", func(c *fiber.Ctx) error {
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

func adaptorHandler(app *fiber.App) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp, err := app.Test(r, -1)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()
		w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
		w.WriteHeader(resp.StatusCode)
		_, _ = io.Copy(w, resp.Body)
	})
}
