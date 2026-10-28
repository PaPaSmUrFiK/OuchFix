package unit

import (
	"context"
	"errors"
	"testing"

	notificationdomain "github.com/PaPaSmUrFiK/OuchFix/internal/notification/domain"
	notificationsvc "github.com/PaPaSmUrFiK/OuchFix/internal/notification/service"
	paymentdomain "github.com/PaPaSmUrFiK/OuchFix/internal/payment/domain"
	paymentsvc "github.com/PaPaSmUrFiK/OuchFix/internal/payment/service"
	pkgerrors "github.com/PaPaSmUrFiK/OuchFix/pkg/errors"
)

type fakePaymentRepo struct {
	calls       int
	lastPatient *int64
}

func (f *fakePaymentRepo) List(_ context.Context, patientID *int64) ([]paymentdomain.Payment, error) {
	f.calls++
	f.lastPatient = patientID
	return []paymentdomain.Payment{}, nil
}

func TestPaymentsFilterAndValidation(t *testing.T) {
	repo := &fakePaymentRepo{}
	svc := paymentsvc.NewPaymentService(repo)

	if _, err := svc.ListPayments(context.Background(), nil); err != nil || repo.lastPatient != nil {
		t.Fatalf("nil patient must list all: err %v", err)
	}

	bad := int64(-5)
	if _, err := svc.ListPayments(context.Background(), &bad); !errors.Is(err, pkgerrors.ErrInvalidArgument) {
		t.Fatalf("want ErrInvalidArgument, got %v", err)
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
	svc := notificationsvc.NewNotificationService(repo)

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
	if repo.calls != 1 {
		t.Fatalf("repository must not be called on invalid input, calls %d", repo.calls)
	}
}
