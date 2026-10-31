package unit

import (
	"context"
	"errors"
	"testing"

	"github.com/PaPaSmUrFiK/OuchFix/pkg/client/appointment"
	"github.com/PaPaSmUrFiK/OuchFix/pkg/client/catalog"
	"github.com/PaPaSmUrFiK/OuchFix/pkg/client/iam"
	pkgerrors "github.com/PaPaSmUrFiK/OuchFix/pkg/errors"
)

func TestStubsKnownAndUnknownIDs(t *testing.T) {
	ctx := context.Background()

	doctor, err := catalog.NewStub().GetDoctor(ctx, 2)
	if err != nil || doctor.SpecialtyID != 2 {
		t.Fatalf("doctor 2: %+v, %v", doctor, err)
	}
	if _, err := catalog.NewStub().GetMedication(ctx, 99); !errors.Is(err, pkgerrors.ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
	if _, err := appointment.NewStub().GetAppointment(ctx, 99); !errors.Is(err, pkgerrors.ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}

	allowed, err := iam.NewStub().CheckPermission(ctx, 5, "appointment:create")
	if err != nil || !allowed {
		t.Fatalf("patient must be allowed to book: %v, %v", allowed, err)
	}
	allowed, err = iam.NewStub().CheckPermission(ctx, 5, "schedule:manage")
	if err != nil || allowed {
		t.Fatalf("patient must not manage schedule: %v, %v", allowed, err)
	}
}

func TestStubsRespectCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := catalog.NewStub().GetPatient(ctx, 1); !errors.Is(err, context.Canceled) {
		t.Fatalf("want context.Canceled, got %v", err)
	}
}
