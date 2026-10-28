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

type AppointmentRepository struct {
	db *sqlx.DB
}

func NewAppointmentRepository(db *sqlx.DB) *AppointmentRepository {
	return &AppointmentRepository{db: db}
}

// ListSlots возвращает слоты врача со статусом status, начинающиеся в [from, to).
func (r *AppointmentRepository) ListSlots(
	ctx context.Context,
	doctorID int64,
	status domain.SlotStatus,
	from, to time.Time,
) ([]domain.Slot, error) {
	const query = `
		SELECT id, doctor_id, starts_at, ends_at, status
		FROM appointment.schedule_slot
		WHERE doctor_id = $1 AND status = $2 AND starts_at >= $3 AND starts_at < $4
		ORDER BY starts_at, id`

	rows := make([]models.Slot, 0)
	if err := r.db.SelectContext(ctx, &rows, query, doctorID, string(status), from, to); err != nil {
		return nil, fmt.Errorf("list slots of doctor %d: %w", doctorID, postgres.WrapDB(err))
	}

	return models.ToDomainList[models.Slot, domain.Slot](rows), nil
}

const appointmentSelect = `
	SELECT id, patient_id, doctor_id, slot_id, service_id, status, queue_type,
	       scheduled_at, conclusion, created_at
	FROM appointment.appointment`

func (r *AppointmentRepository) GetAppointment(ctx context.Context, id int64) (*domain.Appointment, error) {
	const query = appointmentSelect + ` WHERE id = $1`

	var row models.Appointment
	if err := r.db.GetContext(ctx, &row, query, id); err != nil {
		return nil, fmt.Errorf("get appointment %d: %w", id, postgres.WrapDB(err))
	}

	appointment := row.ToDomain()
	return &appointment, nil
}

// ListByPatient возвращает приёмы пациента; status == nil — без фильтра.
func (r *AppointmentRepository) ListByPatient(
	ctx context.Context,
	patientID int64,
	status *domain.Status,
) ([]domain.Appointment, error) {
	const query = appointmentSelect + `
		WHERE patient_id = $1 AND ($2::VARCHAR IS NULL OR status = $2)
		ORDER BY scheduled_at DESC NULLS LAST, created_at DESC, id DESC`

	var statusArg *string
	if status != nil {
		s := string(*status)
		statusArg = &s
	}

	rows := make([]models.Appointment, 0)
	if err := r.db.SelectContext(ctx, &rows, query, patientID, statusArg); err != nil {
		return nil, fmt.Errorf("list appointments of patient %d: %w", patientID, postgres.WrapDB(err))
	}

	return models.ToDomainList[models.Appointment, domain.Appointment](rows), nil
}

func (r *AppointmentRepository) ListDiagnoses(ctx context.Context, appointmentIDs []int64) ([]domain.Diagnosis, error) {
	const query = `
		SELECT id, appointment_id, code, description, created_at
		FROM appointment.diagnosis
		WHERE appointment_id = ANY($1)
		ORDER BY appointment_id, created_at, id`

	rows := make([]models.Diagnosis, 0)
	if err := r.db.SelectContext(ctx, &rows, query, appointmentIDs); err != nil {
		return nil, fmt.Errorf("list diagnoses: %w", postgres.WrapDB(err))
	}

	return models.ToDomainList[models.Diagnosis, domain.Diagnosis](rows), nil
}

func (r *AppointmentRepository) ListPrescriptions(ctx context.Context, appointmentIDs []int64) ([]domain.Prescription, error) {
	const query = `
		SELECT id, appointment_id, text, medication_id, created_at
		FROM appointment.prescription
		WHERE appointment_id = ANY($1)
		ORDER BY appointment_id, created_at, id`

	rows := make([]models.Prescription, 0)
	if err := r.db.SelectContext(ctx, &rows, query, appointmentIDs); err != nil {
		return nil, fmt.Errorf("list prescriptions: %w", postgres.WrapDB(err))
	}

	return models.ToDomainList[models.Prescription, domain.Prescription](rows), nil
}
