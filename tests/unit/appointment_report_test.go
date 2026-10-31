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

func TestListAppointmentsFilter(t *testing.T) {
	repo := &fakeAppointmentRepo{}
	svc := service.NewAppointmentService(repo, catalog.NewStub(), minsk(t))
	ctx := context.Background()

	if _, err := svc.ListAppointments(ctx, service.ListInput{}); err != nil {
		t.Fatal(err)
	}
	if repo.lastFilter.Sort != domain.SortScheduledAtAsc || repo.lastFilter.From != nil {
		t.Fatalf("default filter = %+v", repo.lastFilter)
	}

	doctor, date, st, sort := int64(2), "2030-03-01", domain.StatusScheduled, domain.SortScheduledAtDesc
	if _, err := svc.ListAppointments(ctx, service.ListInput{DoctorID: &doctor, Date: &date, Status: &st, Sort: &sort}); err != nil {
		t.Fatal(err)
	}
	f := repo.lastFilter
	day := time.Date(2030, 3, 1, 0, 0, 0, 0, minsk(t))
	if *f.DoctorID != 2 || *f.Status != st || f.Sort != sort || !f.From.Equal(day) || !f.To.Equal(day.AddDate(0, 0, 1)) {
		t.Fatalf("filter = %+v", f)
	}

	bad := []service.ListInput{
		{Date: ptr("01.03.2030")},
		{Status: ptr(domain.Status("lost"))},
		{Sort: ptr(domain.Sort("name"))},
		{DoctorID: ptr(int64(0))},
	}
	for _, in := range bad {
		if _, err := svc.ListAppointments(ctx, in); !errors.Is(err, pkgerrors.ErrInvalidArgument) {
			t.Fatalf("%+v: want invalid argument, got %v", in, err)
		}
	}
}

func TestReportsPeriodAndNames(t *testing.T) {
	repo := &fakeAppointmentRepo{
		load:    []domain.DoctorLoad{{DoctorID: 1, AppointmentCount: 3}},
		popular: []domain.ServiceStat{{ServiceID: 2, BookingCount: 4}},
	}
	svc := service.NewAppointmentService(repo, catalog.NewStub(), minsk(t))
	ctx := context.Background()

	load, err := svc.DoctorLoad(ctx, "2030-03-01", "2030-03-31")
	if err != nil {
		t.Fatal(err)
	}
	if load[0].FullName == "" {
		t.Fatalf("doctor name not attached: %+v", load)
	}
	from := time.Date(2030, 3, 1, 0, 0, 0, 0, minsk(t))
	if !repo.reportFrom.Equal(from) || !repo.reportTo.Equal(time.Date(2030, 4, 1, 0, 0, 0, 0, minsk(t))) {
		t.Fatalf("period = %s..%s", repo.reportFrom, repo.reportTo)
	}

	stats, err := svc.PopularServices(ctx, "2030-03-01", "2030-03-01", nil)
	if err != nil {
		t.Fatal(err)
	}
	if stats[0].Name == "" || repo.reportLimit != 5 {
		t.Fatalf("stats = %+v, limit = %d", stats, repo.reportLimit)
	}

	for _, tc := range []struct {
		from, to string
		limit    *int
	}{
		{"2030-03-02", "2030-03-01", nil},
		{"2030-03-01", "bad", nil},
		{"2030-01-01", "2031-06-01", nil},
		{"2030-03-01", "2030-03-01", ptr(0)},
		{"2030-03-01", "2030-03-01", ptr(51)},
	} {
		if _, err := svc.PopularServices(ctx, tc.from, tc.to, tc.limit); !errors.Is(err, pkgerrors.ErrInvalidArgument) {
			t.Fatalf("%+v: want invalid argument, got %v", tc, err)
		}
	}
	if _, err := svc.DoctorLoad(ctx, "2030-03-02", "2030-03-01"); !errors.Is(err, pkgerrors.ErrInvalidArgument) {
		t.Fatalf("doctorLoad from>to: %v", err)
	}
}
