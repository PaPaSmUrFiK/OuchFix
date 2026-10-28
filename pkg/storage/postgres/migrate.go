package postgres

import (
	"context"
	"fmt"
	"io/fs"
	"regexp"
	"sort"

	"github.com/jmoiron/sqlx"
)

var schemaName = regexp.MustCompile(`^[a-z_][a-z0-9_]*$`)

// Migrate накатывает *.sql из fsys по порядку имён в схему schema. Применённые
// версии хранятся в <schema>.schema_migrations, каждый файл идёт в своей транзакции.
func Migrate(ctx context.Context, db *sqlx.DB, schema string, fsys fs.FS) error {
	if !schemaName.MatchString(schema) {
		return fmt.Errorf("invalid schema name %q", schema)
	}

	bootstrap := fmt.Sprintf(`
		CREATE SCHEMA IF NOT EXISTS %[1]s;
		CREATE TABLE IF NOT EXISTS %[1]s.schema_migrations (
			version    VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
		);`, schema)
	if _, err := db.ExecContext(ctx, bootstrap); err != nil {
		return fmt.Errorf("bootstrap %s migrations: %w", schema, err)
	}

	files, err := fs.Glob(fsys, "*/*.sql")
	if err != nil {
		return fmt.Errorf("list %s migrations: %w", schema, err)
	}
	sort.Strings(files)

	for _, name := range files {
		if err := applyOne(ctx, db, schema, fsys, name); err != nil {
			return err
		}
	}

	return nil
}

func applyOne(ctx context.Context, db *sqlx.DB, schema string, fsys fs.FS, name string) (err error) {
	var applied bool
	check := fmt.Sprintf(`SELECT EXISTS (SELECT 1 FROM %s.schema_migrations WHERE version = $1)`, schema)
	if err := db.GetContext(ctx, &applied, check, name); err != nil {
		return fmt.Errorf("check migration %s: %w", name, err)
	}
	if applied {
		return nil
	}

	body, err := fs.ReadFile(fsys, name)
	if err != nil {
		return fmt.Errorf("read migration %s: %w", name, err)
	}

	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin migration %s: %w", name, err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if _, err = tx.ExecContext(ctx, string(body)); err != nil {
		return fmt.Errorf("apply migration %s: %w", name, err)
	}
	mark := fmt.Sprintf(`INSERT INTO %s.schema_migrations (version) VALUES ($1)`, schema)
	if _, err = tx.ExecContext(ctx, mark, name); err != nil {
		return fmt.Errorf("mark migration %s: %w", name, err)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit migration %s: %w", name, err)
	}

	return nil
}
