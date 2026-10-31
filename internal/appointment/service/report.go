package service

import (
	"context"
	"fmt"
	"time"

	"github.com/PaPaSmUrFiK/OuchFix/internal/appointment/domain"
	pkgerrors "github.com/PaPaSmUrFiK/OuchFix/pkg/errors"
)

const (
	defaultPopularLimit = 5
	maxPopularLimit     = 50
	maxReportPeriod     = 366 * 24 * time.Hour
)

// ListInput — параметры списка приёмов: Date — день вида 2006-01-02.
type ListInput struct {
	DoctorID *int64
	Date     *string
	Status   *domain.Status
	Sort     *domain.Sort
}

// ListAppointments — расписание врача/регистратуры на день с отбором и сортировкой.
func (s *AppointmentService) ListAppointments(ctx context.Context, in ListInput) ([]domain.Appointment, error) {
	f := domain.AppointmentFilter{DoctorID: in.DoctorID, Status: in.Status, Sort: domain.SortScheduledAtAsc}
	if in.DoctorID != nil {
		if err := pkgerrors.ValidateID("doctorId", *in.DoctorID); err != nil {
			return nil, err
		}
	}
	if in.Status != nil && !in.Status.Valid() {
		return nil, fmt.Errorf("%w: неизвестный статус приёма %q", pkgerrors.ErrInvalidArgument, *in.Status)
	}
	if in.Sort != nil {
		if !in.Sort.Valid() {
			return nil, fmt.Errorf("%w: неизвестная сортировка %q", pkgerrors.ErrInvalidArgument, *in.Sort)
		}
		f.Sort = *in.Sort
	}
	if in.Date != nil {
		day, err := s.parseDay("date", *in.Date)
		if err != nil {
			return nil, err
		}
		next := day.AddDate(0, 0, 1)
		f.From, f.To = &day, &next
	}

	items, err := s.repo.ListAppointments(ctx, f)
	if err != nil {
		return nil, err
	}
	if err := s.attachDetails(ctx, items); err != nil {
		return nil, err
	}
	return items, nil
}

// DoctorLoad — нагрузка врачей за дни [from, to] включительно.
func (s *AppointmentService) DoctorLoad(ctx context.Context, from, to string) ([]domain.DoctorLoad, error) {
	start, end, err := s.parsePeriod(from, to)
	if err != nil {
		return nil, err
	}
	items, err := s.repo.DoctorLoad(ctx, start, end)
	if err != nil {
		return nil, err
	}
	for i := range items {
		doctor, err := s.catalog.GetDoctor(ctx, items[i].DoctorID)
		if err != nil {
			return nil, fmt.Errorf("doctor %d: %w", items[i].DoctorID, err)
		}
		items[i].FullName = doctor.FullName
	}
	return items, nil
}

// PopularServices — самые востребованные услуги за дни [from, to]; limit по умолчанию 5.
func (s *AppointmentService) PopularServices(ctx context.Context, from, to string, limit *int) ([]domain.ServiceStat, error) {
	n := defaultPopularLimit
	if limit != nil {
		n = *limit
	}
	if n < 1 || n > maxPopularLimit {
		return nil, fmt.Errorf("%w: limit должен быть от 1 до %d", pkgerrors.ErrInvalidArgument, maxPopularLimit)
	}
	start, end, err := s.parsePeriod(from, to)
	if err != nil {
		return nil, err
	}
	items, err := s.repo.PopularServices(ctx, start, end, n)
	if err != nil {
		return nil, err
	}
	for i := range items {
		svc, err := s.catalog.GetMedicalService(ctx, items[i].ServiceID)
		if err != nil {
			return nil, fmt.Errorf("medical service %d: %w", items[i].ServiceID, err)
		}
		items[i].Name = svc.Name
	}
	return items, nil
}

func (s *AppointmentService) parseDay(name, raw string) (time.Time, error) {
	day, err := time.ParseInLocation(dateLayout, raw, s.location)
	if err != nil {
		return time.Time{}, fmt.Errorf("%w: %s должна быть в формате YYYY-MM-DD, получено %q", pkgerrors.ErrInvalidArgument, name, raw)
	}
	return day, nil
}

// parsePeriod: дни from..to включительно в поясе медцентра → [start, end).
func (s *AppointmentService) parsePeriod(from, to string) (time.Time, time.Time, error) {
	start, err := s.parseDay("from", from)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	last, err := s.parseDay("to", to)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	if last.Before(start) {
		return time.Time{}, time.Time{}, fmt.Errorf("%w: from должна быть не позже to", pkgerrors.ErrInvalidArgument)
	}
	end := last.AddDate(0, 0, 1)
	if end.Sub(start) > maxReportPeriod {
		return time.Time{}, time.Time{}, fmt.Errorf("%w: период больше года", pkgerrors.ErrInvalidArgument)
	}
	return start, end, nil
}
