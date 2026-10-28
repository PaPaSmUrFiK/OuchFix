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

// List возвращает платежи; patientID == nil — по всем пациентам.
func (r *PaymentRepository) List(ctx context.Context, patientID *int64) ([]domain.Payment, error) {
	const query = `
		SELECT id, appointment_id, patient_id, service_id, amount::FLOAT8 AS amount,
		       method, status, created_at, paid_at
		FROM payment.payment
		WHERE $1::BIGINT IS NULL OR patient_id = $1
		ORDER BY created_at DESC, id DESC`

	rows := make([]models.Payment, 0)
	if err := r.db.SelectContext(ctx, &rows, query, patientID); err != nil {
		return nil, fmt.Errorf("list payments: %w", postgres.WrapDB(err))
	}

	return models.ToDomainList(rows), nil
}
