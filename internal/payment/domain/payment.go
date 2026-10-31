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

func (m Method) Valid() bool {
	return m == MethodCash || m == MethodCard || m == MethodOnline
}

// Filter — условия списка платежей; nil-поле не ограничивает. Период по
// created_at: [From, To).
type Filter struct {
	PatientID *int64
	Status    *Status
	From      *time.Time
	To        *time.Time
}

func (s Status) Valid() bool {
	return s == StatusPending || s == StatusPaid || s == StatusFailed
}

// Summary — выручка за период: число и сумма оплаченных платежей.
type Summary struct {
	Count int
	Total float64
}
