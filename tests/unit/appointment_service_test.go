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

type fakeAppointmentRepo struct {
	appointments  map[int64]domain.Appointment
	diagnoses     []domain.Diagnosis
	prescriptions []domain.Prescription

	slotFrom, slotTo time.Time
	slotStatus       domain.SlotStatus
	slotCalls        int
	lastStatus       *domain.Status
}

func (f *fakeAppointmentRepo) ListSlots(_ context.Context, _ int64, status domain.SlotStatus, from, to time.Time) ([]domain.Slot, error) {
	f.slotCalls++
	f.slotStatus, f.slotFrom, f.slotTo = status, from, to
	return []domain.Slot{}, nil
}

func (f *fakeAppointmentRepo) GetAppointment(_ context.Context, id int64) (*domain.Appointment, error) {
	a, ok := f.appointments[id]
	if !ok {
		return nil, pkgerrors.ErrNotFound
	}
	return &a, nil
}

func (f *fakeAppointmentRepo) ListByPatient(_ context.Context, patientID int64, status *domain.Status) ([]domain.Appointment, error) {
	f.lastStatus = status
	out := []domain.Appointment{}
	for _, a := range f.appointments {
		if a.PatientID == patientID {
			out = append(out, a)
		}
	}
	return out, nil
}

func (f *fakeAppointmentRepo) ListDiagnoses(context.Context, []int64) ([]domain.Diagnosis, error) {
	return f.diagnoses, nil
}

func (f *fakeAppointmentRepo) ListPrescriptions(context.Context, []int64) ([]domain.Prescription, error) {
	return f.prescriptions, nil
}

func minsk(t *testing.T) *time.Location {
	t.Helper()
	loc, err := time.LoadLocation("Europe/Minsk")
	if err != nil {
		t.Fatalf("load location: %v", err)
	}
	return loc
}

func TestFreeSlotsUsesLocalDayBounds(t *testing.T) {
	repo := &fakeAppointmentRepo{}
	svc := service.NewAppointmentService(repo, catalog.NewStub(), minsk(t))

	if _, err := svc.FreeSlots(context.Background(), 1, "2026-11-02"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wantFrom := time.Date(2026, 11, 2, 0, 0, 0, 0, minsk(t))
	if !repo.slotFrom.Equal(wantFrom) || !repo.slotTo.Equal(wantFrom.AddDate(0, 0, 1)) {
		t.Fatalf("bad bounds: %s..%s", repo.slotFrom, repo.slotTo)
	}
	if repo.slotStatus != domain.SlotFree {
		t.Fatalf("must request free slots, got %q", repo.slotStatus)
	}
}

func TestFreeSlotsRejectsBadInput(t *testing.T) {
	repo := &fakeAppointmentRepo{}
	svc := service.NewAppointmentService(repo, catalog.NewStub(), minsk(t))

	for _, date := range []string{"", "02.11.2026", "2026-13-01", "2026-11-02T10:00"} {
		if _, err := svc.FreeSlots(context.Background(), 1, date); !errors.Is(err, pkgerrors.ErrInvalidArgument) {
			t.Fatalf("date %q: want ErrInvalidArgument, got %v", date, err)
		}
	}
	if _, err := svc.FreeSlots(context.Background(), 0, "2026-11-02"); !errors.Is(err, pkgerrors.ErrInvalidArgument) {
		t.Fatalf("want ErrInvalidArgument for doctor 0, got %v", err)
	}
	if repo.slotCalls != 0 {
		t.Fatalf("repository must not be called on invalid input")
	}
}

func TestGetAppointmentAttachesDetails(t *testing.T) {
	repo := &fakeAppointmentRepo{
		appointments:  map[int64]domain.Appointment{1: {ID: 1, PatientID: 1}},
		diagnoses:     []domain.Diagnosis{{ID: 1, AppointmentID: 1}, {ID: 2, AppointmentID: 1}},
		prescriptions: []domain.Prescription{{ID: 1, AppointmentID: 1}},
	}
	svc := service.NewAppointmentService(repo, catalog.NewStub(), time.UTC)

	got, err := svc.GetAppointment(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got.Diagnoses) != 2 || len(got.Prescriptions) != 1 {
		t.Fatalf("details not attached: %+v", got)
	}

	if _, err := svc.GetAppointment(context.Background(), 99); !errors.Is(err, pkgerrors.ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}

func TestPatientAppointmentsEmptyDetailsAreNotNil(t *testing.T) {
	repo := &fakeAppointmentRepo{appointments: map[int64]domain.Appointment{2: {ID: 2, PatientID: 3}}}
	svc := service.NewAppointmentService(repo, catalog.NewStub(), time.UTC)

	status := domain.StatusQueue
	got, err := svc.ListPatientAppointments(context.Background(), 3, &status)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got[0].Diagnoses == nil || got[0].Prescriptions == nil {
		t.Fatalf("details must be empty slices: %+v", got)
	}
	if repo.lastStatus == nil || *repo.lastStatus != domain.StatusQueue {
		t.Fatalf("status filter not passed")
	}

	bad := domain.Status("unknown")
	if _, err := svc.ListPatientAppointments(context.Background(), 3, &bad); !errors.Is(err, pkgerrors.ErrInvalidArgument) {
		t.Fatalf("want ErrInvalidArgument, got %v", err)
	}
}

func TestFreeSlotsUnknownDoctorIsNotFound(t *testing.T) {
	repo := &fakeAppointmentRepo{}
	svc := service.NewAppointmentService(repo, catalog.NewStub(), time.UTC)

	if _, err := svc.FreeSlots(context.Background(), 999, "2026-11-02"); !errors.Is(err, pkgerrors.ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
	if repo.slotCalls != 0 {
		t.Fatalf("slots must not be read for unknown doctor")
	}
}
