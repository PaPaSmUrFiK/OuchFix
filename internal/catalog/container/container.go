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

type Repos struct {
	catalog *catalogrepo.CatalogRepository
}

func (c *Container) NewRepos() *Repos {
	return &Repos{
		catalog: catalogrepo.NewCatalogRepository(c.db),
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
	db       *sqlx.DB
	repos    *Repos
	services *Services
	Handlers *Handlers
}

func New(ctx context.Context, cfg *config.Config) (*Container, error) {
	c := &Container{Config: cfg}

	db, err := postgres.New(ctx, cfg.DB)
	if err != nil {
		return nil, fmt.Errorf("new postgres: %w", err)
	}
	c.db = db

	if err := catalogrepo.Migrate(ctx, db); err != nil {
		closeErr := db.Close()
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
	if c.db == nil {
		return nil
	}

	return c.db.Close()
}
