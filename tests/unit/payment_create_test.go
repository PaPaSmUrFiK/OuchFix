package unit

import (
	"context"
	"errors"
	"testing"
	"time"

	paymentdomain "github.com/PaPaSmUrFiK/OuchFix/internal/payment/domain"
	paymentsvc "github.com/PaPaSmUrFiK/OuchFix/internal/payment/service"
	"github.com/PaPaSmUrFiK/OuchFix/pkg/client/appointment"
	"github.com/PaPaSmUrFiK/OuchFix/pkg/client/catalog"
	pkgerrors "github.com/PaPaSmUrFiK/OuchFix/pkg/errors"
)

func TestCreatePayment(t *testing.T) {
	repo := &fakePaymentRepo{paid: map[int64]bool{}}
	svc := paymentsvc.NewPaymentService(repo, catalog.NewStub(), appointment.NewStub(), minsk(t))
	ctx := context.Background()
	// Приём 1 из заглушки: пациент 1, услуга 1 (35.00), проведён.
	ok := paymentsvc.CreateInput{AppointmentID: 1, PatientID: 1, Amount: 35, Method: paymentdomain.MethodCard}

	p, err := svc.CreatePayment(ctx, ok)
	if err != nil {
		t.Fatal(err)
	}
	if p.Status != paymentdomain.StatusPaid || p.PaidAt == nil || p.ServiceID == nil || *p.ServiceID != 1 || p.Amount != 35 {
		t.Fatalf("payment = %+v", p)
	}

	repo.paid[1] = true
	if _, err := svc.CreatePayment(ctx, ok); !errors.Is(err, pkgerrors.ErrConflict) {
		t.Fatalf("repeat: %v", err)
	}
	repo.paid[1] = false

	wrongSvc := int64(2)
	for name, tc := range map[string]struct {
		in   paymentsvc.CreateInput
		want error
	}{
		"not completed":  {paymentsvc.CreateInput{AppointmentID: 2, PatientID: 1, Amount: 55, Method: "cash"}, pkgerrors.ErrFailedPrecondition},
		"other patient":  {paymentsvc.CreateInput{AppointmentID: 1, PatientID: 2, Amount: 35, Method: "cash"}, pkgerrors.ErrInvalidArgument},
		"wrong amount":   {paymentsvc.CreateInput{AppointmentID: 1, PatientID: 1, Amount: 34.99, Method: "cash"}, pkgerrors.ErrInvalidArgument},
		"wrong service":  {paymentsvc.CreateInput{AppointmentID: 1, PatientID: 1, ServiceID: &wrongSvc, Amount: 35, Method: "cash"}, pkgerrors.ErrInvalidArgument},
		"zero amount":    {paymentsvc.CreateInput{AppointmentID: 1, PatientID: 1, Amount: 0, Method: "cash"}, pkgerrors.ErrInvalidArgument},
		"bad method":     {paymentsvc.CreateInput{AppointmentID: 1, PatientID: 1, Amount: 35, Method: "barter"}, pkgerrors.ErrInvalidArgument},
		"no appointment": {paymentsvc.CreateInput{AppointmentID: 99, PatientID: 1, Amount: 35, Method: "cash"}, pkgerrors.ErrNotFound},
	} {
		if _, err := svc.CreatePayment(ctx, tc.in); !errors.Is(err, tc.want) {
			t.Fatalf("%s: want %v, got %v", name, tc.want, err)
		}
	}
	if len(repo.inserted) != 1 {
		t.Fatalf("rejected payments must not be inserted: %d", len(repo.inserted))
	}
}

func TestPaymentSummaryPeriod(t *testing.T) {
	repo := &fakePaymentRepo{}
	svc := paymentsvc.NewPaymentService(repo, catalog.NewStub(), appointment.NewStub(), minsk(t))
	ctx := context.Background()

	if _, err := svc.PaymentSummary(ctx, "2026-11-01", "2026-11-30"); err != nil {
		t.Fatal(err)
	}
	if !repo.sumFrom.Equal(time.Date(2026, 11, 1, 0, 0, 0, 0, minsk(t))) || !repo.sumTo.Equal(time.Date(2026, 12, 1, 0, 0, 0, 0, minsk(t))) {
		t.Fatalf("period %s..%s", repo.sumFrom, repo.sumTo)
	}
	for _, p := range [][2]string{{"2026-11-02", "2026-11-01"}, {"x", "2026-11-01"}, {"2026-01-01", "2027-03-01"}} {
		if _, err := svc.PaymentSummary(ctx, p[0], p[1]); !errors.Is(err, pkgerrors.ErrInvalidArgument) {
			t.Fatalf("%v: %v", p, err)
		}
	}
}
