package postgres

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/PaPaSmUrFiK/OuchFix/internal/catalog/domain"
	"github.com/PaPaSmUrFiK/OuchFix/internal/catalog/repository/postgres/models"
	"github.com/PaPaSmUrFiK/OuchFix/pkg/storage/postgres"
)

const medicationReturning = ` RETURNING id, name, form, unit, stock_qty::FLOAT8 AS stock_qty`

func (r *CatalogRepository) InTx(ctx context.Context, fn func(context.Context) error) error {
	return postgres.InTx(ctx, r.db, fn)
}

// UpsertMedication создаёт препарат или обновляет существующий с тем же названием.
func (r *CatalogRepository) UpsertMedication(ctx context.Context, m domain.Medication) (*domain.Medication, error) {
	const query = `
		INSERT INTO catalog.medication (name, form, unit, stock_qty)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (name) DO UPDATE
		SET form = EXCLUDED.form, unit = EXCLUDED.unit, stock_qty = EXCLUDED.stock_qty` + medicationReturning

	var row models.Medication
	if err := sqlx.GetContext(ctx, postgres.Querier(ctx, r.db), &row, query, m.Name, m.Form, m.Unit, m.StockQty); err != nil {
		return nil, fmt.Errorf("upsert medication %q: %w", m.Name, postgres.WrapDB(err))
	}
	out := row.ToDomain()
	return &out, nil
}

// TakeStock атомарно списывает qty, только если остатка хватает. ErrNotFound — препарата
// нет или остатка мало (различает сервис).
func (r *CatalogRepository) TakeStock(ctx context.Context, medicationID int64, qty float64) (*domain.Medication, error) {
	const query = `
		UPDATE catalog.medication SET stock_qty = stock_qty - $2
		WHERE id = $1 AND stock_qty >= $2` + medicationReturning

	var row models.Medication
	if err := sqlx.GetContext(ctx, postgres.Querier(ctx, r.db), &row, query, medicationID, qty); err != nil {
		return nil, fmt.Errorf("take stock of medication %d: %w", medicationID, postgres.WrapDB(err))
	}
	out := row.ToDomain()
	return &out, nil
}

func (r *CatalogRepository) InsertDispense(ctx context.Context, d domain.Dispense) error {
	const query = `
		INSERT INTO catalog.medication_dispense (medication_id, patient_id, appointment_id, quantity)
		VALUES ($1, $2, $3, $4)`

	if _, err := postgres.Querier(ctx, r.db).ExecContext(ctx, query, d.MedicationID, d.PatientID, d.AppointmentID, d.Quantity); err != nil {
		return fmt.Errorf("insert dispense: %w", postgres.WrapDB(err))
	}
	return nil
}
