package domain

import "time"

type SlotStatus string

const (
	SlotFree      SlotStatus = "free"
	SlotBusy      SlotStatus = "busy"
	SlotCancelled SlotStatus = "cancelled"
)

type Status string

const (
	StatusScheduled Status = "scheduled"
	StatusQueue     Status = "queue"
	StatusCompleted Status = "completed"
	StatusCancelled Status = "cancelled"
	StatusNoShow    Status = "no_show"
)

func (s Status) Valid() bool {
	switch s {
	case StatusScheduled, StatusQueue, StatusCompleted, StatusCancelled, StatusNoShow:
		return true
	}
	return false
}

type VisitType string

const (
	VisitScheduled VisitType = "scheduled"
	VisitWalkIn    VisitType = "walk_in"
)

type Slot struct {
	ID       int64
	DoctorID int64
	StartsAt time.Time
	EndsAt   time.Time
	Status   SlotStatus
}

type Appointment struct {
	ID            int64
	PatientID     int64
	DoctorID      int64
	SlotID        *int64
	ServiceID     *int64
	Status        Status
	VisitType     VisitType
	ScheduledAt   *time.Time
	Conclusion    *string
	CreatedAt     time.Time
	Diagnoses     []Diagnosis
	Prescriptions []Prescription
}

type Diagnosis struct {
	ID            int64
	AppointmentID int64
	Code          *string
	Description   string
	CreatedAt     time.Time
}

type Prescription struct {
	ID            int64
	AppointmentID int64
	Text          string
	MedicationID  *int64
	CreatedAt     time.Time
}

// SlotDuration — длительность слота расписания: интервал врача режется на такие слоты.
const SlotDuration = 30 * time.Minute

// MaxSlotsPeriod — наибольший интервал, который можно нарезать одной операцией (смена врача).
const MaxSlotsPeriod = 12 * time.Hour

// Overlaps — пересекаются ли интервалы [s.StartsAt, s.EndsAt) и [from, to).
func (s Slot) Overlaps(from, to time.Time) bool {
	return s.StartsAt.Before(to) && s.EndsAt.After(from)
}

// Active — приём ещё не завершён и не отменён: его можно отменить или провести.
func (a Appointment) Active() bool {
	return a.Status == StatusScheduled || a.Status == StatusQueue
}
