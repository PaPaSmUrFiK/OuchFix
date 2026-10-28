// Package client — HTTP-клиенты gateway к внутренним REST API сервисов.
// Реализуют интерфейсы резолверов; в ЛР6 их заменят gRPC-клиенты.
package client

import (
	"context"
	"net/url"
	"strconv"

	"github.com/PaPaSmUrFiK/OuchFix/internal/catalog/domain"
	httpclient "github.com/PaPaSmUrFiK/OuchFix/pkg/transport/http/client"
)

type Catalog struct {
	http *httpclient.Client
}

func NewCatalog(http *httpclient.Client) *Catalog {
	return &Catalog{http: http}
}

func (c *Catalog) ListSpecialties(ctx context.Context) ([]domain.Specialty, error) {
	var out []domain.Specialty
	return out, c.http.Get(ctx, "/specialties", nil, &out)
}

func (c *Catalog) ListDoctors(ctx context.Context, specialtyID *int64) ([]domain.Doctor, error) {
	query := url.Values{}
	if specialtyID != nil {
		query.Set("specialtyId", strconv.FormatInt(*specialtyID, 10))
	}
	var out []domain.Doctor
	return out, c.http.Get(ctx, "/doctors", query, &out)
}

func (c *Catalog) GetPatientByUserID(ctx context.Context, userID int64) (*domain.Patient, error) {
	var out domain.Patient
	if err := c.http.Get(ctx, "/patients/by-user/"+strconv.FormatInt(userID, 10), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Catalog) ListMedicalServices(ctx context.Context) ([]domain.MedicalService, error) {
	var out []domain.MedicalService
	return out, c.http.Get(ctx, "/medical-services", nil, &out)
}

func (c *Catalog) ListMedications(ctx context.Context) ([]domain.Medication, error) {
	var out []domain.Medication
	return out, c.http.Get(ctx, "/medications", nil, &out)
}
