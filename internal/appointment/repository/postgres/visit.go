package postgres

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/PaPaSmUrFiK/OuchFix/internal/appointment/domain"
	"github.com/PaPaSmUrFiK/OuchFix/internal/appointment/repository/postgres/models"
	"github.com/PaPaSmUrFiK/OuchFix/pkg/storage/postgres"
)

const appointmentReturning = `
	RETURNING id, patient_id, doctor_id, slot_id, service_id, status, queue_type,
	          scheduled_at, conclusion, created_at`

// GetAppointmentForUpdate читает приём с блокировкой строки до конца транзакции.
func (r *AppointmentRepository) GetAppointmentForUpdate(ctx context.Context, id int64) (*domain.Appointment, error) {
	const query = appointmentSelect + ` WHERE id = $1 FOR UPDATE`

	var row models.Appointment
	if err := sqlx.GetContext(ctx, postgres.Querier(ctx, r.db), &row, query, id); err != nil {
		return nil, fmt.Errorf("get appointment %d for update: %w", id, postgres.WrapDB(err))
	}
	a := row.ToDomain()
	return &a, nil
}

// HasQueued — стоит ли пациент уже в живой очереди к врачу.
func (r *AppointmentRepository) HasQueued(ctx context.Context, patientID, doctorID int64) (bool, error) {
	const query = `
		SELECT EXISTS (
			SELECT 1 FROM appointment.appointment
			WHERE patient_id = $1 AND doctor_id = $2 AND status = 'queue'
		)`

	var exists bool
	if err := sqlx.GetContext(ctx, postgres.Querier(ctx, r.db), &exists, query, patientID, doctorID); err != nil {
		return false, fmt.Errorf("check queue of patient %d: %w", patientID, postgres.WrapDB(err))
	}
	return exists, nil
}

func (r *AppointmentRepository) InsertAppointment(ctx context.Context, a domain.Appointment) (*domain.Appointment, error) {
	const query = `
		INSERT INTO appointment.appointment
			(patient_id, doctor_id, slot_id, service_id, status, queue_type, scheduled_at, conclusion)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)` + appointmentReturning

	var row models.Appointment
	err := sqlx.GetContext(ctx, postgres.Querier(ctx, r.db), &row, query,
		a.PatientID, a.DoctorID, a.SlotID, a.ServiceID, string(a.Status), string(a.VisitType), a.ScheduledAt, a.Conclusion)
	if err != nil {
		return nil, fmt.Errorf("insert appointment: %w", postgres.WrapDB(err))
	}
	out := row.ToDomain()
	return &out, nil
}

// UpdateAppointment сохраняет изменяемые поля приёма: слот, услугу, статус, время и заключение.
func (r *AppointmentRepository) UpdateAppointment(ctx context.Context, a domain.Appointment) error {
	const query = `
		UPDATE appointment.appointment
		SET slot_id = $2, service_id = $3, status = $4, scheduled_at = $5, conclusion = $6
		WHERE id = $1`

	res, err := postgres.Querier(ctx, r.db).ExecContext(ctx, query,
		a.ID, a.SlotID, a.ServiceID, string(a.Status), a.ScheduledAt, a.Conclusion)
	if err != nil {
		return fmt.Errorf("update appointment %d: %w", a.ID, postgres.WrapDB(err))
	}
	return postgres.RequireAffected(res, fmt.Sprintf("appointment %d", a.ID))
}

func (r *AppointmentRepository) InsertDiagnosis(ctx context.Context, d domain.Diagnosis) (*domain.Diagnosis, error) {
	const query = `
		INSERT INTO appointment.diagnosis (appointment_id, code, description)
		VALUES ($1, $2, $3)
		RETURNING id, appointment_id, code, description, created_at`

	var row models.Diagnosis
	if err := sqlx.GetContext(ctx, postgres.Querier(ctx, r.db), &row, query, d.AppointmentID, d.Code, d.Description); err != nil {
		return nil, fmt.Errorf("insert diagnosis: %w", postgres.WrapDB(err))
	}
	out := row.ToDomain()
	return &out, nil
}

func (r *AppointmentRepository) InsertPrescription(ctx context.Context, p domain.Prescription) (*domain.Prescription, error) {
	const query = `
		INSERT INTO appointment.prescription (appointment_id, text, medication_id)
		VALUES ($1, $2, $3)
		RETURNING id, appointment_id, text, medication_id, created_at`

	var row models.Prescription
	if err := sqlx.GetContext(ctx, postgres.Querier(ctx, r.db), &row, query, p.AppointmentID, p.Text, p.MedicationID); err != nil {
		return nil, fmt.Errorf("insert prescription: %w", postgres.WrapDB(err))
	}
	out := row.ToDomain()
	return &out, nil
}
