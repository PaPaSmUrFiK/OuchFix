package models

import (
	"database/sql"
	"time"

	"github.com/PaPaSmUrFiK/OuchFix/internal/appointment/domain"
)

type Slot struct {
	ID       int64     `db:"id"`
	DoctorID int64     `db:"doctor_id"`
	StartsAt time.Time `db:"starts_at"`
	EndsAt   time.Time `db:"ends_at"`
	Status   string    `db:"status"`
}

func (m Slot) ToDomain() domain.Slot {
	return domain.Slot{
		ID:       m.ID,
		DoctorID: m.DoctorID,
		StartsAt: m.StartsAt,
		EndsAt:   m.EndsAt,
		Status:   domain.SlotStatus(m.Status),
	}
}

type Appointment struct {
	ID          int64          `db:"id"`
	PatientID   int64          `db:"patient_id"`
	DoctorID    int64          `db:"doctor_id"`
	SlotID      sql.NullInt64  `db:"slot_id"`
	ServiceID   sql.NullInt64  `db:"service_id"`
	Status      string         `db:"status"`
	QueueType   string         `db:"queue_type"`
	ScheduledAt sql.NullTime   `db:"scheduled_at"`
	Conclusion  sql.NullString `db:"conclusion"`
	CreatedAt   time.Time      `db:"created_at"`
}

func (m Appointment) ToDomain() domain.Appointment {
	return domain.Appointment{
		ID:          m.ID,
		PatientID:   m.PatientID,
		DoctorID:    m.DoctorID,
		SlotID:      int64Ptr(m.SlotID),
		ServiceID:   int64Ptr(m.ServiceID),
		Status:      domain.Status(m.Status),
		VisitType:   domain.VisitType(m.QueueType),
		ScheduledAt: timePtr(m.ScheduledAt),
		Conclusion:  stringPtr(m.Conclusion),
		CreatedAt:   m.CreatedAt,
	}
}

type Diagnosis struct {
	ID            int64          `db:"id"`
	AppointmentID int64          `db:"appointment_id"`
	Code          sql.NullString `db:"code"`
	Description   string         `db:"description"`
	CreatedAt     time.Time      `db:"created_at"`
}

func (m Diagnosis) ToDomain() domain.Diagnosis {
	return domain.Diagnosis{
		ID:            m.ID,
		AppointmentID: m.AppointmentID,
		Code:          stringPtr(m.Code),
		Description:   m.Description,
		CreatedAt:     m.CreatedAt,
	}
}

type Prescription struct {
	ID            int64         `db:"id"`
	AppointmentID int64         `db:"appointment_id"`
	Text          string        `db:"text"`
	MedicationID  sql.NullInt64 `db:"medication_id"`
	CreatedAt     time.Time     `db:"created_at"`
}

func (m Prescription) ToDomain() domain.Prescription {
	return domain.Prescription{
		ID:            m.ID,
		AppointmentID: m.AppointmentID,
		Text:          m.Text,
		MedicationID:  int64Ptr(m.MedicationID),
		CreatedAt:     m.CreatedAt,
	}
}

func ToDomainList[M interface{ ToDomain() D }, D any](rows []M) []D {
	out := make([]D, 0, len(rows))
	for _, row := range rows {
		out = append(out, row.ToDomain())
	}
	return out
}

func int64Ptr(v sql.NullInt64) *int64 {
	if !v.Valid {
		return nil
	}
	i := v.Int64
	return &i
}

func stringPtr(v sql.NullString) *string {
	if !v.Valid {
		return nil
	}
	s := v.String
	return &s
}

func timePtr(v sql.NullTime) *time.Time {
	if !v.Valid {
		return nil
	}
	t := v.Time
	return &t
}
