package postgres

import (
	"context"
	"fmt"
	"io/fs"
	"regexp"

	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"
)

var schemaName = regexp.MustCompile(`^[a-z_][a-z0-9_]*$`)

// Migrate накатывает goose-миграции из каталога schema в fsys. Версии хранятся
// в <schema>.goose_db_version — в схеме сервиса, куда у его пользователя есть
// права. Саму схему и её владельца заранее создаёт администратор
// (docker/postgres/init.sql): у пользователя сервиса нет прав создавать схемы.
func Migrate(ctx context.Context, db *sqlx.DB, schema string, fsys fs.FS) error {
	if !schemaName.MatchString(schema) {
		return fmt.Errorf("invalid schema name %q", schema)
	}

	dir, err := fs.Sub(fsys, schema)
	if err != nil {
		return fmt.Errorf("open %s migrations: %w", schema, err)
	}

	provider, err := goose.NewProvider(goose.DialectPostgres, db.DB, dir,
		goose.WithTableName(schema+".goose_db_version"),
	)
	if err != nil {
		return fmt.Errorf("init %s migrations: %w", schema, err)
	}

	if _, err := provider.Up(ctx); err != nil {
		return fmt.Errorf("apply %s migrations: %w", schema, err)
	}

	return nil
}
