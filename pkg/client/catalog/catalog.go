// Package catalog — клиент Catalog-сервиса по контракту catalog.proto (CatalogGrpc).
// В ЛР3 есть только заглушка Stub; в ЛР6 рядом появится реализация поверх gRPC.
package catalog

import "context"

type PatientReply struct {
	ID       int64
	FullName string
	UserID   int64
	Active   bool
}

type DoctorReply struct {
	ID          int64
	FullName    string
	SpecialtyID int64
	Office      string
	Active      bool
}

type ServiceReply struct {
	ID           int64
	Name         string
	PriceKopecks int64
	DurationMin  int32
}

type MedicationReply struct {
	ID       int64
	Name     string
	Unit     string
	StockQty float64
}

type Client interface {
	GetPatient(ctx context.Context, patientID int64) (*PatientReply, error)
	GetDoctor(ctx context.Context, doctorID int64) (*DoctorReply, error)
	GetMedicalService(ctx context.Context, serviceID int64) (*ServiceReply, error)
	GetMedication(ctx context.Context, medicationID int64) (*MedicationReply, error)
}
