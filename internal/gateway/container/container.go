package container

import (
	"context"
	"errors"
	"fmt"
	"io/fs"

	"github.com/jmoiron/sqlx"

	"github.com/PaPaSmUrFiK/OuchFix/internal/gateway/config"
	"github.com/PaPaSmUrFiK/OuchFix/migrations"
	pkgconfig "github.com/PaPaSmUrFiK/OuchFix/pkg/config"
	"github.com/PaPaSmUrFiK/OuchFix/pkg/storage/postgres"
)

type Storages struct {
	iam          *sqlx.DB
	catalog      *sqlx.DB
	appointment  *sqlx.DB
	payment      *sqlx.DB
	notification *sqlx.DB
}

func (s *Storages) all() []*sqlx.DB {
	return []*sqlx.DB{s.iam, s.catalog, s.appointment, s.payment, s.notification}
}

func (s *Storages) Close() error {
	var errs []error
	for _, db := range s.all() {
		if db != nil {
			errs = append(errs, db.Close())
		}
	}
	return errors.Join(errs...)
}

// NewStorages открывает БД каждого сервиса и накатывает его миграции. При ошибке
// уже открытые подключения закрываются.
func (c *Container) NewStorages(ctx context.Context) (_ *Storages, err error) {
	s := &Storages{}
	defer func() {
		if err != nil {
			if closeErr := s.Close(); closeErr != nil {
				err = fmt.Errorf("%w (close: %s)", err, closeErr)
			}
		}
	}()

	for _, db := range []struct {
		target *(*sqlx.DB)
		schema string
		cfg    pkgconfig.Postgres
		files  fs.FS
	}{
		{&s.iam, "iam", c.Config.DB.IAM, migrations.IAM},
		{&s.catalog, "catalog", c.Config.DB.Catalog, migrations.Catalog},
		{&s.appointment, "appointment", c.Config.DB.Appointment, migrations.Appointment},
		{&s.payment, "payment", c.Config.DB.Payment, migrations.Payment},
		{&s.notification, "notification", c.Config.DB.Notification, migrations.Notification},
	} {
		conn, err := postgres.New(ctx, db.cfg)
		if err != nil {
			return nil, fmt.Errorf("connect %s db: %w", db.schema, err)
		}
		*db.target = conn

		if err := postgres.Migrate(ctx, conn, db.schema, db.files); err != nil {
			return nil, fmt.Errorf("migrate %s: %w", db.schema, err)
		}
	}

	return s, nil
}

// Container — единственный composition root gateway: здесь собираются БД,
// репозитории, сервисы, клиенты и транспорт.
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

	return c, nil
}

func (c *Container) Close() error {
	if c.storages == nil {
		return nil
	}

	return c.storages.Close()
}
