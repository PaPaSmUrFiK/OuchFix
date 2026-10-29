package httpserver

import (
	"errors"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/gofiber/fiber/v3"

	pkgerrors "github.com/PaPaSmUrFiK/OuchFix/pkg/errors"
)

const internalMessage = "internal server error"

// ErrorBody — тело ошибки внутреннего REST API сервисов.
type ErrorBody struct {
	Error string `json:"error"`
}

// WriteError переводит доменную ошибку в HTTP-статус. Текст внутренних ошибок
// (SQL, сеть) наружу не отдаётся, только в лог.
func WriteError(c fiber.Ctx, err error) error {
	status := fiber.StatusInternalServerError
	switch {
	case errors.Is(err, pkgerrors.ErrNotFound):
		status = fiber.StatusNotFound
	case errors.Is(err, pkgerrors.ErrInvalidArgument):
		status = fiber.StatusBadRequest
	case errors.Is(err, pkgerrors.ErrNotImplemented):
		status = fiber.StatusNotImplemented
	default:
		slog.ErrorContext(c.Context(), "request failed", "path", c.Path(), "error", err)
		return c.Status(status).JSON(ErrorBody{Error: internalMessage})
	}
	return c.Status(status).JSON(ErrorBody{Error: err.Error()})
}

// Reply отдаёт результат сервиса как JSON или ошибку через WriteError.
func Reply(c fiber.Ctx, data any, err error) error {
	if err != nil {
		return WriteError(c, err)
	}
	return c.JSON(data)
}

// ParamID читает числовой id из пути. Нечисловой id — ошибка аргумента.
func ParamID(c fiber.Ctx, name string) (int64, error) {
	return parseID(name, c.Params(name))
}

// QueryID читает необязательный числовой id из query-строки.
func QueryID(c fiber.Ctx, name string) (*int64, error) {
	raw := c.Query(name)
	if raw == "" {
		return nil, nil
	}
	id, err := parseID(name, raw)
	if err != nil {
		return nil, err
	}
	return &id, nil
}

// QueryString читает необязательную строку из query-строки.
func QueryString(c fiber.Ctx, name string) *string {
	raw := c.Query(name)
	if raw == "" {
		return nil
	}
	return &raw
}

func parseID(name, raw string) (int64, error) {
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%w: %s должен быть числом, получено %q", pkgerrors.ErrInvalidArgument, name, raw)
	}
	return id, nil
}
