package unit

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/PaPaSmUrFiK/OuchFix/internal/appointment/domain"
	"github.com/PaPaSmUrFiK/OuchFix/internal/appointment/service"
	"github.com/PaPaSmUrFiK/OuchFix/pkg/client/catalog"
	pkgerrors "github.com/PaPaSmUrFiK/OuchFix/pkg/errors"
)

// Даты в будущем относительно реального времени: сервис сверяет с time.Now.
const futureDay = "2030-03-04"

func at(t *testing.T, hhmm string) string {
	t.Helper()
	return futureDay + "T" + hhmm + ":00+03:00"
}

func ptr[T any](v T) *T { return &v }

func TestCreateSlotsCutsBy30Minutes(t *testing.T) {
	repo := &fakeAppointmentRepo{}
	svc := service.NewAppointmentService(repo, catalog.NewStub(), time.UTC)

	got, err := svc.CreateSlots(context.Background(), service.CreateSlotsInput{DoctorID: 1, StartsAt: at(t, "09:00"), EndsAt: at(t, "10:45")})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 09:00, 09:30, 10:00; хвост 10:30–10:45 короче слота и не создаётся.
	if len(got) != 3 || got[2].EndsAt.Sub(got[2].StartsAt) != domain.SlotDuration || got[0].Status != domain.SlotFree {
		t.Fatalf("bad slots: %+v", got)
	}
	if len(repo.locked) != 1 || repo.locked[0] != 1 {
		t.Fatalf("schedule must be locked, got %v", repo.locked)
	}
}

func TestCreateSlotsChecks(t *testing.T) {
	ctx := context.Background()
	for _, tc := range []struct {
		name string
		in   service.CreateSlotsInput
		want error
	}{
		{"bad doctor id", service.CreateSlotsInput{DoctorID: 0, StartsAt: at(t, "09:00"), EndsAt: at(t, "10:00")}, pkgerrors.ErrInvalidArgument},
		{"bad time", service.CreateSlotsInput{DoctorID: 1, StartsAt: "09:00", EndsAt: at(t, "10:00")}, pkgerrors.ErrInvalidArgument},
		{"reversed", service.CreateSlotsInput{DoctorID: 1, StartsAt: at(t, "10:00"), EndsAt: at(t, "09:00")}, pkgerrors.ErrInvalidArgument},
		{"shorter than slot", service.CreateSlotsInput{DoctorID: 1, StartsAt: at(t, "09:00"), EndsAt: at(t, "09:20")}, pkgerrors.ErrInvalidArgument},
		{"too long", service.CreateSlotsInput{DoctorID: 1, StartsAt: futureDay + "T00:00:00Z", EndsAt: "2030-03-05T00:00:00Z"}, pkgerrors.ErrInvalidArgument},
		{"past", service.CreateSlotsInput{DoctorID: 1, StartsAt: "2020-01-01T09:00:00Z", EndsAt: "2020-01-01T10:00:00Z"}, pkgerrors.ErrFailedPrecondition},
		{"unknown doctor", service.CreateSlotsInput{DoctorID: 99, StartsAt: at(t, "09:00"), EndsAt: at(t, "10:00")}, pkgerrors.ErrNotFound},
	} {
		repo := &fakeAppointmentRepo{}
		svc := service.NewAppointmentService(repo, catalog.NewStub(), time.UTC)
		if _, err := svc.CreateSlots(ctx, tc.in); !errors.Is(err, tc.want) {
			t.Fatalf("%s: want %v, got %v", tc.name, tc.want, err)
		}
		if len(repo.slots) != 0 {
			t.Fatalf("%s: nothing must be saved", tc.name)
		}
	}
}

func TestCreateSlotsOverlapIsConflict(t *testing.T) {
	repo := &fakeAppointmentRepo{}
	svc := service.NewAppointmentService(repo, catalog.NewStub(), time.UTC)
	ctx := context.Background()
	if _, err := svc.CreateSlots(ctx, service.CreateSlotsInput{DoctorID: 1, StartsAt: at(t, "09:00"), EndsAt: at(t, "10:00")}); err != nil {
		t.Fatal(err)
	}
	_, err := svc.CreateSlots(ctx, service.CreateSlotsInput{DoctorID: 1, StartsAt: at(t, "09:30"), EndsAt: at(t, "10:30")})
	if !errors.Is(err, pkgerrors.ErrConflict) {
		t.Fatalf("want ErrConflict, got %v", err)
	}
	// Другому врачу то же время можно.
	if _, err := svc.CreateSlots(ctx, service.CreateSlotsInput{DoctorID: 2, StartsAt: at(t, "09:30"), EndsAt: at(t, "10:30")}); err != nil {
		t.Fatalf("other doctor: %v", err)
	}
}

func newSlotFixture(t *testing.T) (*fakeAppointmentRepo, *service.AppointmentService) {
	t.Helper()
	repo := &fakeAppointmentRepo{}
	svc := service.NewAppointmentService(repo, catalog.NewStub(), time.UTC)
	if _, err := svc.CreateSlots(context.Background(), service.CreateSlotsInput{DoctorID: 1, StartsAt: at(t, "09:00"), EndsAt: at(t, "10:00")}); err != nil {
		t.Fatal(err)
	}
	return repo, svc
}

func TestUpdateSlotsShiftsFreeSlots(t *testing.T) {
	repo, svc := newSlotFixture(t)
	got, err := svc.UpdateSlots(context.Background(), service.UpdateSlotsInput{SlotIDs: []int64{1, 2}, StartsAt: ptr(at(t, "15:00"))})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 || repo.slots[2].StartsAt.UTC().Format("15:04") != "12:30" { // 15:30+03:00
		t.Fatalf("bad shift: %+v", repo.slots)
	}
	// Сдвиг одного слота на соседний внутри выборки не считается пересечением с самим собой.
	if _, err := svc.UpdateSlots(context.Background(), service.UpdateSlotsInput{SlotIDs: []int64{1}, EndsAt: ptr(at(t, "15:20"))}); err != nil {
		t.Fatalf("shrink: %v", err)
	}
}

func TestUpdateSlotsChecks(t *testing.T) {
	ctx := context.Background()
	repo, svc := newSlotFixture(t)
	busy := repo.slots[2]
	busy.Status = domain.SlotBusy
	repo.slots[2] = busy

	for _, tc := range []struct {
		name string
		in   service.UpdateSlotsInput
		want error
	}{
		{"empty", service.UpdateSlotsInput{StartsAt: ptr(at(t, "15:00"))}, pkgerrors.ErrInvalidArgument},
		{"no times", service.UpdateSlotsInput{SlotIDs: []int64{1}}, pkgerrors.ErrInvalidArgument},
		{"endsAt for many", service.UpdateSlotsInput{SlotIDs: []int64{1, 2}, EndsAt: ptr(at(t, "15:00"))}, pkgerrors.ErrInvalidArgument},
		{"missing", service.UpdateSlotsInput{SlotIDs: []int64{1, 42}, StartsAt: ptr(at(t, "15:00"))}, pkgerrors.ErrNotFound},
		{"busy", service.UpdateSlotsInput{SlotIDs: []int64{2}, StartsAt: ptr(at(t, "15:00"))}, pkgerrors.ErrFailedPrecondition},
		{"past", service.UpdateSlotsInput{SlotIDs: []int64{1}, StartsAt: ptr("2020-01-01T09:00:00Z"), EndsAt: ptr("2020-01-01T09:30:00Z")}, pkgerrors.ErrFailedPrecondition},
		{"reversed", service.UpdateSlotsInput{SlotIDs: []int64{1}, EndsAt: ptr(at(t, "08:00"))}, pkgerrors.ErrInvalidArgument},
		{"overlap", service.UpdateSlotsInput{SlotIDs: []int64{1}, StartsAt: ptr(at(t, "09:45")), EndsAt: ptr(at(t, "10:15"))}, pkgerrors.ErrConflict},
	} {
		if _, err := svc.UpdateSlots(ctx, tc.in); !errors.Is(err, tc.want) {
			t.Fatalf("%s: want %v, got %v", tc.name, tc.want, err)
		}
	}
	if repo.slots[1].StartsAt.UTC().Format(time.RFC3339) != "2030-03-04T06:00:00Z" {
		t.Fatalf("failed updates must not change slot: %+v", repo.slots[1])
	}
}

func TestCancelSlots(t *testing.T) {
	ctx := context.Background()
	repo, svc := newSlotFixture(t)

	got, err := svc.CancelSlots(ctx, []int64{1, 1})
	if err != nil || len(got) != 1 || repo.slots[1].Status != domain.SlotCancelled {
		t.Fatalf("cancel: %+v %v", got, err)
	}
	// Повторная отмена ничего не ломает.
	if _, err := svc.CancelSlots(ctx, []int64{1}); err != nil {
		t.Fatalf("repeat cancel: %v", err)
	}

	busy := repo.slots[2]
	busy.Status = domain.SlotBusy
	repo.slots[2] = busy
	if _, err := svc.CancelSlots(ctx, []int64{2}); !errors.Is(err, pkgerrors.ErrFailedPrecondition) {
		t.Fatalf("busy: want ErrFailedPrecondition, got %v", err)
	}
	if _, err := svc.CancelSlots(ctx, []int64{77}); !errors.Is(err, pkgerrors.ErrNotFound) {
		t.Fatalf("missing: want ErrNotFound, got %v", err)
	}
	if _, err := svc.CancelSlots(ctx, nil); !errors.Is(err, pkgerrors.ErrInvalidArgument) {
		t.Fatalf("empty: want ErrInvalidArgument, got %v", err)
	}
}
