package transport

import (
	"github.com/PaPaSmUrFiK/OuchFix/internal/appointment/service"
	httpserver "github.com/PaPaSmUrFiK/OuchFix/pkg/transport/http/server"
)

// Тела запросов в форме GraphQL input gateway: id — строки, gateway шлёт input как есть.

type createSlotsRequest struct {
	DoctorID string `json:"doctorId"`
	StartsAt string `json:"startsAt"`
	EndsAt   string `json:"endsAt"`
}

func (r createSlotsRequest) toInput() (service.CreateSlotsInput, error) {
	doctorID, err := httpserver.ParseID("doctorId", r.DoctorID)
	if err != nil {
		return service.CreateSlotsInput{}, err
	}
	return service.CreateSlotsInput{DoctorID: doctorID, StartsAt: r.StartsAt, EndsAt: r.EndsAt}, nil
}

type updateSlotsRequest struct {
	SlotIDs  []string `json:"slotIds"`
	StartsAt *string  `json:"startsAt"`
	EndsAt   *string  `json:"endsAt"`
}

func (r updateSlotsRequest) toInput() (service.UpdateSlotsInput, error) {
	ids, err := httpserver.ParseIDs("slotIds", r.SlotIDs)
	if err != nil {
		return service.UpdateSlotsInput{}, err
	}
	return service.UpdateSlotsInput{SlotIDs: ids, StartsAt: r.StartsAt, EndsAt: r.EndsAt}, nil
}

type cancelSlotsRequest struct {
	SlotIDs []string `json:"slotIds"`
}

type bookRequest struct {
	PatientID string  `json:"patientId"`
	DoctorID  string  `json:"doctorId"`
	SlotID    string  `json:"slotId"`
	ServiceID *string `json:"serviceId"`
}

func (r bookRequest) toInput() (service.BookInput, error) {
	var in service.BookInput
	var err error
	if in.PatientID, err = httpserver.ParseID("patientId", r.PatientID); err != nil {
		return in, err
	}
	if in.DoctorID, err = httpserver.ParseID("doctorId", r.DoctorID); err != nil {
		return in, err
	}
	if in.SlotID, err = httpserver.ParseID("slotId", r.SlotID); err != nil {
		return in, err
	}
	in.ServiceID, err = httpserver.ParseOptionalID("serviceId", r.ServiceID)
	return in, err
}

type walkInRequest struct {
	PatientID string  `json:"patientId"`
	DoctorID  string  `json:"doctorId"`
	ServiceID *string `json:"serviceId"`
}

func (r walkInRequest) toInput() (service.WalkInInput, error) {
	var in service.WalkInInput
	var err error
	if in.PatientID, err = httpserver.ParseID("patientId", r.PatientID); err != nil {
		return in, err
	}
	if in.DoctorID, err = httpserver.ParseID("doctorId", r.DoctorID); err != nil {
		return in, err
	}
	in.ServiceID, err = httpserver.ParseOptionalID("serviceId", r.ServiceID)
	return in, err
}

type rescheduleRequest struct {
	NewSlotID string `json:"newSlotId"`
}

type completeRequest struct {
	ServiceID  *string `json:"serviceId"`
	Conclusion *string `json:"conclusion"`
}
