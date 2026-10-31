package service

import (
	"context"
	"fmt"

	"github.com/PaPaSmUrFiK/OuchFix/internal/appointment/domain"
	pkgerrors "github.com/PaPaSmUrFiK/OuchFix/pkg/errors"
)

type BookInput struct {
	PatientID int64
	DoctorID  int64
	SlotID    int64
	ServiceID *int64
}

type WalkInInput struct {
	PatientID int64
	DoctorID  int64
	ServiceID *int64
}

type CompleteInput struct {
	AppointmentID int64
	ServiceID     *int64
	Conclusion    *string
}

// Book записывает пациента на свободный слот врача: слот становится занятым,
// приём — запланированным.
func (s *AppointmentService) Book(ctx context.Context, in BookInput) (*domain.Appointment, error) {
	if err := pkgerrors.ValidateID("slotId", in.SlotID); err != nil {
		return nil, err
	}
	if err := s.checkParticipants(ctx, in.PatientID, in.DoctorID, in.ServiceID); err != nil {
		return nil, err
	}

	var id int64
	err := s.repo.InTx(ctx, func(ctx context.Context) error {
		slot, err := s.takeSlot(ctx, in.SlotID, in.DoctorID)
		if err != nil {
			return err
		}
		created, err := s.repo.InsertAppointment(ctx, domain.Appointment{
			PatientID:   in.PatientID,
			DoctorID:    in.DoctorID,
			SlotID:      &slot.ID,
			ServiceID:   in.ServiceID,
			Status:      domain.StatusScheduled,
			VisitType:   domain.VisitScheduled,
			ScheduledAt: &slot.StartsAt,
		})
		if err != nil {
			return err
		}
		id = created.ID
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.GetAppointment(ctx, id)
}

// WalkIn ставит пациента в живую очередь к врачу без записи на слот.
func (s *AppointmentService) WalkIn(ctx context.Context, in WalkInInput) (*domain.Appointment, error) {
	if err := s.checkParticipants(ctx, in.PatientID, in.DoctorID, in.ServiceID); err != nil {
		return nil, err
	}

	var id int64
	err := s.repo.InTx(ctx, func(ctx context.Context) error {
		// Очередь к врачу сериализуется той же блокировкой, что и его расписание.
		if err := s.repo.LockDoctorSchedule(ctx, in.DoctorID); err != nil {
			return err
		}
		queued, err := s.repo.HasQueued(ctx, in.PatientID, in.DoctorID)
		if err != nil {
			return err
		}
		if queued {
			return fmt.Errorf("%w: пациент %d уже в очереди к врачу %d", pkgerrors.ErrConflict, in.PatientID, in.DoctorID)
		}
		created, err := s.repo.InsertAppointment(ctx, domain.Appointment{
			PatientID: in.PatientID,
			DoctorID:  in.DoctorID,
			ServiceID: in.ServiceID,
			Status:    domain.StatusQueue,
			VisitType: domain.VisitWalkIn,
		})
		if err != nil {
			return err
		}
		id = created.ID
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.GetAppointment(ctx, id)
}

// Cancel отменяет запланированный приём или место в очереди и освобождает слот.
func (s *AppointmentService) Cancel(ctx context.Context, id int64) (*domain.Appointment, error) {
	if err := pkgerrors.ValidateID("appointmentId", id); err != nil {
		return nil, err
	}

	err := s.repo.InTx(ctx, func(ctx context.Context) error {
		a, err := s.lockActive(ctx, id, "отменить")
		if err != nil {
			return err
		}
		if a.SlotID != nil {
			if err := s.releaseSlot(ctx, *a.SlotID); err != nil {
				return err
			}
		}
		a.Status = domain.StatusCancelled
		return s.repo.UpdateAppointment(ctx, *a)
	})
	if err != nil {
		return nil, err
	}
	return s.GetAppointment(ctx, id)
}

// Reschedule переносит запланированный приём на другой свободный слот того же врача.
func (s *AppointmentService) Reschedule(ctx context.Context, id, newSlotID int64) (*domain.Appointment, error) {
	if err := pkgerrors.ValidateID("appointmentId", id); err != nil {
		return nil, err
	}
	if err := pkgerrors.ValidateID("newSlotId", newSlotID); err != nil {
		return nil, err
	}

	err := s.repo.InTx(ctx, func(ctx context.Context) error {
		a, err := s.repo.GetAppointmentForUpdate(ctx, id)
		if err != nil {
			return err
		}
		if a.Status != domain.StatusScheduled {
			return fmt.Errorf("%w: перенести можно только запланированный приём, текущий статус %s",
				pkgerrors.ErrFailedPrecondition, a.Status)
		}
		if a.SlotID != nil && *a.SlotID == newSlotID {
			return fmt.Errorf("%w: приём уже записан на слот %d", pkgerrors.ErrInvalidArgument, newSlotID)
		}
		slot, err := s.takeSlot(ctx, newSlotID, a.DoctorID)
		if err != nil {
			return err
		}
		if a.SlotID != nil {
			if err := s.releaseSlot(ctx, *a.SlotID); err != nil {
				return err
			}
		}
		a.SlotID = &slot.ID
		a.ScheduledAt = &slot.StartsAt
		return s.repo.UpdateAppointment(ctx, *a)
	})
	if err != nil {
		return nil, err
	}
	return s.GetAppointment(ctx, id)
}

// Complete проводит приём: фиксирует оказанную услугу и заключение врача.
func (s *AppointmentService) Complete(ctx context.Context, in CompleteInput) (*domain.Appointment, error) {
	if err := pkgerrors.ValidateID("appointmentId", in.AppointmentID); err != nil {
		return nil, err
	}
	if in.ServiceID != nil {
		if err := s.checkService(ctx, *in.ServiceID); err != nil {
			return nil, err
		}
	}

	err := s.repo.InTx(ctx, func(ctx context.Context) error {
		a, err := s.lockActive(ctx, in.AppointmentID, "провести")
		if err != nil {
			return err
		}
		if in.ServiceID != nil {
			a.ServiceID = in.ServiceID
		}
		if a.ServiceID == nil {
			return fmt.Errorf("%w: укажите оказанную услугу serviceId", pkgerrors.ErrInvalidArgument)
		}
		if in.Conclusion != nil {
			a.Conclusion = in.Conclusion
		}
		a.Status = domain.StatusCompleted
		return s.repo.UpdateAppointment(ctx, *a)
	})
	if err != nil {
		return nil, err
	}
	return s.GetAppointment(ctx, in.AppointmentID)
}

func (s *AppointmentService) checkParticipants(ctx context.Context, patientID, doctorID int64, serviceID *int64) error {
	if err := pkgerrors.ValidateID("patientId", patientID); err != nil {
		return err
	}
	if err := pkgerrors.ValidateID("doctorId", doctorID); err != nil {
		return err
	}
	if serviceID != nil {
		if err := s.checkService(ctx, *serviceID); err != nil {
			return err
		}
	}
	if _, err := s.catalog.GetPatient(ctx, patientID); err != nil {
		return fmt.Errorf("check patient %d: %w", patientID, err)
	}
	if _, err := s.catalog.GetDoctor(ctx, doctorID); err != nil {
		return fmt.Errorf("check doctor %d: %w", doctorID, err)
	}
	return nil
}

func (s *AppointmentService) checkService(ctx context.Context, serviceID int64) error {
	if err := pkgerrors.ValidateID("serviceId", serviceID); err != nil {
		return err
	}
	if _, err := s.catalog.GetMedicalService(ctx, serviceID); err != nil {
		return fmt.Errorf("check medical service %d: %w", serviceID, err)
	}
	return nil
}

// lockActive блокирует приём; завершённый, отменённый или неявка — FAILED_PRECONDITION.
func (s *AppointmentService) lockActive(ctx context.Context, id int64, action string) (*domain.Appointment, error) {
	a, err := s.repo.GetAppointmentForUpdate(ctx, id)
	if err != nil {
		return nil, err
	}
	if !a.Active() {
		return nil, fmt.Errorf("%w: нельзя %s приём в статусе %s", pkgerrors.ErrFailedPrecondition, action, a.Status)
	}
	return a, nil
}

// takeSlot занимает свободный будущий слот врача doctorID.
func (s *AppointmentService) takeSlot(ctx context.Context, slotID, doctorID int64) (*domain.Slot, error) {
	slots, err := s.lockSlots(ctx, []int64{slotID})
	if err != nil {
		return nil, err
	}
	slot := slots[0]
	switch {
	case slot.DoctorID != doctorID:
		return nil, fmt.Errorf("%w: слот %d принадлежит другому врачу", pkgerrors.ErrInvalidArgument, slotID)
	case slot.Status == domain.SlotBusy:
		return nil, fmt.Errorf("%w: слот %d уже занят", pkgerrors.ErrConflict, slotID)
	case slot.Status != domain.SlotFree:
		return nil, fmt.Errorf("%w: слот %d закрыт для записи", pkgerrors.ErrFailedPrecondition, slotID)
	case slot.StartsAt.Before(s.now()):
		return nil, fmt.Errorf("%w: слот %d уже прошёл", pkgerrors.ErrFailedPrecondition, slotID)
	}
	slot.Status = domain.SlotBusy
	if err := s.repo.UpdateSlot(ctx, slot); err != nil {
		return nil, err
	}
	return &slot, nil
}

// releaseSlot освобождает слот отменённого или перенесённого приёма. Слот,
// отменённый врачом, так и остаётся закрытым.
func (s *AppointmentService) releaseSlot(ctx context.Context, slotID int64) error {
	slots, err := s.lockSlots(ctx, []int64{slotID})
	if err != nil {
		return err
	}
	slot := slots[0]
	if slot.Status != domain.SlotBusy {
		return nil
	}
	slot.Status = domain.SlotFree
	return s.repo.UpdateSlot(ctx, slot)
}
