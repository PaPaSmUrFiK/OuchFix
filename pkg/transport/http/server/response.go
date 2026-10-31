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
	case errors.Is(err, pkgerrors.ErrConflict):
		status = fiber.StatusConflict
	case errors.Is(err, pkgerrors.ErrFailedPrecondition):
		status = fiber.StatusUnprocessableEntity
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

// BindJSON декодирует тело запроса; битый JSON — ошибка аргумента.
func BindJSON(c fiber.Ctx, out any) error {
	if err := c.Bind().JSON(out); err != nil {
		return fmt.Errorf("%w: некорректное тело запроса", pkgerrors.ErrInvalidArgument)
	}
	return nil
}

// QueryInt читает необязательное целое из query-строки.
func QueryInt(c fiber.Ctx, name string) (*int, error) {
	raw := c.Query(name)
	if raw == "" {
		return nil, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return nil, fmt.Errorf("%w: %s должен быть целым числом, получено %q", pkgerrors.ErrInvalidArgument, name, raw)
	}
	return &n, nil
}

// ParseID разбирает числовой id из тела запроса (в JSON id — строки, как в GraphQL).
func ParseID(name, raw string) (int64, error) {
	return parseID(name, raw)
}

// ParseIDs разбирает список id; пустой список — ошибка аргумента.
func ParseIDs(name string, raw []string) ([]int64, error) {
	if len(raw) == 0 {
		return nil, fmt.Errorf("%w: %s не должен быть пустым", pkgerrors.ErrInvalidArgument, name)
	}
	ids := make([]int64, 0, len(raw))
	for _, r := range raw {
		id, err := parseID(name, r)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// ParseOptionalID разбирает необязательный id; nil — значение не задано.
func ParseOptionalID(name string, raw *string) (*int64, error) {
	if raw == nil {
		return nil, nil
	}
	id, err := parseID(name, *raw)
	if err != nil {
		return nil, err
	}
	return &id, nil
}

func parseID(name, raw string) (int64, error) {
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%w: %s должен быть числом, получено %q", pkgerrors.ErrInvalidArgument, name, raw)
	}
	return id, nil
}
