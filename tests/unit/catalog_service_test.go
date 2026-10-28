package unit

import (
	"context"
	"errors"
	"testing"

	"github.com/PaPaSmUrFiK/OuchFix/internal/catalog/domain"
	"github.com/PaPaSmUrFiK/OuchFix/internal/catalog/service"
	pkgerrors "github.com/PaPaSmUrFiK/OuchFix/pkg/errors"
)

type fakeCatalogRepo struct {
	service.CatalogRepository

	doctors        []domain.Doctor
	lastSpecialty  *int64
	listCalls      int
	getDoctorCalls int
}

func (f *fakeCatalogRepo) ListDoctors(_ context.Context, specialtyID *int64) ([]domain.Doctor, error) {
	f.listCalls++
	f.lastSpecialty = specialtyID
	return f.doctors, nil
}

func (f *fakeCatalogRepo) GetDoctor(_ context.Context, id int64) (*domain.Doctor, error) {
	f.getDoctorCalls++
	for i := range f.doctors {
		if f.doctors[i].ID == id {
			return &f.doctors[i], nil
		}
	}
	return nil, pkgerrors.ErrNotFound
}

func TestCatalogListDoctorsPassesFilter(t *testing.T) {
	repo := &fakeCatalogRepo{doctors: []domain.Doctor{{ID: 1}}}
	svc := service.NewCatalogService(repo)

	specialty := int64(2)
	got, err := svc.ListDoctors(context.Background(), &specialty)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || repo.lastSpecialty == nil || *repo.lastSpecialty != 2 {
		t.Fatalf("filter not passed: got %v, specialty %v", got, repo.lastSpecialty)
	}

	if _, err := svc.ListDoctors(context.Background(), nil); err != nil || repo.lastSpecialty != nil {
		t.Fatalf("nil filter must list all: err %v, specialty %v", err, repo.lastSpecialty)
	}
}

func TestCatalogRejectsNonPositiveIDs(t *testing.T) {
	repo := &fakeCatalogRepo{}
	svc := service.NewCatalogService(repo)

	bad := int64(0)
	if _, err := svc.ListDoctors(context.Background(), &bad); !errors.Is(err, pkgerrors.ErrInvalidArgument) {
		t.Fatalf("want ErrInvalidArgument, got %v", err)
	}
	if _, err := svc.GetDoctor(context.Background(), -1); !errors.Is(err, pkgerrors.ErrInvalidArgument) {
		t.Fatalf("want ErrInvalidArgument, got %v", err)
	}
	if repo.listCalls != 0 || repo.getDoctorCalls != 0 {
		t.Fatalf("repository must not be called on invalid input")
	}
}

func TestCatalogGetDoctorNotFound(t *testing.T) {
	svc := service.NewCatalogService(&fakeCatalogRepo{})

	if _, err := svc.GetDoctor(context.Background(), 42); !errors.Is(err, pkgerrors.ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}
