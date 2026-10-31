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
