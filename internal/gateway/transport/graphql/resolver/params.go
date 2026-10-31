package resolver

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/PaPaSmUrFiK/OuchFix/internal/gateway/transport/graphql/model"
	pkgerrors "github.com/PaPaSmUrFiK/OuchFix/pkg/errors"
)

// parseID проверяет GraphQL ID до запроса к сервису: нечисловой id — ошибка
// аргумента, сервис не вызывается.
func parseID(name, raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if _, err := strconv.ParseInt(raw, 10, 64); err != nil {
		return "", fmt.Errorf("%w: %s должен быть числом, получено %q", pkgerrors.ErrInvalidArgument, name, raw)
	}
	return raw, nil
}

// optionalQuery добавляет необязательный параметр в query-строку запроса к сервису.
func optionalQuery[T ~string](query url.Values, name string, value *T) {
	if value != nil {
		query.Set(name, string(*value))
	}
}

// notImplemented — операции следующих лабораторных: схема уже есть, логики ещё нет.
func notImplemented(field, lab string) error {
	return fmt.Errorf("%w: %s появится в %s", pkgerrors.ErrNotImplemented, field, lab)
}

// postAppointment — бизнес-операция над приёмом: сервис отвечает приёмом целиком.
func postAppointment(ctx context.Context, svc Service, path string, body any) (*model.Appointment, error) {
	var out model.Appointment
	if err := svc.Post(ctx, path, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
