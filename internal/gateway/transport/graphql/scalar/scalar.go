// Package scalar — скаляры Date и DateTime внешней GraphQL-схемы. Значения
// хранятся строкой как есть: gateway только проверяет формат и проксирует
// их в сервисы без пересчёта поясов.
package scalar

import (
	"fmt"
	"io"
	"strconv"
	"time"

	pkgerrors "github.com/PaPaSmUrFiK/OuchFix/pkg/errors"
)

const dateLayout = "2006-01-02"

// Date — календарный день YYYY-MM-DD.
type Date string

// DateTime — момент времени в RFC 3339 с поясом, например 2026-11-20T09:00:00+03:00.
type DateTime string

func (d *Date) UnmarshalGQL(v any) error {
	s, err := parse(v, "Date", dateLayout, "YYYY-MM-DD")
	*d = Date(s)
	return err
}

func (d Date) MarshalGQL(w io.Writer) { writeString(w, string(d)) }

func (t *DateTime) UnmarshalGQL(v any) error {
	s, err := parse(v, "DateTime", time.RFC3339, "RFC 3339, например 2026-11-20T09:00:00+03:00")
	*t = DateTime(s)
	return err
}

func (t DateTime) MarshalGQL(w io.Writer) { writeString(w, string(t)) }

func parse(v any, name, layout, hint string) (string, error) {
	s, ok := v.(string)
	if !ok {
		return "", fmt.Errorf("%w: %s должен быть строкой, получено %T", pkgerrors.ErrInvalidArgument, name, v)
	}
	if _, err := time.Parse(layout, s); err != nil {
		return "", fmt.Errorf("%w: %s должен быть в формате %s, получено %q", pkgerrors.ErrInvalidArgument, name, hint, s)
	}
	return s, nil
}

func writeString(w io.Writer, s string) {
	_, _ = io.WriteString(w, strconv.Quote(s))
}
