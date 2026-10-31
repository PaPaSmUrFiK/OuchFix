package catalog

import (
	"context"
	"fmt"

	pkgerrors "github.com/PaPaSmUrFiK/OuchFix/pkg/errors"
)

// Stub — фейковый CatalogGrpc для ЛР3: фиксированные данные, совпадающие с seed
// миграций, без сети. Неизвестный id даёт ErrNotFound, как будет у gRPC NotFound.
type Stub struct {
	patients    map[int64]PatientReply
	doctors     map[int64]DoctorReply
	services    map[int64]ServiceReply
	medications map[int64]MedicationReply
}

var _ Client = (*Stub)(nil)

func NewStub() *Stub {
	return &Stub{
		patients: map[int64]PatientReply{
			1: {ID: 1, FullName: "Козлов Артём Андреевич", UserID: 5, Active: true},
			2: {ID: 2, FullName: "Смирнова Мария Олеговна", UserID: 6, Active: true},
			3: {ID: 3, FullName: "Новик Денис Павлович", UserID: 7, Active: true},
		},
		doctors: map[int64]DoctorReply{
			1: {ID: 1, FullName: "Иванов Пётр Сергеевич", SpecialtyID: 1, Office: "101", Active: true},
			2: {ID: 2, FullName: "Петрова Анна Игоревна", SpecialtyID: 2, Office: "205", Active: true},
			3: {ID: 3, FullName: "Сидоров Олег Викторович", SpecialtyID: 3, Active: true},
		},
		services: map[int64]ServiceReply{
			1: {ID: 1, Name: "Первичная консультация терапевта", PriceKopecks: 3500, DurationMin: 20},
			2: {ID: 2, Name: "Консультация кардиолога", PriceKopecks: 5500, DurationMin: 30},
			3: {ID: 3, Name: "ЭКГ", PriceKopecks: 2550, DurationMin: 15},
			4: {ID: 4, Name: "Консультация невролога", PriceKopecks: 5000},
		},
		medications: map[int64]MedicationReply{
			1: {ID: 1, Name: "Парацетамол 500 мг", Unit: "шт", StockQty: 200},
			2: {ID: 2, Name: "Бисопролол 5 мг", Unit: "шт", StockQty: 90},
			3: {ID: 3, Name: "Натрия хлорид 0,9%", Unit: "мл", StockQty: 1500.5},
		},
	}
}

func (s *Stub) GetPatient(ctx context.Context, patientID int64) (*PatientReply, error) {
	return lookup(ctx, s.patients, "patient", patientID)
}

func (s *Stub) GetDoctor(ctx context.Context, doctorID int64) (*DoctorReply, error) {
	return lookup(ctx, s.doctors, "doctor", doctorID)
}

func (s *Stub) GetMedicalService(ctx context.Context, serviceID int64) (*ServiceReply, error) {
	return lookup(ctx, s.services, "medical service", serviceID)
}

func (s *Stub) GetMedication(ctx context.Context, medicationID int64) (*MedicationReply, error) {
	return lookup(ctx, s.medications, "medication", medicationID)
}

func lookup[T any](ctx context.Context, items map[int64]T, kind string, id int64) (*T, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	item, ok := items[id]
	if !ok {
		return nil, fmt.Errorf("catalog stub: %s %d: %w", kind, id, pkgerrors.ErrNotFound)
	}
	return &item, nil
}
