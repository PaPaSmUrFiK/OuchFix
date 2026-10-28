package container

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/PaPaSmUrFiK/OuchFix/internal/notification/config"
	notificationrepo "github.com/PaPaSmUrFiK/OuchFix/internal/notification/repository/postgres"
	notificationsvc "github.com/PaPaSmUrFiK/OuchFix/internal/notification/service"
	"github.com/PaPaSmUrFiK/OuchFix/migrations"
	iamclient "github.com/PaPaSmUrFiK/OuchFix/pkg/client/iam"
	"github.com/PaPaSmUrFiK/OuchFix/pkg/storage/postgres"
)

// Container — composition root notification-сервиса: своя БД, репозиторий, сервис.
// Межсервисные клиенты в ЛР3 — заглушки по proto из ЛР2, в ЛР6 — gRPC.
type Container struct {
	Config  *config.Config
	db      *sqlx.DB
	service *notificationsvc.NotificationService
}

func New(ctx context.Context, cfg *config.Config) (*Container, error) {
	db, err := postgres.New(ctx, cfg.DB)
	if err != nil {
		return nil, fmt.Errorf("connect notification db: %w", err)
	}
	if err := postgres.Migrate(ctx, db, "notification", migrations.Notification); err != nil {
		if closeErr := db.Close(); closeErr != nil {
			return nil, fmt.Errorf("migrate notification: %w (close: %s)", err, closeErr)
		}
		return nil, fmt.Errorf("migrate notification: %w", err)
	}

	return &Container{
		Config:  cfg,
		db:      db,
		service: notificationsvc.NewNotificationService(notificationrepo.NewNotificationRepository(db), iamclient.NewStub()),
	}, nil
}

func (c *Container) Service() *notificationsvc.NotificationService {
	return c.service
}

func (c *Container) Close() error {
	return c.db.Close()
}
