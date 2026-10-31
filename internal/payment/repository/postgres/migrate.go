package postgres

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"

	"github.com/PaPaSmUrFiK/OuchFix/migrations/payment"
)

func Migrate(ctx context.Context, db *sqlx.DB) error {
	goose.SetBaseFS(payment.Files)
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set goose dialect: %w", err)
	}
	if err := goose.UpContext(ctx, db.DB, "."); err != nil {
		return fmt.Errorf("goose up: %w", err)
	}
	return nil
}
