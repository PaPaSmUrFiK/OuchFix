package postgres

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/PaPaSmUrFiK/OuchFix/internal/catalog/domain"
	"github.com/PaPaSmUrFiK/OuchFix/internal/catalog/repository/postgres/models"
	"github.com/PaPaSmUrFiK/OuchFix/pkg/storage/postgres"
)

type CatalogRepository struct {
	db *sqlx.DB
}

func NewCatalogRepository(db *sqlx.DB) *CatalogRepository {
	return &CatalogRepository{db: db}
}

func (r *CatalogRepository) ListSpecialties(ctx context.Context) ([]domain.Specialty, error) {
	const query = `SELECT id, name FROM catalog.specialty ORDER BY name, id`

	rows := make([]models.Specialty, 0)
	if err := r.db.SelectContext(ctx, &rows, query); err != nil {
		return nil, fmt.Errorf("list specialties: %w", postgres.WrapDB(err))
	}

	return models.ToDomainList[models.Specialty, domain.Specialty](rows), nil
}

const doctorSelect = `
	SELECT d.id, d.user_id, d.full_name, d.office, d.created_at,
	       s.id AS specialty_id, s.name AS specialty_name
	FROM catalog.doctor d
	JOIN catalog.specialty s ON s.id = d.specialty_id`

// ListDoctors возвращает врачей; specialtyID == nil — без фильтра.
func (r *CatalogRepository) ListDoctors(ctx context.Context, specialtyID *int64) ([]domain.Doctor, error) {
	const query = doctorSelect + `
	WHERE $1::BIGINT IS NULL OR d.specialty_id = $1
	ORDER BY d.full_name, d.id`

	rows := make([]models.Doctor, 0)
	if err := r.db.SelectContext(ctx, &rows, query, specialtyID); err != nil {
		return nil, fmt.Errorf("list doctors: %w", postgres.WrapDB(err))
	}

	return models.ToDomainList[models.Doctor, domain.Doctor](rows), nil
}

func (r *CatalogRepository) GetDoctor(ctx context.Context, id int64) (*domain.Doctor, error) {
	const query = doctorSelect + ` WHERE d.id = $1`

	var row models.Doctor
	if err := r.db.GetContext(ctx, &row, query, id); err != nil {
		return nil, fmt.Errorf("get doctor %d: %w", id, postgres.WrapDB(err))
	}

	doctor := row.ToDomain()
	return &doctor, nil
}

const patientSelect = `
	SELECT id, user_id, full_name, birth_date, gender, phone, address, created_at
	FROM catalog.patient`

func (r *CatalogRepository) GetPatient(ctx context.Context, id int64) (*domain.Patient, error) {
	const query = patientSelect + ` WHERE id = $1`

	var row models.Patient
	if err := r.db.GetContext(ctx, &row, query, id); err != nil {
		return nil, fmt.Errorf("get patient %d: %w", id, postgres.WrapDB(err))
	}

	patient := row.ToDomain()
	return &patient, nil
}

func (r *CatalogRepository) GetPatientByUserID(ctx context.Context, userID int64) (*domain.Patient, error) {
	const query = patientSelect + ` WHERE user_id = $1`

	var row models.Patient
	if err := r.db.GetContext(ctx, &row, query, userID); err != nil {
		return nil, fmt.Errorf("get patient by user %d: %w", userID, postgres.WrapDB(err))
	}

	patient := row.ToDomain()
	return &patient, nil
}

const medicalServiceSelect = `SELECT id, name, price::FLOAT8 AS price, duration_min FROM catalog.medical_service`

func (r *CatalogRepository) ListMedicalServices(ctx context.Context) ([]domain.MedicalService, error) {
	const query = medicalServiceSelect + ` ORDER BY name, id`

	rows := make([]models.MedicalService, 0)
	if err := r.db.SelectContext(ctx, &rows, query); err != nil {
		return nil, fmt.Errorf("list medical services: %w", postgres.WrapDB(err))
	}

	return models.ToDomainList[models.MedicalService, domain.MedicalService](rows), nil
}

func (r *CatalogRepository) GetMedicalService(ctx context.Context, id int64) (*domain.MedicalService, error) {
	const query = medicalServiceSelect + ` WHERE id = $1`

	var row models.MedicalService
	if err := r.db.GetContext(ctx, &row, query, id); err != nil {
		return nil, fmt.Errorf("get medical service %d: %w", id, postgres.WrapDB(err))
	}

	service := row.ToDomain()
	return &service, nil
}

const medicationSelect = `SELECT id, name, form, unit, stock_qty::FLOAT8 AS stock_qty FROM catalog.medication`

func (r *CatalogRepository) ListMedications(ctx context.Context) ([]domain.Medication, error) {
	const query = medicationSelect + ` ORDER BY name, id`

	rows := make([]models.Medication, 0)
	if err := r.db.SelectContext(ctx, &rows, query); err != nil {
		return nil, fmt.Errorf("list medications: %w", postgres.WrapDB(err))
	}

	return models.ToDomainList[models.Medication, domain.Medication](rows), nil
}

func (r *CatalogRepository) GetMedication(ctx context.Context, id int64) (*domain.Medication, error) {
	const query = medicationSelect + ` WHERE id = $1`

	var row models.Medication
	if err := r.db.GetContext(ctx, &row, query, id); err != nil {
		return nil, fmt.Errorf("get medication %d: %w", id, postgres.WrapDB(err))
	}

	medication := row.ToDomain()
	return &medication, nil
}
