package container

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/PaPaSmUrFiK/OuchFix/internal/iam/config"
	iamrepo "github.com/PaPaSmUrFiK/OuchFix/internal/iam/repository/postgres"
	"github.com/PaPaSmUrFiK/OuchFix/pkg/storage/postgres"
)

type Storages struct {
	db *sqlx.DB
}

func (c *Container) NewStorages(ctx context.Context) (*Storages, error) {
	db, err := postgres.New(ctx, c.Config.DB)
	if err != nil {
		return nil, fmt.Errorf("new postgres: %w", err)
	}

	return &Storages{db: db}, nil
}

// Container — composition root IAM-сервиса. В ЛР3 он только владеет схемой iam
// (миграции и сид); регистрация, вход и права — ЛР5.
type Container struct {
	Config   *config.Config
	storages *Storages
}

func New(ctx context.Context, cfg *config.Config) (*Container, error) {
	c := &Container{Config: cfg}

	storages, err := c.NewStorages(ctx)
	if err != nil {
		return nil, err
	}
	c.storages = storages

	if err := iamrepo.Migrate(ctx, storages.db); err != nil {
		closeErr := storages.db.Close()
		if closeErr != nil {
			return nil, fmt.Errorf("migrate: %w (close: %s)", err, closeErr)
		}
		return nil, fmt.Errorf("migrate: %w", err)
	}

	return c, nil
}

func (c *Container) Close() error {
	if c.storages == nil {
		return nil
	}

	return c.storages.db.Close()
}
