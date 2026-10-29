package transport

import (
	"time"

	"github.com/PaPaSmUrFiK/OuchFix/internal/appointment/domain"
	httpserver "github.com/PaPaSmUrFiK/OuchFix/pkg/transport/http/server"
)

// Ответы в форме GraphQL-моделей gateway: gateway декодирует их без преобразований.

type slotResponse struct {
	ID       string `json:"id"`
	DoctorID string `json:"doctorId"`
	StartsAt string `json:"startsAt"`
	EndsAt   string `json:"endsAt"`
	Status   string `json:"status"`
}

type diagnosisResponse struct {
	ID          string  `json:"id"`
	Code        *string `json:"code"`
	Description string  `json:"description"`
}

type prescriptionResponse struct {
	ID           string  `json:"id"`
	Text         string  `json:"text"`
	MedicationID *string `json:"medicationId"`
}

type appointmentResponse struct {
	ID            string                 `json:"id"`
	PatientID     string                 `json:"patientId"`
	DoctorID      string                 `json:"doctorId"`
	SlotID        *string                `json:"slotId"`
	ServiceID     *string                `json:"serviceId"`
	Status        string                 `json:"status"`
	VisitType     string                 `json:"visitType"`
	ScheduledAt   *string                `json:"scheduledAt"`
	Conclusion    *string                `json:"conclusion"`
	Diagnoses     []diagnosisResponse    `json:"diagnoses"`
	Prescriptions []prescriptionResponse `json:"prescriptions"`
}

func toSlot(loc *time.Location) func(*domain.Slot) slotResponse {
	return func(s *domain.Slot) slotResponse {
		return slotResponse{
			ID:       httpserver.FormatID(s.ID),
			DoctorID: httpserver.FormatID(s.DoctorID),
			StartsAt: httpserver.FormatTime(s.StartsAt, loc),
			EndsAt:   httpserver.FormatTime(s.EndsAt, loc),
			Status:   httpserver.FormatEnum(s.Status),
		}
	}
}

func toAppointment(loc *time.Location) func(*domain.Appointment) appointmentResponse {
	return func(a *domain.Appointment) appointmentResponse {
		return appointmentResponse{
			ID:          httpserver.FormatID(a.ID),
			PatientID:   httpserver.FormatID(a.PatientID),
			DoctorID:    httpserver.FormatID(a.DoctorID),
			SlotID:      httpserver.FormatOptionalID(a.SlotID),
			ServiceID:   httpserver.FormatOptionalID(a.ServiceID),
			Status:      httpserver.FormatEnum(a.Status),
			VisitType:   httpserver.FormatEnum(a.VisitType),
			ScheduledAt: httpserver.FormatOptionalTime(a.ScheduledAt, loc),
			Conclusion:  a.Conclusion,
			Diagnoses: httpserver.MapList(a.Diagnoses, func(d *domain.Diagnosis) diagnosisResponse {
				return diagnosisResponse{ID: httpserver.FormatID(d.ID), Code: d.Code, Description: d.Description}
			}),
			Prescriptions: httpserver.MapList(a.Prescriptions, func(p *domain.Prescription) prescriptionResponse {
				return prescriptionResponse{ID: httpserver.FormatID(p.ID), Text: p.Text, MedicationID: httpserver.FormatOptionalID(p.MedicationID)}
			}),
		}
	}
}
