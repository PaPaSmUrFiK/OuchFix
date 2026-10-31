package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/PaPaSmUrFiK/OuchFix/internal/appointment/domain"
	"github.com/PaPaSmUrFiK/OuchFix/internal/appointment/repository/postgres/models"
	"github.com/PaPaSmUrFiK/OuchFix/pkg/storage/postgres"
)

// InTx выполняет fn в одной транзакции; методы репозитория внутри fn берут её из ctx.
func (r *AppointmentRepository) InTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return postgres.InTx(ctx, r.db, fn)
}

// LockDoctorSchedule — блокировка расписания врача до конца транзакции, чтобы
// параллельные createSlots/updateSlots не создали пересекающиеся слоты.
func (r *AppointmentRepository) LockDoctorSchedule(ctx context.Context, doctorID int64) error {
	const query = `SELECT pg_advisory_xact_lock($1)`
	if _, err := postgres.Querier(ctx, r.db).ExecContext(ctx, query, doctorID); err != nil {
		return fmt.Errorf("lock schedule of doctor %d: %w", doctorID, postgres.WrapDB(err))
	}
	return nil
}

// ListActiveSlots возвращает неотменённые слоты врача, пересекающие [from, to).
func (r *AppointmentRepository) ListActiveSlots(ctx context.Context, doctorID int64, from, to time.Time) ([]domain.Slot, error) {
	const query = `
		SELECT id, doctor_id, starts_at, ends_at, status
		FROM appointment.schedule_slot
		WHERE doctor_id = $1 AND status <> 'cancelled' AND starts_at < $3 AND ends_at > $2
		ORDER BY starts_at, id`

	rows := make([]models.Slot, 0)
	if err := sqlx.SelectContext(ctx, postgres.Querier(ctx, r.db), &rows, query, doctorID, from, to); err != nil {
		return nil, fmt.Errorf("list active slots of doctor %d: %w", doctorID, postgres.WrapDB(err))
	}
	return models.ToDomainList[models.Slot, domain.Slot](rows), nil
}

// InsertSlots сохраняет новые слоты и возвращает их с id.
func (r *AppointmentRepository) InsertSlots(ctx context.Context, slots []domain.Slot) ([]domain.Slot, error) {
	const query = `
		INSERT INTO appointment.schedule_slot (doctor_id, starts_at, ends_at, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id, doctor_id, starts_at, ends_at, status`

	q := postgres.Querier(ctx, r.db)
	out := make([]domain.Slot, 0, len(slots))
	for _, s := range slots {
		var row models.Slot
		if err := sqlx.GetContext(ctx, q, &row, query, s.DoctorID, s.StartsAt, s.EndsAt, string(s.Status)); err != nil {
			return nil, fmt.Errorf("insert slot: %w", postgres.WrapDB(err))
		}
		out = append(out, row.ToDomain())
	}
	return out, nil
}

// GetSlotsForUpdate читает слоты по id с блокировкой строк до конца транзакции.
// Отсутствующие id в ответе просто не будут.
func (r *AppointmentRepository) GetSlotsForUpdate(ctx context.Context, ids []int64) ([]domain.Slot, error) {
	const query = `
		SELECT id, doctor_id, starts_at, ends_at, status
		FROM appointment.schedule_slot
		WHERE id = ANY($1)
		ORDER BY id
		FOR UPDATE`

	rows := make([]models.Slot, 0)
	if err := sqlx.SelectContext(ctx, postgres.Querier(ctx, r.db), &rows, query, ids); err != nil {
		return nil, fmt.Errorf("get slots for update: %w", postgres.WrapDB(err))
	}
	return models.ToDomainList[models.Slot, domain.Slot](rows), nil
}

// UpdateSlot сохраняет время и статус слота.
func (r *AppointmentRepository) UpdateSlot(ctx context.Context, slot domain.Slot) error {
	const query = `
		UPDATE appointment.schedule_slot
		SET starts_at = $2, ends_at = $3, status = $4
		WHERE id = $1`

	res, err := postgres.Querier(ctx, r.db).ExecContext(ctx, query, slot.ID, slot.StartsAt, slot.EndsAt, string(slot.Status))
	if err != nil {
		return fmt.Errorf("update slot %d: %w", slot.ID, postgres.WrapDB(err))
	}
	return postgres.RequireAffected(res, fmt.Sprintf("slot %d", slot.ID))
}
