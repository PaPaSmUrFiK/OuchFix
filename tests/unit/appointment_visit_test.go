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

// Слоты 1 и 2 у врача 1 (09:00 и 09:30), слот 3 у врача 2, слот 4 закрыт, слот 5 в прошлом.
func newVisitFixture(t *testing.T) (*fakeAppointmentRepo, *service.AppointmentService) {
	t.Helper()
	future := time.Date(2030, 3, 4, 9, 0, 0, 0, time.UTC)
	repo := &fakeAppointmentRepo{slots: map[int64]domain.Slot{
		1: {ID: 1, DoctorID: 1, StartsAt: future, EndsAt: future.Add(30 * time.Minute), Status: domain.SlotFree},
		2: {ID: 2, DoctorID: 1, StartsAt: future.Add(30 * time.Minute), EndsAt: future.Add(time.Hour), Status: domain.SlotFree},
		3: {ID: 3, DoctorID: 2, StartsAt: future, EndsAt: future.Add(30 * time.Minute), Status: domain.SlotFree},
		4: {ID: 4, DoctorID: 1, StartsAt: future.Add(time.Hour), EndsAt: future.Add(90 * time.Minute), Status: domain.SlotCancelled},
		5: {ID: 5, DoctorID: 1, StartsAt: time.Date(2020, 1, 1, 9, 0, 0, 0, time.UTC), EndsAt: time.Date(2020, 1, 1, 9, 30, 0, 0, time.UTC), Status: domain.SlotFree},
	}}
	return repo, service.NewAppointmentService(repo, catalog.NewStub(), time.UTC)
}

func TestBookTakesSlot(t *testing.T) {
	repo, svc := newVisitFixture(t)
	got, err := svc.Book(context.Background(), service.BookInput{PatientID: 1, DoctorID: 1, SlotID: 1, ServiceID: ptr(int64(1))})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Status != domain.StatusScheduled || got.VisitType != domain.VisitScheduled || *got.SlotID != 1 || !got.ScheduledAt.Equal(repo.slots[1].StartsAt) {
		t.Fatalf("bad appointment: %+v", got)
	}
	if repo.slots[1].Status != domain.SlotBusy {
		t.Fatalf("slot must be busy")
	}
}

func TestBookChecks(t *testing.T) {
	ctx := context.Background()
	for _, tc := range []struct {
		name string
		in   service.BookInput
		want error
	}{
		{"bad slot id", service.BookInput{PatientID: 1, DoctorID: 1}, pkgerrors.ErrInvalidArgument},
		{"unknown patient", service.BookInput{PatientID: 99, DoctorID: 1, SlotID: 1}, pkgerrors.ErrNotFound},
		{"unknown service", service.BookInput{PatientID: 1, DoctorID: 1, SlotID: 1, ServiceID: ptr(int64(99))}, pkgerrors.ErrNotFound},
		{"unknown slot", service.BookInput{PatientID: 1, DoctorID: 1, SlotID: 42}, pkgerrors.ErrNotFound},
		{"other doctor", service.BookInput{PatientID: 1, DoctorID: 1, SlotID: 3}, pkgerrors.ErrInvalidArgument},
		{"closed slot", service.BookInput{PatientID: 1, DoctorID: 1, SlotID: 4}, pkgerrors.ErrFailedPrecondition},
		{"past slot", service.BookInput{PatientID: 1, DoctorID: 1, SlotID: 5}, pkgerrors.ErrFailedPrecondition},
	} {
		repo, svc := newVisitFixture(t)
		if _, err := svc.Book(ctx, tc.in); !errors.Is(err, tc.want) {
			t.Fatalf("%s: want %v, got %v", tc.name, tc.want, err)
		}
		if len(repo.appointments) != 0 {
			t.Fatalf("%s: appointment must not be created", tc.name)
		}
	}

	_, svc := newVisitFixture(t)
	if _, err := svc.Book(ctx, service.BookInput{PatientID: 1, DoctorID: 1, SlotID: 1}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Book(ctx, service.BookInput{PatientID: 2, DoctorID: 1, SlotID: 1}); !errors.Is(err, pkgerrors.ErrConflict) {
		t.Fatalf("double booking: want ErrConflict, got %v", err)
	}
}

func TestWalkInQueueOnce(t *testing.T) {
	ctx := context.Background()
	_, svc := newVisitFixture(t)
	got, err := svc.WalkIn(ctx, service.WalkInInput{PatientID: 2, DoctorID: 1})
	if err != nil || got.Status != domain.StatusQueue || got.VisitType != domain.VisitWalkIn || got.SlotID != nil {
		t.Fatalf("walk-in: %+v %v", got, err)
	}
	if _, err := svc.WalkIn(ctx, service.WalkInInput{PatientID: 2, DoctorID: 1}); !errors.Is(err, pkgerrors.ErrConflict) {
		t.Fatalf("second walk-in: want ErrConflict, got %v", err)
	}
}

func TestCancelFreesSlotAndRejectsFinished(t *testing.T) {
	ctx := context.Background()
	repo, svc := newVisitFixture(t)
	booked, err := svc.Book(ctx, service.BookInput{PatientID: 1, DoctorID: 1, SlotID: 1})
	if err != nil {
		t.Fatal(err)
	}

	got, err := svc.Cancel(ctx, booked.ID)
	if err != nil || got.Status != domain.StatusCancelled {
		t.Fatalf("cancel: %+v %v", got, err)
	}
	if repo.slots[1].Status != domain.SlotFree {
		t.Fatalf("slot must be free after cancel")
	}
	if _, err := svc.Cancel(ctx, booked.ID); !errors.Is(err, pkgerrors.ErrFailedPrecondition) {
		t.Fatalf("cancel twice: want ErrFailedPrecondition, got %v", err)
	}
	if _, err := svc.Cancel(ctx, 999); !errors.Is(err, pkgerrors.ErrNotFound) {
		t.Fatalf("missing: want ErrNotFound, got %v", err)
	}
}

func TestRescheduleMovesSlot(t *testing.T) {
	ctx := context.Background()
	repo, svc := newVisitFixture(t)
	booked, err := svc.Book(ctx, service.BookInput{PatientID: 1, DoctorID: 1, SlotID: 1})
	if err != nil {
		t.Fatal(err)
	}

	got, err := svc.Reschedule(ctx, booked.ID, 2)
	if err != nil || *got.SlotID != 2 || !got.ScheduledAt.Equal(repo.slots[2].StartsAt) {
		t.Fatalf("reschedule: %+v %v", got, err)
	}
	if repo.slots[1].Status != domain.SlotFree || repo.slots[2].Status != domain.SlotBusy {
		t.Fatalf("slots not swapped: %+v", repo.slots)
	}

	for _, tc := range []struct {
		name string
		slot int64
		want error
	}{
		{"same slot", 2, pkgerrors.ErrInvalidArgument},
		{"other doctor", 3, pkgerrors.ErrInvalidArgument},
		{"closed", 4, pkgerrors.ErrFailedPrecondition},
	} {
		if _, err := svc.Reschedule(ctx, booked.ID, tc.slot); !errors.Is(err, tc.want) {
			t.Fatalf("%s: want %v, got %v", tc.name, tc.want, err)
		}
	}

	walkIn, err := svc.WalkIn(ctx, service.WalkInInput{PatientID: 3, DoctorID: 1})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Reschedule(ctx, walkIn.ID, 1); !errors.Is(err, pkgerrors.ErrFailedPrecondition) {
		t.Fatalf("walk-in: want ErrFailedPrecondition, got %v", err)
	}
}

func TestCompleteRequiresServiceAndActiveStatus(t *testing.T) {
	ctx := context.Background()
	_, svc := newVisitFixture(t)
	walkIn, err := svc.WalkIn(ctx, service.WalkInInput{PatientID: 1, DoctorID: 1})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := svc.Complete(ctx, service.CompleteInput{AppointmentID: walkIn.ID}); !errors.Is(err, pkgerrors.ErrInvalidArgument) {
		t.Fatalf("no service: want ErrInvalidArgument, got %v", err)
	}
	if _, err := svc.Complete(ctx, service.CompleteInput{AppointmentID: walkIn.ID, ServiceID: ptr(int64(99))}); !errors.Is(err, pkgerrors.ErrNotFound) {
		t.Fatalf("unknown service: want ErrNotFound, got %v", err)
	}

	got, err := svc.Complete(ctx, service.CompleteInput{AppointmentID: walkIn.ID, ServiceID: ptr(int64(1)), Conclusion: ptr("Здоров")})
	if err != nil || got.Status != domain.StatusCompleted || *got.ServiceID != 1 || *got.Conclusion != "Здоров" {
		t.Fatalf("complete: %+v %v", got, err)
	}
	if _, err := svc.Complete(ctx, service.CompleteInput{AppointmentID: walkIn.ID, ServiceID: ptr(int64(1))}); !errors.Is(err, pkgerrors.ErrFailedPrecondition) {
		t.Fatalf("complete twice: want ErrFailedPrecondition, got %v", err)
	}
	if _, err := svc.Cancel(ctx, walkIn.ID); !errors.Is(err, pkgerrors.ErrFailedPrecondition) {
		t.Fatalf("cancel completed: want ErrFailedPrecondition, got %v", err)
	}
}

func TestMedicalRecords(t *testing.T) {
	ctx := context.Background()
	repo, svc := newVisitFixture(t)
	repo.appointments = map[int64]domain.Appointment{
		1: {ID: 1, PatientID: 1, DoctorID: 1, Status: domain.StatusQueue},
		2: {ID: 2, PatientID: 1, DoctorID: 1, Status: domain.StatusCancelled},
	}

	d, err := svc.AddDiagnosis(ctx, service.DiagnosisInput{AppointmentID: 1, Code: ptr("  "), Description: " ОРВИ "})
	if err != nil || d.Description != "ОРВИ" || d.Code != nil {
		t.Fatalf("diagnosis: %+v %v", d, err)
	}
	if _, err := svc.AddDiagnosis(ctx, service.DiagnosisInput{AppointmentID: 1, Description: " "}); !errors.Is(err, pkgerrors.ErrInvalidArgument) {
		t.Fatalf("empty description: %v", err)
	}
	if _, err := svc.AddDiagnosis(ctx, service.DiagnosisInput{AppointmentID: 2, Description: "x"}); !errors.Is(err, pkgerrors.ErrFailedPrecondition) {
		t.Fatalf("cancelled: %v", err)
	}
	if _, err := svc.AddPrescription(ctx, service.PrescriptionInput{AppointmentID: 1, Text: "x", MedicationID: ptr(int64(99))}); !errors.Is(err, pkgerrors.ErrNotFound) {
		t.Fatalf("unknown medication: %v", err)
	}
	p, err := svc.AddPrescription(ctx, service.PrescriptionInput{AppointmentID: 1, Text: "Парацетамол", MedicationID: ptr(int64(1))})
	if err != nil || *p.MedicationID != 1 {
		t.Fatalf("prescription: %+v %v", p, err)
	}
	if len(repo.diagnoses) != 1 || len(repo.prescriptions) != 1 {
		t.Fatalf("failed operations must not save records")
	}
}
