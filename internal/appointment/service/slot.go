package service

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/PaPaSmUrFiK/OuchFix/internal/appointment/domain"
	pkgerrors "github.com/PaPaSmUrFiK/OuchFix/pkg/errors"
)

// CreateSlotsInput — интервал приёма врача, который режется на слоты по 30 минут.
type CreateSlotsInput struct {
	DoctorID int64
	StartsAt string
	EndsAt   string
}

// UpdateSlotsInput — перенос свободных слотов. Для одного слота можно задать
// новое начало и/или конец. Для нескольких — только новое начало: все слоты
// сдвигаются на одну величину так, что самый ранний начинается в StartsAt.
type UpdateSlotsInput struct {
	SlotIDs  []int64
	StartsAt *string
	EndsAt   *string
}

// CreateSlots открывает запись к врачу на интервал [StartsAt, EndsAt).
func (s *AppointmentService) CreateSlots(ctx context.Context, in CreateSlotsInput) ([]domain.Slot, error) {
	if err := pkgerrors.ValidateID("doctorId", in.DoctorID); err != nil {
		return nil, err
	}
	from, err := parseTime("startsAt", in.StartsAt)
	if err != nil {
		return nil, err
	}
	to, err := parseTime("endsAt", in.EndsAt)
	if err != nil {
		return nil, err
	}
	if !from.Before(to) {
		return nil, fmt.Errorf("%w: startsAt должен быть раньше endsAt", pkgerrors.ErrInvalidArgument)
	}
	if to.Sub(from) > domain.MaxSlotsPeriod {
		return nil, fmt.Errorf("%w: интервал больше %s", pkgerrors.ErrInvalidArgument, domain.MaxSlotsPeriod)
	}
	if from.Before(s.now()) {
		return nil, fmt.Errorf("%w: нельзя открыть запись на прошедшее время", pkgerrors.ErrFailedPrecondition)
	}

	slots := cutSlots(in.DoctorID, from, to)
	if len(slots) == 0 {
		return nil, fmt.Errorf("%w: интервал короче слота %s", pkgerrors.ErrInvalidArgument, domain.SlotDuration)
	}

	if _, err := s.catalog.GetDoctor(ctx, in.DoctorID); err != nil {
		return nil, fmt.Errorf("check doctor %d: %w", in.DoctorID, err)
	}

	var created []domain.Slot
	err = s.repo.InTx(ctx, func(ctx context.Context) error {
		if err := s.repo.LockDoctorSchedule(ctx, in.DoctorID); err != nil {
			return err
		}
		if err := s.ensureNoOverlap(ctx, in.DoctorID, slots, nil); err != nil {
			return err
		}
		created, err = s.repo.InsertSlots(ctx, slots)
		return err
	})
	if err != nil {
		return nil, err
	}
	return created, nil
}

// UpdateSlots переносит свободные слоты на другое время.
func (s *AppointmentService) UpdateSlots(ctx context.Context, in UpdateSlotsInput) ([]domain.Slot, error) {
	ids, err := uniqueIDs("slotIds", in.SlotIDs)
	if err != nil {
		return nil, err
	}
	if in.StartsAt == nil && in.EndsAt == nil {
		return nil, fmt.Errorf("%w: нужно задать startsAt или endsAt", pkgerrors.ErrInvalidArgument)
	}
	if len(ids) > 1 && in.EndsAt != nil {
		return nil, fmt.Errorf("%w: endsAt задаётся только для одного слота", pkgerrors.ErrInvalidArgument)
	}
	var newStart, newEnd *time.Time
	if in.StartsAt != nil {
		t, err := parseTime("startsAt", *in.StartsAt)
		if err != nil {
			return nil, err
		}
		newStart = &t
	}
	if in.EndsAt != nil {
		t, err := parseTime("endsAt", *in.EndsAt)
		if err != nil {
			return nil, err
		}
		newEnd = &t
	}

	var updated []domain.Slot
	err = s.repo.InTx(ctx, func(ctx context.Context) error {
		slots, err := s.lockSlots(ctx, ids)
		if err != nil {
			return err
		}
		doctorID := slots[0].DoctorID
		for _, slot := range slots {
			if slot.DoctorID != doctorID {
				return fmt.Errorf("%w: слоты разных врачей переносятся отдельно", pkgerrors.ErrInvalidArgument)
			}
			if slot.Status != domain.SlotFree {
				return fmt.Errorf("%w: слот %d в статусе %s, переносить можно только свободный",
					pkgerrors.ErrFailedPrecondition, slot.ID, slot.Status)
			}
		}

		moved := moveSlots(slots, newStart, newEnd)
		for _, slot := range moved {
			if !slot.StartsAt.Before(slot.EndsAt) {
				return fmt.Errorf("%w: начало слота должно быть раньше конца", pkgerrors.ErrInvalidArgument)
			}
			if slot.StartsAt.Before(s.now()) {
				return fmt.Errorf("%w: нельзя перенести слот на прошедшее время", pkgerrors.ErrFailedPrecondition)
			}
		}

		if err := s.repo.LockDoctorSchedule(ctx, doctorID); err != nil {
			return err
		}
		if err := s.ensureNoOverlap(ctx, doctorID, moved, ids); err != nil {
			return err
		}
		for _, slot := range moved {
			if err := s.repo.UpdateSlot(ctx, slot); err != nil {
				return err
			}
		}
		updated = moved
		return nil
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}

// CancelSlots закрывает свободные слоты. Слот с записью отменить нельзя, пока
// не отменён приём; уже отменённый слот возвращается как есть.
func (s *AppointmentService) CancelSlots(ctx context.Context, slotIDs []int64) ([]domain.Slot, error) {
	ids, err := uniqueIDs("slotIds", slotIDs)
	if err != nil {
		return nil, err
	}

	var out []domain.Slot
	err = s.repo.InTx(ctx, func(ctx context.Context) error {
		slots, err := s.lockSlots(ctx, ids)
		if err != nil {
			return err
		}
		for _, slot := range slots {
			if slot.Status == domain.SlotBusy {
				return fmt.Errorf("%w: на слот %d есть запись, сначала отмените приём",
					pkgerrors.ErrFailedPrecondition, slot.ID)
			}
		}
		for i := range slots {
			if slots[i].Status == domain.SlotCancelled {
				continue
			}
			slots[i].Status = domain.SlotCancelled
			if err := s.repo.UpdateSlot(ctx, slots[i]); err != nil {
				return err
			}
		}
		out = slots
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// lockSlots блокирует слоты; если хоть одного нет — NOT_FOUND.
func (s *AppointmentService) lockSlots(ctx context.Context, ids []int64) ([]domain.Slot, error) {
	slots, err := s.repo.GetSlotsForUpdate(ctx, ids)
	if err != nil {
		return nil, err
	}
	if len(slots) != len(ids) {
		for _, id := range ids {
			if !slices.ContainsFunc(slots, func(sl domain.Slot) bool { return sl.ID == id }) {
				return nil, fmt.Errorf("slot %d: %w", id, pkgerrors.ErrNotFound)
			}
		}
	}
	return slots, nil
}

// ensureNoOverlap: новые слоты не пересекаются друг с другом (гарантирует нарезка
// или сдвиг на одну величину) и с действующими слотами врача, кроме exclude.
func (s *AppointmentService) ensureNoOverlap(ctx context.Context, doctorID int64, slots []domain.Slot, exclude []int64) error {
	from, to := slots[0].StartsAt, slots[0].EndsAt
	for _, sl := range slots[1:] {
		if sl.StartsAt.Before(from) {
			from = sl.StartsAt
		}
		if sl.EndsAt.After(to) {
			to = sl.EndsAt
		}
	}
	existing, err := s.repo.ListActiveSlots(ctx, doctorID, from, to)
	if err != nil {
		return err
	}
	for _, ex := range existing {
		if slices.Contains(exclude, ex.ID) {
			continue
		}
		for _, sl := range slots {
			if ex.Overlaps(sl.StartsAt, sl.EndsAt) {
				return fmt.Errorf("%w: время пересекается со слотом %d", pkgerrors.ErrConflict, ex.ID)
			}
		}
	}
	return nil
}

func cutSlots(doctorID int64, from, to time.Time) []domain.Slot {
	var slots []domain.Slot
	for start := from; !start.Add(domain.SlotDuration).After(to); start = start.Add(domain.SlotDuration) {
		slots = append(slots, domain.Slot{
			DoctorID: doctorID,
			StartsAt: start,
			EndsAt:   start.Add(domain.SlotDuration),
			Status:   domain.SlotFree,
		})
	}
	return slots
}

// moveSlots: один слот — задаются начало и/или конец; несколько — сдвиг всех
// на одну величину по новому началу самого раннего (slots отсортированы по id,
// поэтому ранний ищется явно).
func moveSlots(slots []domain.Slot, newStart, newEnd *time.Time) []domain.Slot {
	out := slices.Clone(slots)
	if len(out) == 1 {
		if newStart != nil {
			out[0].StartsAt = *newStart
		}
		if newEnd != nil {
			out[0].EndsAt = *newEnd
		}
		return out
	}
	earliest := out[0].StartsAt
	for _, sl := range out[1:] {
		if sl.StartsAt.Before(earliest) {
			earliest = sl.StartsAt
		}
	}
	shift := newStart.Sub(earliest)
	for i := range out {
		out[i].StartsAt = out[i].StartsAt.Add(shift)
		out[i].EndsAt = out[i].EndsAt.Add(shift)
	}
	return out
}

func parseTime(name, raw string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return time.Time{}, fmt.Errorf("%w: %s должен быть в формате RFC 3339 (2026-11-05T09:00:00+03:00), получено %q",
			pkgerrors.ErrInvalidArgument, name, raw)
	}
	return t, nil
}

// uniqueIDs проверяет id и убирает повторы, сохраняя порядок.
func uniqueIDs(name string, ids []int64) ([]int64, error) {
	if len(ids) == 0 {
		return nil, fmt.Errorf("%w: %s не должен быть пустым", pkgerrors.ErrInvalidArgument, name)
	}
	out := make([]int64, 0, len(ids))
	for _, id := range ids {
		if err := pkgerrors.ValidateID(name, id); err != nil {
			return nil, err
		}
		if !slices.Contains(out, id) {
			out = append(out, id)
		}
	}
	return out, nil
}
