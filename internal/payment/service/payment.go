package service

import (
	"context"
	"fmt"
	"time"

	"github.com/PaPaSmUrFiK/OuchFix/internal/payment/domain"
	"github.com/PaPaSmUrFiK/OuchFix/pkg/client/appointment"
	"github.com/PaPaSmUrFiK/OuchFix/pkg/client/catalog"
	pkgerrors "github.com/PaPaSmUrFiK/OuchFix/pkg/errors"
)

type PaymentRepository interface {
	List(ctx context.Context, patientID *int64) ([]domain.Payment, error)
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

// ListPayments возвращает платежи пациента; patientID == nil — все платежи.
func (s *PaymentService) ListPayments(ctx context.Context, patientID *int64) ([]domain.Payment, error) {
	if patientID != nil {
		if err := pkgerrors.ValidateID("patientId", *patientID); err != nil {
			return nil, err
		}
		if _, err := s.catalog.GetPatient(ctx, *patientID); err != nil {
			return nil, fmt.Errorf("check patient %d: %w", *patientID, err)
		}
	}

	return s.repo.List(ctx, patientID)
}
