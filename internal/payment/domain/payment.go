package domain

import "time"

type Status string

const (
	StatusPending Status = "pending"
	StatusPaid    Status = "paid"
	StatusFailed  Status = "failed"
)

type Method string

const (
	MethodCash   Method = "cash"
	MethodCard   Method = "card"
	MethodOnline Method = "online"
)

type Payment struct {
	ID            int64
	AppointmentID int64
	PatientID     int64
	ServiceID     *int64
	Amount        float64
	Method        Method
	Status        Status
	CreatedAt     time.Time
	PaidAt        *time.Time
}
