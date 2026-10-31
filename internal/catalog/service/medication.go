package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"unicode/utf8"

	"github.com/PaPaSmUrFiK/OuchFix/internal/catalog/domain"
	pkgerrors "github.com/PaPaSmUrFiK/OuchFix/pkg/errors"
)

// maxQty — предел NUMERIC(12,2) в catalog.medication.
const maxQty = 1e10

// UpsertMedication — приход/корректировка склада: препарат ищется по названию.
func (s *CatalogService) UpsertMedication(ctx context.Context, m domain.Medication) (*domain.Medication, error) {
	m.Name = strings.TrimSpace(m.Name)
	m.Unit = strings.TrimSpace(m.Unit)
	if m.Form != nil {
		form := strings.TrimSpace(*m.Form)
		m.Form = &form
		if form == "" {
			m.Form = nil
		}
	}
	switch {
	case m.Name == "" || utf8.RuneCountInString(m.Name) > 255:
		return nil, fmt.Errorf("%w: name — от 1 до 255 символов", pkgerrors.ErrInvalidArgument)
	case m.Unit == "" || utf8.RuneCountInString(m.Unit) > 32:
		return nil, fmt.Errorf("%w: unit — от 1 до 32 символов", pkgerrors.ErrInvalidArgument)
	case m.Form != nil && utf8.RuneCountInString(*m.Form) > 128:
		return nil, fmt.Errorf("%w: form — не больше 128 символов", pkgerrors.ErrInvalidArgument)
	case !validQty(m.StockQty) && m.StockQty != 0:
		return nil, fmt.Errorf("%w: stockQty должно быть от 0 до 10^10", pkgerrors.ErrInvalidArgument)
	}
	return s.repo.UpsertMedication(ctx, m)
}

// DispenseMedication — выдача пациенту: остаток уменьшается атомарно, выдача
// записывается в журнал в той же транзакции.
func (s *CatalogService) DispenseMedication(ctx context.Context, d domain.Dispense) (*domain.Medication, error) {
	if err := pkgerrors.ValidateID("medicationId", d.MedicationID); err != nil {
		return nil, err
	}
	if err := pkgerrors.ValidateID("patientId", d.PatientID); err != nil {
		return nil, err
	}
	if d.AppointmentID != nil {
		if err := pkgerrors.ValidateID("appointmentId", *d.AppointmentID); err != nil {
			return nil, err
		}
	}
	if !validQty(d.Quantity) {
		return nil, fmt.Errorf("%w: quantity должно быть больше нуля", pkgerrors.ErrInvalidArgument)
	}
	if _, err := s.repo.GetPatient(ctx, d.PatientID); err != nil {
		return nil, fmt.Errorf("patient %d: %w", d.PatientID, err)
	}

	var out *domain.Medication
	err := s.repo.InTx(ctx, func(ctx context.Context) error {
		m, err := s.repo.TakeStock(ctx, d.MedicationID, d.Quantity)
		if errors.Is(err, pkgerrors.ErrNotFound) {
			current, getErr := s.repo.GetMedication(ctx, d.MedicationID)
			if getErr != nil {
				return fmt.Errorf("medication %d: %w", d.MedicationID, getErr)
			}
			return fmt.Errorf("%w: на складе %.2f %s, запрошено %.2f",
				pkgerrors.ErrFailedPrecondition, current.StockQty, current.Unit, d.Quantity)
		}
		if err != nil {
			return err
		}
		out = m
		return s.repo.InsertDispense(ctx, d)
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// validQty — положительное конечное количество в пределах колонки.
func validQty(q float64) bool {
	return q > 0 && q < maxQty && !math.IsNaN(q) && !math.IsInf(q, 0)
}
