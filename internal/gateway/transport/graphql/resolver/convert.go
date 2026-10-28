package resolver

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	appointmentdomain "github.com/PaPaSmUrFiK/OuchFix/internal/appointment/domain"
	catalogdomain "github.com/PaPaSmUrFiK/OuchFix/internal/catalog/domain"
	"github.com/PaPaSmUrFiK/OuchFix/internal/gateway/transport/graphql/model"
	notificationdomain "github.com/PaPaSmUrFiK/OuchFix/internal/notification/domain"
	paymentdomain "github.com/PaPaSmUrFiK/OuchFix/internal/payment/domain"
	pkgerrors "github.com/PaPaSmUrFiK/OuchFix/pkg/errors"
)

const (
	dateLayout = "2006-01-02"
	timeLayout = time.RFC3339
)

// parseID переводит GraphQL ID в id сервиса. Нечисловой id — ошибка аргумента.
func parseID(name, raw string) (int64, error) {
	id, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%w: %s должен быть числом, получено %q", pkgerrors.ErrInvalidArgument, name, raw)
	}
	return id, nil
}

func parseOptionalID(name string, raw *string) (*int64, error) {
	if raw == nil {
		return nil, nil
	}
	id, err := parseID(name, *raw)
	if err != nil {
		return nil, err
	}
	return &id, nil
}

func formatID(id int64) string {
	return strconv.FormatInt(id, 10)
}

func formatOptionalID(id *int64) *string {
	if id == nil {
		return nil
	}
	s := formatID(*id)
	return &s
}

func (r *Resolver) formatTime(t time.Time) string {
	return t.In(r.location).Format(timeLayout)
}

func (r *Resolver) formatOptionalTime(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := r.formatTime(*t)
	return &s
}

// Enum'ы: в БД значения в нижнем регистре (scheduled, walk_in), в GraphQL — в верхнем.
func toEnum[E ~string, D ~string](value D) E {
	return E(strings.ToUpper(string(value)))
}

func fromEnum[D ~string, E ~string](value *E) *D {
	if value == nil {
		return nil
	}
	d := D(strings.ToLower(string(*value)))
	return &d
}

func mapList[S any, D any](items []S, convert func(*S) *D) []*D {
	out := make([]*D, 0, len(items))
	for i := range items {
		out = append(out, convert(&items[i]))
	}
	return out
}

func specialtyToModel(s *catalogdomain.Specialty) *model.Specialty {
	return &model.Specialty{ID: formatID(s.ID), Name: s.Name}
}

func doctorToModel(d *catalogdomain.Doctor) *model.Doctor {
	return &model.Doctor{
		ID:        formatID(d.ID),
		FullName:  d.FullName,
		Office:    d.Office,
		Specialty: specialtyToModel(&d.Specialty),
	}
}

func medicalServiceToModel(s *catalogdomain.MedicalService) *model.MedicalService {
	return &model.MedicalService{
		ID:          formatID(s.ID),
		Name:        s.Name,
		Price:       s.Price,
		DurationMin: s.DurationMin,
	}
}

func medicationToModel(m *catalogdomain.Medication) *model.Medication {
	return &model.Medication{
		ID:       formatID(m.ID),
		Name:     m.Name,
		Form:     m.Form,
		Unit:     m.Unit,
		StockQty: m.StockQty,
	}
}

func (r *Resolver) slotToModel(s *appointmentdomain.Slot) *model.ScheduleSlot {
	return &model.ScheduleSlot{
		ID:       formatID(s.ID),
		DoctorID: formatID(s.DoctorID),
		StartsAt: r.formatTime(s.StartsAt),
		EndsAt:   r.formatTime(s.EndsAt),
		Status:   toEnum[model.SlotStatus](s.Status),
	}
}

func (r *Resolver) appointmentToModel(a *appointmentdomain.Appointment) *model.Appointment {
	return &model.Appointment{
		ID:          formatID(a.ID),
		PatientID:   formatID(a.PatientID),
		DoctorID:    formatID(a.DoctorID),
		SlotID:      formatOptionalID(a.SlotID),
		ServiceID:   formatOptionalID(a.ServiceID),
		Status:      toEnum[model.AppointmentStatus](a.Status),
		VisitType:   toEnum[model.VisitType](a.VisitType),
		ScheduledAt: r.formatOptionalTime(a.ScheduledAt),
		Conclusion:  a.Conclusion,
		Diagnoses: mapList(a.Diagnoses, func(d *appointmentdomain.Diagnosis) *model.Diagnosis {
			return &model.Diagnosis{ID: formatID(d.ID), Code: d.Code, Description: d.Description}
		}),
		Prescriptions: mapList(a.Prescriptions, func(p *appointmentdomain.Prescription) *model.Prescription {
			return &model.Prescription{ID: formatID(p.ID), Text: p.Text, MedicationID: formatOptionalID(p.MedicationID)}
		}),
	}
}

func (r *Resolver) paymentToModel(p *paymentdomain.Payment) *model.Payment {
	return &model.Payment{
		ID:            formatID(p.ID),
		AppointmentID: formatID(p.AppointmentID),
		PatientID:     formatID(p.PatientID),
		ServiceID:     formatOptionalID(p.ServiceID),
		Amount:        p.Amount,
		Method:        toEnum[model.PaymentMethod](p.Method),
		Status:        toEnum[model.PaymentStatus](p.Status),
		PaidAt:        r.formatOptionalTime(p.PaidAt),
	}
}

func (r *Resolver) notificationToModel(n *notificationdomain.Notification) *model.Notification {
	return &model.Notification{
		ID:          formatID(n.ID),
		Title:       n.Title,
		Body:        n.Body,
		Status:      toEnum[model.NotificationStatus](n.Status),
		RelatedType: n.RelatedType,
		RelatedID:   formatOptionalID(n.RelatedID),
		CreatedAt:   r.formatTime(n.CreatedAt),
	}
}

// notImplemented — операции следующих лабораторных: схема уже есть, логики ещё нет.
func notImplemented(field, lab string) error {
	return fmt.Errorf("%w: %s появится в %s", pkgerrors.ErrNotImplemented, field, lab)
}
