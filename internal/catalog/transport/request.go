package transport

import (
	"github.com/PaPaSmUrFiK/OuchFix/internal/catalog/domain"
	httpserver "github.com/PaPaSmUrFiK/OuchFix/pkg/transport/http/server"
)

// Тела запросов в форме аргументов GraphQL: id — строки.

type upsertMedicationRequest struct {
	Name     string  `json:"name"`
	Form     *string `json:"form"`
	Unit     string  `json:"unit"`
	StockQty float64 `json:"stockQty"`
}

func (r upsertMedicationRequest) toDomain() domain.Medication {
	return domain.Medication{Name: r.Name, Form: r.Form, Unit: r.Unit, StockQty: r.StockQty}
}

type dispenseRequest struct {
	MedicationID  string  `json:"medicationId"`
	PatientID     string  `json:"patientId"`
	AppointmentID *string `json:"appointmentId"`
	Quantity      float64 `json:"quantity"`
}

func (r dispenseRequest) toDomain() (domain.Dispense, error) {
	var (
		d   = domain.Dispense{Quantity: r.Quantity}
		err error
	)
	if d.MedicationID, err = httpserver.ParseID("medicationId", r.MedicationID); err != nil {
		return d, err
	}
	if d.PatientID, err = httpserver.ParseID("patientId", r.PatientID); err != nil {
		return d, err
	}
	d.AppointmentID, err = httpserver.ParseOptionalID("appointmentId", r.AppointmentID)
	return d, err
}
