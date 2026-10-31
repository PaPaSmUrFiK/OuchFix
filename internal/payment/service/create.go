package service

import (
	"context"
	"fmt"
	"time"

	"github.com/PaPaSmUrFiK/OuchFix/internal/payment/domain"
	pkgerrors "github.com/PaPaSmUrFiK/OuchFix/pkg/errors"
)

const (
	dateLayout      = "2006-01-02"
	maxReportPeriod = 366 * 24 * time.Hour
	// appointmentCompleted — статус приёма в AppointmentGrpc.
	appointmentCompleted = "completed"
)

type CreateInput struct {
	AppointmentID int64
	PatientID     int64
	ServiceID     *int64
	AmountKopecks int64
	Method        domain.Method
}

// CreatePayment — оплата проведённого приёма по цене его услуги. Приём оплачивается
// один раз: повторная оплата — CONFLICT.
func (s *PaymentService) CreatePayment(ctx context.Context, in CreateInput) (*domain.Payment, error) {
	if err := pkgerrors.ValidateID("appointmentId", in.AppointmentID); err != nil {
		return nil, err
	}
	if err := pkgerrors.ValidateID("patientId", in.PatientID); err != nil {
		return nil, err
	}
	if in.ServiceID != nil {
		if err := pkgerrors.ValidateID("serviceId", *in.ServiceID); err != nil {
			return nil, err
		}
	}
	if !in.Method.Valid() {
		return nil, fmt.Errorf("%w: неизвестный способ оплаты %q", pkgerrors.ErrInvalidArgument, in.Method)
	}
	if in.AmountKopecks <= 0 {
		return nil, fmt.Errorf("%w: amountKopecks должна быть больше нуля", pkgerrors.ErrInvalidArgument)
	}

	appt, err := s.appointments.GetAppointment(ctx, in.AppointmentID)
	if err != nil {
		return nil, fmt.Errorf("appointment %d: %w", in.AppointmentID, err)
	}
	if appt.PatientID != in.PatientID {
		return nil, fmt.Errorf("%w: приём %d не принадлежит пациенту %d", pkgerrors.ErrInvalidArgument, appt.ID, in.PatientID)
	}
	if appt.Status != appointmentCompleted {
		return nil, fmt.Errorf("%w: оплатить можно только проведённый приём, статус %q", pkgerrors.ErrFailedPrecondition, appt.Status)
	}
	if appt.ServiceID == 0 {
		return nil, fmt.Errorf("%w: у приёма не указана услуга", pkgerrors.ErrFailedPrecondition)
	}
	if in.ServiceID != nil && *in.ServiceID != appt.ServiceID {
		return nil, fmt.Errorf("%w: в приёме оказана услуга %d, а не %d", pkgerrors.ErrInvalidArgument, appt.ServiceID, *in.ServiceID)
	}
	svc, err := s.catalog.GetMedicalService(ctx, appt.ServiceID)
	if err != nil {
		return nil, fmt.Errorf("medical service %d: %w", appt.ServiceID, err)
	}
	if in.AmountKopecks != svc.PriceKopecks {
		return nil, fmt.Errorf("%w: сумма %d коп. не совпадает с ценой услуги %d коп.", pkgerrors.ErrInvalidArgument, in.AmountKopecks, svc.PriceKopecks)
	}

	paid, err := s.repo.HasPaid(ctx, in.AppointmentID)
	if err != nil {
		return nil, err
	}
	if paid {
		return nil, fmt.Errorf("%w: приём %d уже оплачен", pkgerrors.ErrConflict, in.AppointmentID)
	}

	now := s.now()
	serviceID := appt.ServiceID
	return s.repo.Insert(ctx, domain.Payment{
		AppointmentID: in.AppointmentID,
		PatientID:     in.PatientID,
		ServiceID:     &serviceID,
		AmountKopecks: svc.PriceKopecks,
		Method:        in.Method,
		Status:        domain.StatusPaid,
		PaidAt:        &now,
	})
}

// PaymentSummary — выручка за дни [from, to] включительно в поясе медцентра.
func (s *PaymentService) PaymentSummary(ctx context.Context, from, to string) (*domain.Summary, error) {
	start, err := s.parseDay("from", from)
	if err != nil {
		return nil, err
	}
	last, err := s.parseDay("to", to)
	if err != nil {
		return nil, err
	}
	if last.Before(start) {
		return nil, fmt.Errorf("%w: from должна быть не позже to", pkgerrors.ErrInvalidArgument)
	}
	end := last.AddDate(0, 0, 1)
	if end.Sub(start) > maxReportPeriod {
		return nil, fmt.Errorf("%w: период больше года", pkgerrors.ErrInvalidArgument)
	}
	return s.repo.Summary(ctx, start, end)
}

func (s *PaymentService) parseDay(name, raw string) (time.Time, error) {
	day, err := time.ParseInLocation(dateLayout, raw, s.location)
	if err != nil {
		return time.Time{}, fmt.Errorf("%w: %s должна быть в формате YYYY-MM-DD, получено %q", pkgerrors.ErrInvalidArgument, name, raw)
	}
	return day, nil
}
