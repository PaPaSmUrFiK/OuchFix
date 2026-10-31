package service

import (
	"context"
	"fmt"

	"github.com/PaPaSmUrFiK/OuchFix/internal/payment/domain"
	"github.com/PaPaSmUrFiK/OuchFix/pkg/client/appointment"
	"github.com/PaPaSmUrFiK/OuchFix/pkg/client/catalog"
	pkgerrors "github.com/PaPaSmUrFiK/OuchFix/pkg/errors"
)

type PaymentRepository interface {
	List(ctx context.Context, patientID *int64) ([]domain.Payment, error)
}

// CatalogClient — то, что Payment спрашивает у Catalog (CatalogGrpc).
type CatalogClient interface {
	GetPatient(ctx context.Context, patientID int64) (*catalog.PatientReply, error)
}

// AppointmentClient — то, что Payment спрашивает у Appointment (AppointmentGrpc).
// В ЛР3 только читаем платежи; понадобится createPayment в ЛР4.
type AppointmentClient interface {
	GetAppointment(ctx context.Context, appointmentID int64) (*appointment.AppointmentReply, error)
}

type PaymentService struct {
	repo         PaymentRepository
	catalog      CatalogClient
	appointments AppointmentClient
}

func NewPaymentService(repo PaymentRepository, catalog CatalogClient, appointments AppointmentClient) *PaymentService {
	return &PaymentService{repo: repo, catalog: catalog, appointments: appointments}
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
