package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/PaPaSmUrFiK/OuchFix/internal/payment/domain"
	"github.com/PaPaSmUrFiK/OuchFix/pkg/client/appointment"
	"github.com/PaPaSmUrFiK/OuchFix/pkg/client/catalog"
	pkgerrors "github.com/PaPaSmUrFiK/OuchFix/pkg/errors"
)

type PaymentRepository interface {
	List(ctx context.Context, f domain.Filter) ([]domain.Payment, error)
	HasPaid(ctx context.Context, appointmentID int64) (bool, error)
	Insert(ctx context.Context, p domain.Payment) (*domain.Payment, error)
	Summary(ctx context.Context, from, to time.Time) (*domain.Summary, error)
}

// CatalogClient — то, что Payment спрашивает у Catalog (CatalogGrpc).
type CatalogClient interface {
	GetPatient(ctx context.Context, patientID int64) (*catalog.PatientReply, error)
	GetMedicalService(ctx context.Context, serviceID int64) (*catalog.ServiceReply, error)
}

// AppointmentClient — то, что Payment спрашивает у Appointment (AppointmentGrpc).
// createPayment проверяет по нему, что приём проведён и чей он.
type AppointmentClient interface {
	GetAppointment(ctx context.Context, appointmentID int64) (*appointment.AppointmentReply, error)
}

type PaymentService struct {
	repo         PaymentRepository
	catalog      CatalogClient
	appointments AppointmentClient
	location     *time.Location
	now          func() time.Time
}

func NewPaymentService(repo PaymentRepository, catalog CatalogClient, appointments AppointmentClient, location *time.Location) *PaymentService {
	return &PaymentService{repo: repo, catalog: catalog, appointments: appointments, location: location, now: time.Now}
}

// ListInput — сырые параметры payments: status в UPPERCASE GraphQL, даты YYYY-MM-DD.
type ListInput struct {
	PatientID *int64
	Status    *string
	From      *string
	To        *string
}

// ListPayments — платежи с фильтром по пациенту, статусу и дням создания
// [from, to] включительно в поясе медцентра.
func (s *PaymentService) ListPayments(ctx context.Context, in ListInput) ([]domain.Payment, error) {
	f := domain.Filter{PatientID: in.PatientID}
	if in.Status != nil {
		st := domain.Status(strings.ToLower(*in.Status))
		if !st.Valid() {
			return nil, fmt.Errorf("%w: неизвестный статус платежа %q", pkgerrors.ErrInvalidArgument, *in.Status)
		}
		f.Status = &st
	}
	if in.From != nil {
		from, err := s.parseDay("from", *in.From)
		if err != nil {
			return nil, err
		}
		f.From = &from
	}
	if in.To != nil {
		last, err := s.parseDay("to", *in.To)
		if err != nil {
			return nil, err
		}
		end := last.AddDate(0, 0, 1)
		f.To = &end
	}
	if f.From != nil && f.To != nil && !f.From.Before(*f.To) {
		return nil, fmt.Errorf("%w: from должна быть не позже to", pkgerrors.ErrInvalidArgument)
	}
	if in.PatientID != nil {
		if err := pkgerrors.ValidateID("patientId", *in.PatientID); err != nil {
			return nil, err
		}
		if _, err := s.catalog.GetPatient(ctx, *in.PatientID); err != nil {
			return nil, fmt.Errorf("check patient %d: %w", *in.PatientID, err)
		}
	}

	return s.repo.List(ctx, f)
}
