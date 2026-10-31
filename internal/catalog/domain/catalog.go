package domain

import "time"

type Specialty struct {
	ID   int64
	Name string
}

type Patient struct {
	ID        int64
	UserID    int64
	FullName  string
	BirthDate *time.Time
	Gender    *string
	Phone     *string
	Address   *string
	CreatedAt time.Time
}

type Doctor struct {
	ID        int64
	UserID    int64
	FullName  string
	Office    *string
	Specialty Specialty
	CreatedAt time.Time
}

type MedicalService struct {
	ID          int64
	Name        string
	Price       float64
	DurationMin *int32
}

type Medication struct {
	ID       int64
	Name     string
	Form     *string
	Unit     string
	StockQty float64
}

// Dispense — выдача препарата пациенту со склада.
type Dispense struct {
	MedicationID  int64
	PatientID     int64
	AppointmentID *int64
	Quantity      float64
}
