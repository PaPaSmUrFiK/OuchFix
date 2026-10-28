package postgres

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"

	pkgerrors "github.com/PaPaSmUrFiK/OuchFix/pkg/errors"
)

const invalidTextRepresentation = "22P02"

// WrapDB переводит технические ошибки БД в доменные, сохраняя исходную причину.
func WrapDB(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("%w: %w", pkgerrors.ErrNotFound, err)
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == invalidTextRepresentation {
		return fmt.Errorf("%w: %w", pkgerrors.ErrInvalidArgument, err)
	}

	return err
}
