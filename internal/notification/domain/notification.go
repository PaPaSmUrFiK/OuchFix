package domain

import "time"

type Status string

const (
	StatusPending Status = "pending"
	StatusSent    Status = "sent"
	StatusRead    Status = "read"
	StatusFailed  Status = "failed"
)

func (s Status) Valid() bool {
	switch s {
	case StatusPending, StatusSent, StatusRead, StatusFailed:
		return true
	}
	return false
}

type Notification struct {
	ID          int64
	UserID      int64
	Channel     string
	Title       string
	Body        string
	Status      Status
	RelatedType *string
	RelatedID   *int64
	CreatedAt   time.Time
	SentAt      *time.Time
}
