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
	lastFilter  paymentdomain.Filter

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

func (f *fakePaymentRepo) List(_ context.Context, filter paymentdomain.Filter) ([]paymentdomain.Payment, error) {
	f.calls++
	f.lastPatient = filter.PatientID
	f.lastFilter = filter
	return []paymentdomain.Payment{}, nil
}

func TestPaymentsFilterAndValidation(t *testing.T) {
	repo := &fakePaymentRepo{}
	svc := paymentsvc.NewPaymentService(repo, catalog.NewStub(), appointment.NewStub(), time.UTC)

	if _, err := svc.ListPayments(context.Background(), paymentsvc.ListInput{}); err != nil || repo.lastPatient != nil {
		t.Fatalf("nil patient must list all: err %v", err)
	}

	bad := int64(-5)
	if _, err := svc.ListPayments(context.Background(), paymentsvc.ListInput{PatientID: &bad}); !errors.Is(err, pkgerrors.ErrInvalidArgument) {
		t.Fatalf("want ErrInvalidArgument, got %v", err)
	}
	unknown := int64(404)
	if _, err := svc.ListPayments(context.Background(), paymentsvc.ListInput{PatientID: &unknown}); !errors.Is(err, pkgerrors.ErrNotFound) {
		t.Fatalf("want ErrNotFound for unknown patient, got %v", err)
	}
	str := func(v string) *string { return &v }
	for _, bad := range []paymentsvc.ListInput{
		{Status: str("REFUNDED")},
		{From: str("01.11.2026")},
		{To: str("2026-13-01")},
		{From: str("2026-11-10"), To: str("2026-11-09")},
	} {
		if _, err := svc.ListPayments(context.Background(), bad); !errors.Is(err, pkgerrors.ErrInvalidArgument) {
			t.Fatalf("%+v: want ErrInvalidArgument, got %v", bad, err)
		}
	}
	if repo.calls != 1 {
		t.Fatalf("repository must not be called on invalid input, calls %d", repo.calls)
	}

	if _, err := svc.ListPayments(context.Background(), paymentsvc.ListInput{
		Status: str("PAID"), From: str("2026-11-01"), To: str("2026-11-01"),
	}); err != nil {
		t.Fatal(err)
	}
	f := repo.lastFilter
	if f.Status == nil || *f.Status != paymentdomain.StatusPaid ||
		!f.From.Equal(time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)) ||
		!f.To.Equal(time.Date(2026, 11, 2, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("filter = %+v", f)
	}
}

type fakeNotificationRepo struct {
	calls int
	items map[int64]notificationdomain.Notification
}

func (f *fakeNotificationRepo) MarkRead(_ context.Context, userID, id int64) (*notificationdomain.Notification, error) {
	n, ok := f.items[id]
	if !ok || n.UserID != userID || n.Status == notificationdomain.StatusFailed {
		return nil, pkgerrors.ErrNotFound
	}
	n.Status = notificationdomain.StatusRead
	f.items[id] = n
	return &n, nil
}

func (f *fakeNotificationRepo) Get(_ context.Context, userID, id int64) (*notificationdomain.Notification, error) {
	n, ok := f.items[id]
	if !ok || n.UserID != userID {
		return nil, pkgerrors.ErrNotFound
	}
	return &n, nil
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

func TestMarkNotificationRead(t *testing.T) {
	repo := &fakeNotificationRepo{items: map[int64]notificationdomain.Notification{
		1: {ID: 1, UserID: 5, Status: notificationdomain.StatusSent},
		4: {ID: 4, UserID: 6, Status: notificationdomain.StatusFailed},
		7: {ID: 7, UserID: 5, Status: notificationdomain.StatusFailed},
	}}
	svc := notificationsvc.NewNotificationService(repo, iam.NewStub())
	ctx := context.Background()

	for i := 0; i < 2; i++ { // повтор идемпотентен
		n, err := svc.MarkRead(ctx, 5, 1)
		if err != nil || n.Status != notificationdomain.StatusRead {
			t.Fatalf("try %d: %v %+v", i, err, n)
		}
	}
	if _, err := svc.MarkRead(ctx, 5, 4); !errors.Is(err, pkgerrors.ErrNotFound) {
		t.Fatalf("foreign notification: %v", err)
	}
	if _, err := svc.MarkRead(ctx, 5, 7); !errors.Is(err, pkgerrors.ErrFailedPrecondition) {
		t.Fatalf("failed notification: %v", err)
	}
	if _, err := svc.MarkRead(ctx, 5, 0); !errors.Is(err, pkgerrors.ErrInvalidArgument) {
		t.Fatalf("bad id: %v", err)
	}
}
