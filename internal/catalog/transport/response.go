package transport

import (
	"github.com/PaPaSmUrFiK/OuchFix/internal/catalog/domain"
	httpserver "github.com/PaPaSmUrFiK/OuchFix/pkg/transport/http/server"
)

// Ответы в форме GraphQL-моделей gateway: gateway декодирует их без преобразований.

type specialtyResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type doctorResponse struct {
	ID        string            `json:"id"`
	FullName  string            `json:"fullName"`
	Office    *string           `json:"office"`
	Specialty specialtyResponse `json:"specialty"`
}

// patientResponse — только то, что нужно gateway для myAppointments.
type patientResponse struct {
	ID string `json:"id"`
}

type medicalServiceResponse struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	PriceKopecks int64  `json:"priceKopecks"`
	DurationMin  *int32 `json:"durationMin"`
}

type medicationResponse struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Form     *string `json:"form"`
	Unit     string  `json:"unit"`
	StockQty float64 `json:"stockQty"`
}

func toSpecialty(s *domain.Specialty) specialtyResponse {
	return specialtyResponse{ID: httpserver.FormatID(s.ID), Name: s.Name}
}

func toDoctor(d *domain.Doctor) doctorResponse {
	return doctorResponse{
		ID:        httpserver.FormatID(d.ID),
		FullName:  d.FullName,
		Office:    d.Office,
		Specialty: toSpecialty(&d.Specialty),
	}
}

func toPatient(p *domain.Patient) patientResponse {
	return patientResponse{ID: httpserver.FormatID(p.ID)}
}

func toMedicalService(s *domain.MedicalService) medicalServiceResponse {
	return medicalServiceResponse{
		ID:           httpserver.FormatID(s.ID),
		Name:         s.Name,
		PriceKopecks: s.PriceKopecks,
		DurationMin:  s.DurationMin,
	}
}

func toMedication(m *domain.Medication) medicationResponse {
	return medicationResponse{
		ID:       httpserver.FormatID(m.ID),
		Name:     m.Name,
		Form:     m.Form,
		Unit:     m.Unit,
		StockQty: m.StockQty,
	}
}
