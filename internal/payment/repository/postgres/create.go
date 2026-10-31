package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/PaPaSmUrFiK/OuchFix/internal/payment/domain"
	"github.com/PaPaSmUrFiK/OuchFix/internal/payment/repository/postgres/models"
	"github.com/PaPaSmUrFiK/OuchFix/pkg/storage/postgres"
)

// HasPaid — оплачен ли приём (частичный уникальный индекс страхует гонку).
func (r *PaymentRepository) HasPaid(ctx context.Context, appointmentID int64) (bool, error) {
	const query = `SELECT EXISTS (SELECT 1 FROM payment.payment WHERE appointment_id = $1 AND status = 'paid')`

	var exists bool
	if err := r.db.GetContext(ctx, &exists, query, appointmentID); err != nil {
		return false, fmt.Errorf("check paid appointment %d: %w", appointmentID, postgres.WrapDB(err))
	}
	return exists, nil
}

func (r *PaymentRepository) Insert(ctx context.Context, p domain.Payment) (*domain.Payment, error) {
	const query = `
		INSERT INTO payment.payment (appointment_id, patient_id, service_id, amount, method, status, paid_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, appointment_id, patient_id, service_id, amount::FLOAT8 AS amount,
		          method, status, created_at, paid_at`

	var row models.Payment
	err := r.db.GetContext(ctx, &row, query,
		p.AppointmentID, p.PatientID, p.ServiceID, p.Amount, string(p.Method), string(p.Status), p.PaidAt)
	if err != nil {
		return nil, fmt.Errorf("insert payment: %w", postgres.WrapDB(err))
	}
	out := row.ToDomain()
	return &out, nil
}

// Summary — оплаченные платежи с paid_at в [from, to).
func (r *PaymentRepository) Summary(ctx context.Context, from, to time.Time) (*domain.Summary, error) {
	const query = `
		SELECT COUNT(*) AS cnt, COALESCE(SUM(amount), 0)::FLOAT8 AS total
		FROM payment.payment
		WHERE status = 'paid' AND paid_at >= $1 AND paid_at < $2`

	var row struct {
		Count int     `db:"cnt"`
		Total float64 `db:"total"`
	}
	if err := r.db.GetContext(ctx, &row, query, from, to); err != nil {
		return nil, fmt.Errorf("payment summary: %w", postgres.WrapDB(err))
	}
	return &domain.Summary{Count: row.Count, Total: row.Total}, nil
}
