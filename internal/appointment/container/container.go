package container

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/PaPaSmUrFiK/OuchFix/internal/appointment/config"
	appointmentrepo "github.com/PaPaSmUrFiK/OuchFix/internal/appointment/repository/postgres"
	appointmentsvc "github.com/PaPaSmUrFiK/OuchFix/internal/appointment/service"
	catalogclient "github.com/PaPaSmUrFiK/OuchFix/pkg/client/catalog"
	"github.com/PaPaSmUrFiK/OuchFix/pkg/storage/postgres"
)

// Container — composition root appointment-сервиса: своя БД, репозиторий, сервис.
// Межсервисные клиенты в ЛР3 — заглушки по proto из ЛР2, в ЛР6 — gRPC.
type Container struct {
	Config  *config.Config
	db      *sqlx.DB
	service *appointmentsvc.AppointmentService
}

func New(ctx context.Context, cfg *config.Config) (*Container, error) {
	db, err := postgres.New(ctx, cfg.DB)
	if err != nil {
		return nil, fmt.Errorf("connect appointment db: %w", err)
	}
	if err := appointmentrepo.Migrate(ctx, db); err != nil {
		if closeErr := db.Close(); closeErr != nil {
			return nil, fmt.Errorf("migrate appointment: %w (close: %s)", err, closeErr)
		}
		return nil, fmt.Errorf("migrate appointment: %w", err)
	}

	return &Container{
		Config:  cfg,
		db:      db,
		service: appointmentsvc.NewAppointmentService(appointmentrepo.NewAppointmentRepository(db), catalogclient.NewStub(), cfg.Location),
	}, nil
}

func (c *Container) Service() *appointmentsvc.AppointmentService {
	return c.service
}

func (c *Container) Close() error {
	return c.db.Close()
}
