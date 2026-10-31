package service

import (
	"context"

	"github.com/PaPaSmUrFiK/OuchFix/internal/catalog/domain"
	pkgerrors "github.com/PaPaSmUrFiK/OuchFix/pkg/errors"
)

type CatalogRepository interface {
	ListSpecialties(ctx context.Context) ([]domain.Specialty, error)
	ListDoctors(ctx context.Context, specialtyID *int64) ([]domain.Doctor, error)
	GetDoctor(ctx context.Context, id int64) (*domain.Doctor, error)
	GetPatient(ctx context.Context, id int64) (*domain.Patient, error)
	GetPatientByUserID(ctx context.Context, userID int64) (*domain.Patient, error)
	ListMedicalServices(ctx context.Context) ([]domain.MedicalService, error)
	GetMedicalService(ctx context.Context, id int64) (*domain.MedicalService, error)
	ListMedications(ctx context.Context) ([]domain.Medication, error)
	GetMedication(ctx context.Context, id int64) (*domain.Medication, error)

	InTx(ctx context.Context, fn func(context.Context) error) error
	ReceiveMedication(ctx context.Context, m domain.Medication, qty float64) (*domain.Medication, error)
	TakeStock(ctx context.Context, medicationID int64, qty float64) (*domain.Medication, error)
	InsertDispense(ctx context.Context, d domain.Dispense) error
}

type CatalogService struct {
	repo CatalogRepository
}

func NewCatalogService(repo CatalogRepository) *CatalogService {
	return &CatalogService{repo: repo}
}

func (s *CatalogService) ListSpecialties(ctx context.Context) ([]domain.Specialty, error) {
	return s.repo.ListSpecialties(ctx)
}

func (s *CatalogService) ListDoctors(ctx context.Context, specialtyID *int64) ([]domain.Doctor, error) {
	if specialtyID != nil {
		if err := pkgerrors.ValidateID("specialtyId", *specialtyID); err != nil {
			return nil, err
		}
	}
	return s.repo.ListDoctors(ctx, specialtyID)
}

func (s *CatalogService) GetDoctor(ctx context.Context, id int64) (*domain.Doctor, error) {
	if err := pkgerrors.ValidateID("doctorId", id); err != nil {
		return nil, err
	}
	return s.repo.GetDoctor(ctx, id)
}

func (s *CatalogService) GetPatient(ctx context.Context, id int64) (*domain.Patient, error) {
	if err := pkgerrors.ValidateID("patientId", id); err != nil {
		return nil, err
	}
	return s.repo.GetPatient(ctx, id)
}

func (s *CatalogService) GetPatientByUserID(ctx context.Context, userID int64) (*domain.Patient, error) {
	if err := pkgerrors.ValidateID("userId", userID); err != nil {
		return nil, err
	}
	return s.repo.GetPatientByUserID(ctx, userID)
}

func (s *CatalogService) ListMedicalServices(ctx context.Context) ([]domain.MedicalService, error) {
	return s.repo.ListMedicalServices(ctx)
}

func (s *CatalogService) GetMedicalService(ctx context.Context, id int64) (*domain.MedicalService, error) {
	if err := pkgerrors.ValidateID("serviceId", id); err != nil {
		return nil, err
	}
	return s.repo.GetMedicalService(ctx, id)
}

func (s *CatalogService) ListMedications(ctx context.Context) ([]domain.Medication, error) {
	return s.repo.ListMedications(ctx)
}

func (s *CatalogService) GetMedication(ctx context.Context, id int64) (*domain.Medication, error) {
	if err := pkgerrors.ValidateID("medicationId", id); err != nil {
		return nil, err
	}
	return s.repo.GetMedication(ctx, id)
}
