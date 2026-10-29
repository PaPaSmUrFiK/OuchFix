package container

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/PaPaSmUrFiK/OuchFix/internal/payment/config"
	paymentrepo "github.com/PaPaSmUrFiK/OuchFix/internal/payment/repository/postgres"
	paymentsvc "github.com/PaPaSmUrFiK/OuchFix/internal/payment/service"
	appointmentclient "github.com/PaPaSmUrFiK/OuchFix/pkg/client/appointment"
	catalogclient "github.com/PaPaSmUrFiK/OuchFix/pkg/client/catalog"
	"github.com/PaPaSmUrFiK/OuchFix/pkg/storage/postgres"
)

// Container — composition root payment-сервиса: своя БД, репозиторий, сервис.
// Межсервисные клиенты в ЛР3 — заглушки по proto из ЛР2, в ЛР6 — gRPC.
type Container struct {
	Config  *config.Config
	db      *sqlx.DB
	service *paymentsvc.PaymentService
}

func New(ctx context.Context, cfg *config.Config) (*Container, error) {
	db, err := postgres.New(ctx, cfg.DB)
	if err != nil {
		return nil, fmt.Errorf("connect payment db: %w", err)
	}
	if err := paymentrepo.Migrate(ctx, db); err != nil {
		if closeErr := db.Close(); closeErr != nil {
			return nil, fmt.Errorf("migrate payment: %w (close: %s)", err, closeErr)
		}
		return nil, fmt.Errorf("migrate payment: %w", err)
	}

	return &Container{
		Config:  cfg,
		db:      db,
		service: paymentsvc.NewPaymentService(paymentrepo.NewPaymentRepository(db), catalogclient.NewStub(), appointmentclient.NewStub()),
	}, nil
}

func (c *Container) Service() *paymentsvc.PaymentService {
	return c.service
}

func (c *Container) Close() error {
	return c.db.Close()
}
