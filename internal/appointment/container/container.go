package container

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v3"
	"github.com/jmoiron/sqlx"

	"github.com/PaPaSmUrFiK/OuchFix/internal/appointment/config"
	appointmentrepo "github.com/PaPaSmUrFiK/OuchFix/internal/appointment/repository/postgres"
	appointmentsvc "github.com/PaPaSmUrFiK/OuchFix/internal/appointment/service"
	"github.com/PaPaSmUrFiK/OuchFix/internal/appointment/transport"
	catalogclient "github.com/PaPaSmUrFiK/OuchFix/pkg/client/catalog"
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
	appointment *appointmentrepo.AppointmentRepository
}

func (c *Container) NewRepos() *Repos {
	return &Repos{
		appointment: appointmentrepo.NewAppointmentRepository(c.storages.db),
	}
}

type Services struct {
	appointment *appointmentsvc.AppointmentService
}

// Межсервисные клиенты в ЛР3 — заглушки по proto из ЛР2, в ЛР6 — gRPC.
func (c *Container) NewServices() *Services {
	return &Services{
		appointment: appointmentsvc.NewAppointmentService(c.repos.appointment, catalogclient.NewStub(), c.Config.Location),
	}
}

type Handlers struct {
	Router func(fiber.Router)
}

func (c *Container) NewHandlers() *Handlers {
	return &Handlers{
		Router: transport.New(c.services.appointment, c.Config.Location),
	}
}

// Container — composition root appointment-сервиса: БД, репозиторий, сервис и маршруты.
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

	if err := appointmentrepo.Migrate(ctx, storages.db); err != nil {
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
