package appointment

import (
	"context"
	"fmt"

	pkgerrors "github.com/PaPaSmUrFiK/OuchFix/pkg/errors"
)

// Stub — фейковый AppointmentGrpc для ЛР3: приёмы как в seed миграций.
type Stub struct {
	appointments map[int64]AppointmentReply
}

var _ Client = (*Stub)(nil)

func NewStub() *Stub {
	return &Stub{
		appointments: map[int64]AppointmentReply{
			1: {ID: 1, PatientID: 1, DoctorID: 1, ServiceID: 1, Status: "completed"},
			2: {ID: 2, PatientID: 1, DoctorID: 2, ServiceID: 2, Status: "scheduled"},
			3: {ID: 3, PatientID: 2, DoctorID: 3, ServiceID: 4, Status: "scheduled"},
			4: {ID: 4, PatientID: 3, DoctorID: 1, ServiceID: 1, Status: "queue"},
			5: {ID: 5, PatientID: 2, DoctorID: 2, ServiceID: 3, Status: "cancelled"},
		},
	}
}

func (s *Stub) GetAppointment(ctx context.Context, appointmentID int64) (*AppointmentReply, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	item, ok := s.appointments[appointmentID]
	if !ok {
		return nil, fmt.Errorf("appointment stub: appointment %d: %w", appointmentID, pkgerrors.ErrNotFound)
	}
	return &item, nil
}
