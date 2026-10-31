package unit

import (
	"context"
	"errors"
	"testing"
	"time"

	notificationdomain "github.com/PaPaSmUrFiK/OuchFix/internal/notification/domain"
	notificationsvc "github.com/PaPaSmUrFiK/OuchFix/internal/notification/service"
	paymentdomain "github.com/PaPaSmUrFiK/OuchFix/internal/payment/domain"
	paymentsvc "github.com/PaPaSmUrFiK/OuchFix/internal/payment/service"
	"github.com/PaPaSmUrFiK/OuchFix/pkg/client/appointment"
	"github.com/PaPaSmUrFiK/OuchFix/pkg/client/catalog"
	"github.com/PaPaSmUrFiK/OuchFix/pkg/client/iam"
	pkgerrors "github.com/PaPaSmUrFiK/OuchFix/pkg/errors"
)

type fakePaymentRepo struct {
	calls       int
	lastPatient *int64

	paid           map[int64]bool
	inserted       []paymentdomain.Payment
	sumFrom, sumTo time.Time
}

func (f *fakePaymentRepo) HasPaid(_ context.Context, appointmentID int64) (bool, error) {
	return f.paid[appointmentID], nil
}

func (f *fakePaymentRepo) Insert(_ context.Context, p paymentdomain.Payment) (*paymentdomain.Payment, error) {
	p.ID = int64(len(f.inserted) + 10)
	f.inserted = append(f.inserted, p)
	return &p, nil
}

func (f *fakePaymentRepo) Summary(_ context.Context, from, to time.Time) (*paymentdomain.Summary, error) {
	f.sumFrom, f.sumTo = from, to
	return &paymentdomain.Summary{Count: 1, Total: 35}, nil
}

func (f *fakePaymentRepo) List(_ context.Context, patientID *int64) ([]paymentdomain.Payment, error) {
	f.calls++
	f.lastPatient = patientID
	return []paymentdomain.Payment{}, nil
}

func TestPaymentsFilterAndValidation(t *testing.T) {
	repo := &fakePaymentRepo{}
	svc := paymentsvc.NewPaymentService(repo, catalog.NewStub(), appointment.NewStub(), time.UTC)

	if _, err := svc.ListPayments(context.Background(), nil); err != nil || repo.lastPatient != nil {
		t.Fatalf("nil patient must list all: err %v", err)
	}

	bad := int64(-5)
	if _, err := svc.ListPayments(context.Background(), &bad); !errors.Is(err, pkgerrors.ErrInvalidArgument) {
		t.Fatalf("want ErrInvalidArgument, got %v", err)
	}
	unknown := int64(404)
	if _, err := svc.ListPayments(context.Background(), &unknown); !errors.Is(err, pkgerrors.ErrNotFound) {
		t.Fatalf("want ErrNotFound for unknown patient, got %v", err)
	}
	if repo.calls != 1 {
		t.Fatalf("repository must not be called on invalid input, calls %d", repo.calls)
	}
}

type fakeNotificationRepo struct {
	calls int
}

func (f *fakeNotificationRepo) ListByUser(context.Context, int64, *notificationdomain.Status) ([]notificationdomain.Notification, error) {
	f.calls++
	return []notificationdomain.Notification{}, nil
}

func TestNotificationsValidation(t *testing.T) {
	repo := &fakeNotificationRepo{}
	svc := notificationsvc.NewNotificationService(repo, iam.NewStub())

	read := notificationdomain.StatusRead
	if _, err := svc.ListUserNotifications(context.Background(), 5, &read); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	bad := notificationdomain.Status("archived")
	if _, err := svc.ListUserNotifications(context.Background(), 5, &bad); !errors.Is(err, pkgerrors.ErrInvalidArgument) {
		t.Fatalf("want ErrInvalidArgument, got %v", err)
	}
	if _, err := svc.ListUserNotifications(context.Background(), 0, nil); !errors.Is(err, pkgerrors.ErrInvalidArgument) {
		t.Fatalf("want ErrInvalidArgument, got %v", err)
	}
	if _, err := svc.ListUserNotifications(context.Background(), 404, nil); !errors.Is(err, pkgerrors.ErrNotFound) {
		t.Fatalf("want ErrNotFound for unknown user, got %v", err)
	}
	if repo.calls != 1 {
		t.Fatalf("repository must not be called on invalid input, calls %d", repo.calls)
	}
}
