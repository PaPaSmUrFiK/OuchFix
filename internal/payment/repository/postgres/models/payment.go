package models

import (
	"database/sql"
	"time"

	"github.com/PaPaSmUrFiK/OuchFix/internal/payment/domain"
)

type Payment struct {
	ID            int64         `db:"id"`
	AppointmentID int64         `db:"appointment_id"`
	PatientID     int64         `db:"patient_id"`
	ServiceID     sql.NullInt64 `db:"service_id"`
	AmountKopecks int64         `db:"amount_kopecks"`
	Method        string        `db:"method"`
	Status        string        `db:"status"`
	CreatedAt     time.Time     `db:"created_at"`
	PaidAt        sql.NullTime  `db:"paid_at"`
}

func (m Payment) ToDomain() domain.Payment {
	p := domain.Payment{
		ID:            m.ID,
		AppointmentID: m.AppointmentID,
		PatientID:     m.PatientID,
		AmountKopecks: m.AmountKopecks,
		Method:        domain.Method(m.Method),
		Status:        domain.Status(m.Status),
		CreatedAt:     m.CreatedAt,
	}
	if m.ServiceID.Valid {
		v := m.ServiceID.Int64
		p.ServiceID = &v
	}
	if m.PaidAt.Valid {
		v := m.PaidAt.Time
		p.PaidAt = &v
	}
	return p
}

func ToDomainList(rows []Payment) []domain.Payment {
	out := make([]domain.Payment, 0, len(rows))
	for _, row := range rows {
		out = append(out, row.ToDomain())
	}
	return out
}
