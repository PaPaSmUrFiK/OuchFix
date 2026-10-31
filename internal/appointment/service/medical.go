package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/PaPaSmUrFiK/OuchFix/internal/appointment/domain"
	pkgerrors "github.com/PaPaSmUrFiK/OuchFix/pkg/errors"
)

type DiagnosisInput struct {
	AppointmentID int64
	Code          *string
	Description   string
}

type PrescriptionInput struct {
	AppointmentID int64
	Text          string
	MedicationID  *int64
}

// AddDiagnosis записывает диагноз в приём, который идёт или уже проведён.
func (s *AppointmentService) AddDiagnosis(ctx context.Context, in DiagnosisInput) (*domain.Diagnosis, error) {
	if err := pkgerrors.ValidateID("appointmentId", in.AppointmentID); err != nil {
		return nil, err
	}
	description := strings.TrimSpace(in.Description)
	if description == "" {
		return nil, fmt.Errorf("%w: description не должен быть пустым", pkgerrors.ErrInvalidArgument)
	}
	code := trimOptional(in.Code)

	var out *domain.Diagnosis
	err := s.repo.InTx(ctx, func(ctx context.Context) error {
		if err := s.lockMedicalRecord(ctx, in.AppointmentID); err != nil {
			return err
		}
		var err error
		out, err = s.repo.InsertDiagnosis(ctx, domain.Diagnosis{AppointmentID: in.AppointmentID, Code: code, Description: description})
		return err
	})
	return out, err
}

// AddPrescription записывает назначение; препарат, если указан, должен быть в справочнике.
func (s *AppointmentService) AddPrescription(ctx context.Context, in PrescriptionInput) (*domain.Prescription, error) {
	if err := pkgerrors.ValidateID("appointmentId", in.AppointmentID); err != nil {
		return nil, err
	}
	text := strings.TrimSpace(in.Text)
	if text == "" {
		return nil, fmt.Errorf("%w: text не должен быть пустым", pkgerrors.ErrInvalidArgument)
	}
	if in.MedicationID != nil {
		if err := pkgerrors.ValidateID("medicationId", *in.MedicationID); err != nil {
			return nil, err
		}
		if _, err := s.catalog.GetMedication(ctx, *in.MedicationID); err != nil {
			return nil, fmt.Errorf("check medication %d: %w", *in.MedicationID, err)
		}
	}

	var out *domain.Prescription
	err := s.repo.InTx(ctx, func(ctx context.Context) error {
		if err := s.lockMedicalRecord(ctx, in.AppointmentID); err != nil {
			return err
		}
		var err error
		out, err = s.repo.InsertPrescription(ctx, domain.Prescription{AppointmentID: in.AppointmentID, Text: text, MedicationID: in.MedicationID})
		return err
	})
	return out, err
}

// lockMedicalRecord: в отменённый приём и неявку медицинские записи не вносятся.
func (s *AppointmentService) lockMedicalRecord(ctx context.Context, id int64) error {
	a, err := s.repo.GetAppointmentForUpdate(ctx, id)
	if err != nil {
		return err
	}
	if a.Status == domain.StatusCancelled || a.Status == domain.StatusNoShow {
		return fmt.Errorf("%w: приём в статусе %s, медицинские записи не вносятся", pkgerrors.ErrFailedPrecondition, a.Status)
	}
	return nil
}

func trimOptional(v *string) *string {
	if v == nil {
		return nil
	}
	t := strings.TrimSpace(*v)
	if t == "" {
		return nil
	}
	return &t
}
