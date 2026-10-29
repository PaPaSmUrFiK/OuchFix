package container

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/PaPaSmUrFiK/OuchFix/internal/catalog/config"
	catalogrepo "github.com/PaPaSmUrFiK/OuchFix/internal/catalog/repository/postgres"
	catalogsvc "github.com/PaPaSmUrFiK/OuchFix/internal/catalog/service"
	"github.com/PaPaSmUrFiK/OuchFix/pkg/storage/postgres"
)

// Container — composition root catalog-сервиса: своя БД, репозиторий, сервис.
// Межсервисные клиенты в ЛР3 — заглушки по proto из ЛР2, в ЛР6 — gRPC.
type Container struct {
	Config  *config.Config
	db      *sqlx.DB
	service *catalogsvc.CatalogService
}

func New(ctx context.Context, cfg *config.Config) (*Container, error) {
	db, err := postgres.New(ctx, cfg.DB)
	if err != nil {
		return nil, fmt.Errorf("connect catalog db: %w", err)
	}
	if err := catalogrepo.Migrate(ctx, db); err != nil {
		if closeErr := db.Close(); closeErr != nil {
			return nil, fmt.Errorf("migrate catalog: %w (close: %s)", err, closeErr)
		}
		return nil, fmt.Errorf("migrate catalog: %w", err)
	}

	return &Container{
		Config:  cfg,
		db:      db,
		service: catalogsvc.NewCatalogService(catalogrepo.NewCatalogRepository(db)),
	}, nil
}

func (c *Container) Service() *catalogsvc.CatalogService {
	return c.service
}

func (c *Container) Close() error {
	return c.db.Close()
}
