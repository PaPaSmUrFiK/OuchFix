package postgres

import (
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"

	pkgerrors "github.com/PaPaSmUrFiK/OuchFix/pkg/errors"
)

const invalidTextRepresentation = "22P02"

// WrapDB переводит технические ошибки БД в доменные. Для «не найдено» и
// «неверный формат» причина не сохраняется: её текст (sql: no rows, текст
// ошибки postgres) уходил бы клиенту в сообщении GraphQL-ошибки. Остальные
// ошибки возвращаются как есть и на границе маскируются как INTERNAL.
func WrapDB(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return pkgerrors.ErrNotFound
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == invalidTextRepresentation {
		return pkgerrors.ErrInvalidArgument
	}

	return err
}
