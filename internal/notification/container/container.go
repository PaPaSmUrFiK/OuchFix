package container

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v3"
	"github.com/jmoiron/sqlx"

	"github.com/PaPaSmUrFiK/OuchFix/internal/notification/config"
	notificationrepo "github.com/PaPaSmUrFiK/OuchFix/internal/notification/repository/postgres"
	notificationsvc "github.com/PaPaSmUrFiK/OuchFix/internal/notification/service"
	"github.com/PaPaSmUrFiK/OuchFix/internal/notification/transport"
	iamclient "github.com/PaPaSmUrFiK/OuchFix/pkg/client/iam"
	"github.com/PaPaSmUrFiK/OuchFix/pkg/storage/postgres"
)

type Repos struct {
	notification *notificationrepo.NotificationRepository
}

func (c *Container) NewRepos() *Repos {
	return &Repos{
		notification: notificationrepo.NewNotificationRepository(c.db),
	}
}

type Services struct {
	notification *notificationsvc.NotificationService
}

// Межсервисные клиенты в ЛР3 — заглушки по proto из ЛР2, в ЛР6 — gRPC.
func (c *Container) NewServices() *Services {
	return &Services{
		notification: notificationsvc.NewNotificationService(c.repos.notification, iamclient.NewStub()),
	}
}

type Handlers struct {
	Router func(fiber.Router)
}

func (c *Container) NewHandlers() *Handlers {
	return &Handlers{
		Router: transport.New(c.services.notification, c.Config.Location),
	}
}

// Container — composition root notification-сервиса: БД, репозиторий, сервис и маршруты.
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

	if err := notificationrepo.Migrate(ctx, db); err != nil {
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
