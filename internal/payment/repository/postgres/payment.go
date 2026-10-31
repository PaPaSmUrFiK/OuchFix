package postgres

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/PaPaSmUrFiK/OuchFix/internal/payment/domain"
	"github.com/PaPaSmUrFiK/OuchFix/internal/payment/repository/postgres/models"
	"github.com/PaPaSmUrFiK/OuchFix/pkg/storage/postgres"
)

type PaymentRepository struct {
	db *sqlx.DB
}

func NewPaymentRepository(db *sqlx.DB) *PaymentRepository {
	return &PaymentRepository{db: db}
}

// List возвращает платежи по фильтру; пустой фильтр — все платежи.
func (r *PaymentRepository) List(ctx context.Context, f domain.Filter) ([]domain.Payment, error) {
	const query = `
		SELECT id, appointment_id, patient_id, service_id, amount::FLOAT8 AS amount,
		       method, status, created_at, paid_at
		FROM payment.payment
		WHERE ($1::BIGINT IS NULL OR patient_id = $1)
		  AND ($2::TEXT IS NULL OR status = $2)
		  AND ($3::TIMESTAMPTZ IS NULL OR created_at >= $3)
		  AND ($4::TIMESTAMPTZ IS NULL OR created_at < $4)
		ORDER BY created_at DESC, id DESC`

	rows := make([]models.Payment, 0)
	if err := r.db.SelectContext(ctx, &rows, query, f.PatientID, f.Status, f.From, f.To); err != nil {
		return nil, fmt.Errorf("list payments: %w", postgres.WrapDB(err))
	}

	return models.ToDomainList(rows), nil
}
