package unit

import (
	"context"
	"errors"
	"math"
	"testing"

	"github.com/PaPaSmUrFiK/OuchFix/internal/catalog/domain"
	"github.com/PaPaSmUrFiK/OuchFix/internal/catalog/service"
	pkgerrors "github.com/PaPaSmUrFiK/OuchFix/pkg/errors"
)

type fakeStockRepo struct {
	service.CatalogRepository

	stock      map[int64]*domain.Medication
	dispensed  []domain.Dispense
	upserted   *domain.Medication
	inTx       bool
	failInsert bool
}

func (f *fakeStockRepo) InTx(ctx context.Context, fn func(context.Context) error) error {
	f.inTx = true
	snapshot := map[int64]domain.Medication{}
	for id, m := range f.stock {
		snapshot[id] = *m
	}
	err := fn(ctx)
	f.inTx = false
	if err != nil { // откат
		for id, m := range snapshot {
			*f.stock[id] = m
		}
	}
	return err
}

func (f *fakeStockRepo) GetPatient(_ context.Context, id int64) (*domain.Patient, error) {
	if id > 3 {
		return nil, pkgerrors.ErrNotFound
	}
	return &domain.Patient{ID: id}, nil
}

func (f *fakeStockRepo) GetMedication(_ context.Context, id int64) (*domain.Medication, error) {
	m, ok := f.stock[id]
	if !ok {
		return nil, pkgerrors.ErrNotFound
	}
	out := *m
	return &out, nil
}

func (f *fakeStockRepo) TakeStock(_ context.Context, id int64, qty float64) (*domain.Medication, error) {
	if !f.inTx {
		return nil, errors.New("take stock outside tx")
	}
	m, ok := f.stock[id]
	if !ok || m.StockQty < qty {
		return nil, pkgerrors.ErrNotFound
	}
	m.StockQty -= qty
	out := *m
	return &out, nil
}

func (f *fakeStockRepo) InsertDispense(_ context.Context, d domain.Dispense) error {
	if f.failInsert {
		return errors.New("db down")
	}
	f.dispensed = append(f.dispensed, d)
	return nil
}

func (f *fakeStockRepo) ReceiveMedication(_ context.Context, m domain.Medication, qty float64) (*domain.Medication, error) {
	m.StockQty = qty
	f.upserted = &m
	return &m, nil
}

func TestDispenseMedication(t *testing.T) {
	repo := &fakeStockRepo{stock: map[int64]*domain.Medication{1: {ID: 1, Unit: "шт", StockQty: 10}}}
	svc := service.NewCatalogService(repo)
	ctx := context.Background()

	m, err := svc.DispenseMedication(ctx, domain.Dispense{MedicationID: 1, PatientID: 1, Quantity: 4})
	if err != nil || m.StockQty != 6 || len(repo.dispensed) != 1 {
		t.Fatalf("dispense: %v %+v %+v", err, m, repo.dispensed)
	}
	if _, err := svc.DispenseMedication(ctx, domain.Dispense{MedicationID: 1, PatientID: 1, Quantity: 7}); !errors.Is(err, pkgerrors.ErrFailedPrecondition) {
		t.Fatalf("not enough stock: %v", err)
	}
	if _, err := svc.DispenseMedication(ctx, domain.Dispense{MedicationID: 9, PatientID: 1, Quantity: 1}); !errors.Is(err, pkgerrors.ErrNotFound) {
		t.Fatalf("unknown medication: %v", err)
	}
	if _, err := svc.DispenseMedication(ctx, domain.Dispense{MedicationID: 1, PatientID: 9, Quantity: 1}); !errors.Is(err, pkgerrors.ErrNotFound) {
		t.Fatalf("unknown patient: %v", err)
	}
	for _, q := range []float64{0, -1, math.NaN(), math.Inf(1)} {
		if _, err := svc.DispenseMedication(ctx, domain.Dispense{MedicationID: 1, PatientID: 1, Quantity: q}); !errors.Is(err, pkgerrors.ErrInvalidArgument) {
			t.Fatalf("quantity %v: %v", q, err)
		}
	}

	repo.failInsert = true
	if _, err := svc.DispenseMedication(ctx, domain.Dispense{MedicationID: 1, PatientID: 1, Quantity: 1}); err == nil {
		t.Fatal("insert failure must fail dispense")
	}
	if repo.stock[1].StockQty != 6 {
		t.Fatalf("stock must roll back, got %v", repo.stock[1].StockQty)
	}
}

func TestReceiveMedicationValidation(t *testing.T) {
	repo := &fakeStockRepo{}
	svc := service.NewCatalogService(repo)
	ctx := context.Background()

	blank := "  "
	m, err := svc.ReceiveMedication(ctx, domain.Medication{Name: " Аспирин ", Form: &blank, Unit: " шт "}, 3)
	if err != nil || m.Name != "Аспирин" || m.Unit != "шт" || m.Form != nil || m.StockQty != 3 {
		t.Fatalf("receive: %v %+v", err, m)
	}
	repo.upserted = nil
	for _, bad := range []struct {
		m   domain.Medication
		qty float64
	}{
		{domain.Medication{Name: "", Unit: "шт"}, 1},
		{domain.Medication{Name: "A", Unit: " "}, 1},
		{domain.Medication{Name: "A", Unit: "шт"}, 0},
		{domain.Medication{Name: "A", Unit: "шт"}, -1},
		{domain.Medication{Name: "A", Unit: "шт"}, math.NaN()},
	} {
		if _, err := svc.ReceiveMedication(ctx, bad.m, bad.qty); !errors.Is(err, pkgerrors.ErrInvalidArgument) {
			t.Fatalf("%+v %v: %v", bad.m, bad.qty, err)
		}
	}
	if repo.upserted != nil {
		t.Fatal("repository must not be called on invalid input")
	}
}
