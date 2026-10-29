package httpserver

import (
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
)

// Ответы внутреннего REST API сразу в форме GraphQL-моделей gateway:
// id — строки, enum — в верхнем регистре, время — RFC 3339 в поясе медцентра.

func FormatID(id int64) string {
	return strconv.FormatInt(id, 10)
}

func FormatOptionalID(id *int64) *string {
	if id == nil {
		return nil
	}
	s := FormatID(*id)
	return &s
}

func FormatTime(t time.Time, loc *time.Location) string {
	return t.In(loc).Format(time.RFC3339)
}

func FormatOptionalTime(t *time.Time, loc *time.Location) *string {
	if t == nil {
		return nil
	}
	s := FormatTime(*t, loc)
	return &s
}

// FormatEnum: в БД значения в нижнем регистре (walk_in), в GraphQL — в верхнем (WALK_IN).
func FormatEnum[E ~string](value E) string {
	return strings.ToUpper(string(value))
}

// QueryEnum читает необязательный enum из query-строки в формате GraphQL
// (COMPLETED) и переводит в значение сервиса (completed).
func QueryEnum[E ~string](c fiber.Ctx, name string) *E {
	raw := c.Query(name)
	if raw == "" {
		return nil
	}
	e := E(strings.ToLower(raw))
	return &e
}

// MapList переводит срез доменных структур в ответ; пустой срез — [], не null.
func MapList[S any, D any](items []S, convert func(*S) D) []D {
	out := make([]D, 0, len(items))
	for i := range items {
		out = append(out, convert(&items[i]))
	}
	return out
}
