package service

import (
	"context"
	"fmt"
	"time"

	"github.com/PaPaSmUrFiK/OuchFix/internal/appointment/domain"
	"github.com/PaPaSmUrFiK/OuchFix/pkg/client/catalog"
	pkgerrors "github.com/PaPaSmUrFiK/OuchFix/pkg/errors"
)

const dateLayout = "2006-01-02"

type AppointmentRepository interface {
	ListSlots(ctx context.Context, doctorID int64, status domain.SlotStatus, from, to time.Time) ([]domain.Slot, error)
	GetAppointment(ctx context.Context, id int64) (*domain.Appointment, error)
	ListByPatient(ctx context.Context, patientID int64, status *domain.Status) ([]domain.Appointment, error)
	ListDiagnoses(ctx context.Context, appointmentIDs []int64) ([]domain.Diagnosis, error)
	ListPrescriptions(ctx context.Context, appointmentIDs []int64) ([]domain.Prescription, error)

	InTx(ctx context.Context, fn func(ctx context.Context) error) error
	LockDoctorSchedule(ctx context.Context, doctorID int64) error
	ListActiveSlots(ctx context.Context, doctorID int64, from, to time.Time) ([]domain.Slot, error)
	InsertSlots(ctx context.Context, slots []domain.Slot) ([]domain.Slot, error)
	GetSlotsForUpdate(ctx context.Context, ids []int64) ([]domain.Slot, error)
	UpdateSlot(ctx context.Context, slot domain.Slot) error

	GetAppointmentForUpdate(ctx context.Context, id int64) (*domain.Appointment, error)
	HasQueued(ctx context.Context, patientID, doctorID int64) (bool, error)
	InsertAppointment(ctx context.Context, a domain.Appointment) (*domain.Appointment, error)
	UpdateAppointment(ctx context.Context, a domain.Appointment) error
	InsertDiagnosis(ctx context.Context, d domain.Diagnosis) (*domain.Diagnosis, error)
	InsertPrescription(ctx context.Context, p domain.Prescription) (*domain.Prescription, error)
}

// CatalogClient — то, что Appointment спрашивает у Catalog (CatalogGrpc).
type CatalogClient interface {
	GetDoctor(ctx context.Context, doctorID int64) (*catalog.DoctorReply, error)
	GetPatient(ctx context.Context, patientID int64) (*catalog.PatientReply, error)
	GetMedicalService(ctx context.Context, serviceID int64) (*catalog.ServiceReply, error)
	GetMedication(ctx context.Context, medicationID int64) (*catalog.MedicationReply, error)
}

type AppointmentService struct {
	repo     AppointmentRepository
	catalog  CatalogClient
	location *time.Location
	now      func() time.Time
}

// NewAppointmentService: location — часовой пояс медцентра, в нём считаются
// границы календарного дня для freeSlots.
func NewAppointmentService(repo AppointmentRepository, catalog CatalogClient, location *time.Location) *AppointmentService {
	return &AppointmentService{repo: repo, catalog: catalog, location: location, now: time.Now}
}

// FreeSlots возвращает свободные слоты врача на дату вида 2006-01-02.
func (s *AppointmentService) FreeSlots(ctx context.Context, doctorID int64, date string) ([]domain.Slot, error) {
	if err := pkgerrors.ValidateID("doctorId", doctorID); err != nil {
		return nil, err
	}

	day, err := time.ParseInLocation(dateLayout, date, s.location)
	if err != nil {
		return nil, fmt.Errorf("%w: date должна быть в формате YYYY-MM-DD, получено %q", pkgerrors.ErrInvalidArgument, date)
	}

	if _, err := s.catalog.GetDoctor(ctx, doctorID); err != nil {
		return nil, fmt.Errorf("check doctor %d: %w", doctorID, err)
	}

	return s.repo.ListSlots(ctx, doctorID, domain.SlotFree, day, day.AddDate(0, 0, 1))
}

// GetAppointment возвращает приём вместе с диагнозами и назначениями.
func (s *AppointmentService) GetAppointment(ctx context.Context, id int64) (*domain.Appointment, error) {
	if err := pkgerrors.ValidateID("appointmentId", id); err != nil {
		return nil, err
	}

	appointment, err := s.repo.GetAppointment(ctx, id)
	if err != nil {
		return nil, err
	}

	list := []domain.Appointment{*appointment}
	if err := s.attachDetails(ctx, list); err != nil {
		return nil, err
	}

	return &list[0], nil
}

// ListPatientAppointments возвращает приёмы пациента с деталями; status == nil — все.
func (s *AppointmentService) ListPatientAppointments(
	ctx context.Context,
	patientID int64,
	status *domain.Status,
) ([]domain.Appointment, error) {
	if err := pkgerrors.ValidateID("patientId", patientID); err != nil {
		return nil, err
	}
	if status != nil && !status.Valid() {
		return nil, fmt.Errorf("%w: неизвестный статус приёма %q", pkgerrors.ErrInvalidArgument, *status)
	}

	appointments, err := s.repo.ListByPatient(ctx, patientID, status)
	if err != nil {
		return nil, err
	}
	if err := s.attachDetails(ctx, appointments); err != nil {
		return nil, err
	}

	return appointments, nil
}

// attachDetails подтягивает диагнозы и назначения двумя запросами на весь список.
func (s *AppointmentService) attachDetails(ctx context.Context, appointments []domain.Appointment) error {
	if len(appointments) == 0 {
		return nil
	}

	ids := make([]int64, 0, len(appointments))
	index := make(map[int64]int, len(appointments))
	for i := range appointments {
		ids = append(ids, appointments[i].ID)
		index[appointments[i].ID] = i
		appointments[i].Diagnoses = []domain.Diagnosis{}
		appointments[i].Prescriptions = []domain.Prescription{}
	}

	diagnoses, err := s.repo.ListDiagnoses(ctx, ids)
	if err != nil {
		return err
	}
	for _, d := range diagnoses {
		if i, ok := index[d.AppointmentID]; ok {
			appointments[i].Diagnoses = append(appointments[i].Diagnoses, d)
		}
	}

	prescriptions, err := s.repo.ListPrescriptions(ctx, ids)
	if err != nil {
		return err
	}
	for _, p := range prescriptions {
		if i, ok := index[p.AppointmentID]; ok {
			appointments[i].Prescriptions = append(appointments[i].Prescriptions, p)
		}
	}

	return nil
}
