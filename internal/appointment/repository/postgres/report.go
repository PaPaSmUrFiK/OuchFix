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

// visitTime — время приёма: по записи — время слота, в живой очереди — постановка в очередь.
const visitTime = `COALESCE(scheduled_at, created_at)`

// orderBy — только из белого списка, в SQL не попадает пользовательский текст.
var orderBy = map[domain.Sort]string{
	domain.SortScheduledAtAsc:  visitTime + ` ASC, id ASC`,
	domain.SortScheduledAtDesc: visitTime + ` DESC, id DESC`,
}

func (r *AppointmentRepository) ListAppointments(ctx context.Context, f domain.AppointmentFilter) ([]domain.Appointment, error) {
	order, ok := orderBy[f.Sort]
	if !ok {
		return nil, fmt.Errorf("unknown sort %q", f.Sort)
	}
	query := appointmentSelect + `
		WHERE ($1::BIGINT IS NULL OR doctor_id = $1)
		  AND ($2::VARCHAR IS NULL OR status = $2)
		  AND ($3::TIMESTAMPTZ IS NULL OR ` + visitTime + ` >= $3)
		  AND ($4::TIMESTAMPTZ IS NULL OR ` + visitTime + ` < $4)
		ORDER BY ` + order

	var status *string
	if f.Status != nil {
		s := string(*f.Status)
		status = &s
	}

	rows := make([]models.Appointment, 0)
	if err := sqlx.SelectContext(ctx, r.db, &rows, query, f.DoctorID, status, f.From, f.To); err != nil {
		return nil, fmt.Errorf("list appointments: %w", postgres.WrapDB(err))
	}
	return models.ToDomainList[models.Appointment, domain.Appointment](rows), nil
}

// DoctorLoad — число неотменённых приёмов каждого врача за [from, to), по убыванию.
func (r *AppointmentRepository) DoctorLoad(ctx context.Context, from, to time.Time) ([]domain.DoctorLoad, error) {
	const query = `
		SELECT doctor_id, COUNT(*) AS cnt
		FROM appointment.appointment
		WHERE status <> 'cancelled' AND ` + visitTime + ` >= $1 AND ` + visitTime + ` < $2
		GROUP BY doctor_id
		ORDER BY cnt DESC, doctor_id`

	var rows []struct {
		DoctorID int64 `db:"doctor_id"`
		Count    int   `db:"cnt"`
	}
	if err := sqlx.SelectContext(ctx, r.db, &rows, query, from, to); err != nil {
		return nil, fmt.Errorf("doctor load: %w", postgres.WrapDB(err))
	}
	out := make([]domain.DoctorLoad, 0, len(rows))
	for _, row := range rows {
		out = append(out, domain.DoctorLoad{DoctorID: row.DoctorID, AppointmentCount: row.Count})
	}
	return out, nil
}

// PopularServices — услуги проведённых за [from, to) приёмов по числу оказаний, не больше limit.
func (r *AppointmentRepository) PopularServices(ctx context.Context, from, to time.Time, limit int) ([]domain.ServiceStat, error) {
	const query = `
		SELECT service_id, COUNT(*) AS cnt
		FROM appointment.appointment
		WHERE status = 'completed' AND service_id IS NOT NULL
		  AND ` + visitTime + ` >= $1 AND ` + visitTime + ` < $2
		GROUP BY service_id
		ORDER BY cnt DESC, service_id
		LIMIT $3`

	var rows []struct {
		ServiceID int64 `db:"service_id"`
		Count     int   `db:"cnt"`
	}
	if err := sqlx.SelectContext(ctx, r.db, &rows, query, from, to, limit); err != nil {
		return nil, fmt.Errorf("popular services: %w", postgres.WrapDB(err))
	}
	out := make([]domain.ServiceStat, 0, len(rows))
	for _, row := range rows {
		out = append(out, domain.ServiceStat{ServiceID: row.ServiceID, BookingCount: row.Count})
	}
	return out, nil
}
