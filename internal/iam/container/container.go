package container

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/PaPaSmUrFiK/OuchFix/internal/iam/config"
	"github.com/PaPaSmUrFiK/OuchFix/migrations"
	"github.com/PaPaSmUrFiK/OuchFix/pkg/storage/postgres"
)

// Container — composition root IAM-сервиса. В ЛР3 он только владеет схемой iam
// (миграции и сид); регистрация, вход и права — ЛР5.
type Container struct {
	Config *config.Config
	db     *sqlx.DB
}

func New(ctx context.Context, cfg *config.Config) (*Container, error) {
	db, err := postgres.New(ctx, cfg.DB)
	if err != nil {
		return nil, fmt.Errorf("connect iam db: %w", err)
	}
	if err := postgres.Migrate(ctx, db, "iam", migrations.IAM); err != nil {
		if closeErr := db.Close(); closeErr != nil {
			return nil, fmt.Errorf("migrate iam: %w (close: %s)", err, closeErr)
		}
		return nil, fmt.Errorf("migrate iam: %w", err)
	}

	return &Container{Config: cfg, db: db}, nil
}

func (c *Container) Close() error {
	return c.db.Close()
}
