package container

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v3"
	"github.com/jmoiron/sqlx"

	"github.com/PaPaSmUrFiK/OuchFix/internal/payment/config"
	paymentrepo "github.com/PaPaSmUrFiK/OuchFix/internal/payment/repository/postgres"
	paymentsvc "github.com/PaPaSmUrFiK/OuchFix/internal/payment/service"
	"github.com/PaPaSmUrFiK/OuchFix/internal/payment/transport"
	appointmentclient "github.com/PaPaSmUrFiK/OuchFix/pkg/client/appointment"
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
	payment *paymentrepo.PaymentRepository
}

func (c *Container) NewRepos() *Repos {
	return &Repos{
		payment: paymentrepo.NewPaymentRepository(c.storages.db),
	}
}

type Services struct {
	payment *paymentsvc.PaymentService
}

// Межсервисные клиенты в ЛР3 — заглушки по proto из ЛР2, в ЛР6 — gRPC.
func (c *Container) NewServices() *Services {
	return &Services{
		payment: paymentsvc.NewPaymentService(c.repos.payment, catalogclient.NewStub(), appointmentclient.NewStub()),
	}
}

type Handlers struct {
	Router func(fiber.Router)
}

func (c *Container) NewHandlers() *Handlers {
	return &Handlers{
		Router: transport.New(c.services.payment, c.Config.Location),
	}
}

// Container — composition root payment-сервиса: БД, репозиторий, сервис и маршруты.
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

	if err := paymentrepo.Migrate(ctx, storages.db); err != nil {
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
