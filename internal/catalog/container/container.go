package container

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v3"
	"github.com/jmoiron/sqlx"

	"github.com/PaPaSmUrFiK/OuchFix/internal/catalog/config"
	catalogrepo "github.com/PaPaSmUrFiK/OuchFix/internal/catalog/repository/postgres"
	catalogsvc "github.com/PaPaSmUrFiK/OuchFix/internal/catalog/service"
	"github.com/PaPaSmUrFiK/OuchFix/internal/catalog/transport"
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

type Repos struct {
	catalog *catalogrepo.CatalogRepository
}

func (c *Container) NewRepos() *Repos {
	return &Repos{
		catalog: catalogrepo.NewCatalogRepository(c.storages.db),
	}
}

type Services struct {
	catalog *catalogsvc.CatalogService
}

func (c *Container) NewServices() *Services {
	return &Services{
		catalog: catalogsvc.NewCatalogService(c.repos.catalog),
	}
}

type Handlers struct {
	Router func(fiber.Router)
}

func (c *Container) NewHandlers() *Handlers {
	return &Handlers{
		Router: transport.New(c.services.catalog),
	}
}

// Container — composition root catalog-сервиса: БД, репозиторий, сервис и маршруты.
type Container struct {
	Config   *config.Config
	storages *Storages
	repos    *Repos
	services *Services
	Handlers *Handlers
}

func New(ctx context.Context, cfg *config.Config) (*Container, error) {
	c := &Container{Config: cfg}

	storages, err := c.NewStorages(ctx)
	if err != nil {
		return nil, err
	}
	c.storages = storages

	if err := catalogrepo.Migrate(ctx, storages.db); err != nil {
		closeErr := storages.db.Close()
		if closeErr != nil {
			return nil, fmt.Errorf("migrate: %w (close: %s)", err, closeErr)
		}
		return nil, fmt.Errorf("migrate: %w", err)
	}

	c.repos = c.NewRepos()
	c.services = c.NewServices()
	c.Handlers = c.NewHandlers()

	return c, nil
}

func (c *Container) Close() error {
	if c.storages == nil {
		return nil
	}

	return c.storages.db.Close()
}
