package postgres

import (
	"context"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"

	"github.com/PaPaSmUrFiK/OuchFix/pkg/config"
)

const pingTimeout = 5 * time.Second

func New(ctx context.Context, cfg config.Postgres) (*sqlx.DB, error) {
	db, err := sqlx.Open("pgx", cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.MaxLifetime)
	db.SetConnMaxIdleTime(cfg.MaxIdleTime)

	pingCtx, cancel := context.WithTimeout(ctx, pingTimeout)
	defer cancel()

	if err := db.PingContext(pingCtx); err != nil {
		if closeErr := db.Close(); closeErr != nil {
			return nil, fmt.Errorf("ping db: %w (close: %s)", err, closeErr)
		}
		return nil, fmt.Errorf("ping db: %w", err)
	}

	return db, nil
}
