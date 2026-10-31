package graphql

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	gqlgraphql "github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/v2/gqlerror"

	pkgerrors "github.com/PaPaSmUrFiK/OuchFix/pkg/errors"
)

// Коды ошибок в extensions.code — контракт для клиентов API.
const (
	codeNotFound       = "NOT_FOUND"
	codeBadUserInput   = "BAD_USER_INPUT"
	codeNotImplemented = "NOT_IMPLEMENTED"
	codeConflict       = "CONFLICT"
	codePrecondition   = "FAILED_PRECONDITION"
	codeInternal       = "INTERNAL"
)

const internalMessage = "internal server error"

// presentError — единственное место, где доменные ошибки превращаются в
// GraphQL-ошибки. Сервисы и резолверы возвращают обычные error с %w.
func presentError(ctx context.Context, err error) *gqlerror.Error {
	gqlErr := gqlgraphql.DefaultErrorPresenter(ctx, err)

	var original *gqlerror.Error
	if errors.As(err, &original) && original.Err == nil {
		// Ошибки разбора и валидации запроса от самого gqlgen: у них уже свой код.
		return gqlErr
	}

	switch {
	case errors.Is(err, pkgerrors.ErrNotFound):
		setCode(gqlErr, codeNotFound)
	case errors.Is(err, pkgerrors.ErrInvalidArgument):
		setCode(gqlErr, codeBadUserInput)
	case errors.Is(err, pkgerrors.ErrNotImplemented):
		setCode(gqlErr, codeNotImplemented)
	case errors.Is(err, pkgerrors.ErrConflict):
		setCode(gqlErr, codeConflict)
	case errors.Is(err, pkgerrors.ErrFailedPrecondition):
		setCode(gqlErr, codePrecondition)
	default:
		// Текст внутренних ошибок (SQL, сеть) наружу не отдаём, только в лог.
		slog.ErrorContext(ctx, "graphql request failed", "path", gqlErr.Path.String(), "error", err)
		gqlErr.Message = internalMessage
		setCode(gqlErr, codeInternal)
	}

	return gqlErr
}

func recoverPanic(ctx context.Context, p any) error {
	slog.ErrorContext(ctx, "graphql resolver panic", "panic", fmt.Sprint(p))
	return errors.New("panic in resolver")
}

func setCode(e *gqlerror.Error, code string) {
	if e.Extensions == nil {
		e.Extensions = map[string]any{}
	}
	e.Extensions["code"] = code
}
