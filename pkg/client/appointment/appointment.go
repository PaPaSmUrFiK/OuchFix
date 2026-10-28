// Package appointment — клиент Appointment-сервиса по контракту appointment.proto
// (AppointmentGrpc). В ЛР3 есть только заглушка Stub; gRPC — ЛР6.
package appointment

import "context"

type AppointmentReply struct {
	ID        int64
	PatientID int64
	DoctorID  int64
	ServiceID int64
	Status    string
}

type Client interface {
	GetAppointment(ctx context.Context, appointmentID int64) (*AppointmentReply, error)
}
