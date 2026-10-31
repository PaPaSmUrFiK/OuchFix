package models

import (
	"database/sql"
	"time"

	"github.com/PaPaSmUrFiK/OuchFix/internal/catalog/domain"
)

type Specialty struct {
	ID   int64  `db:"id"`
	Name string `db:"name"`
}

func (m Specialty) ToDomain() domain.Specialty {
	return domain.Specialty{ID: m.ID, Name: m.Name}
}

type Patient struct {
	ID        int64          `db:"id"`
	UserID    int64          `db:"user_id"`
	FullName  string         `db:"full_name"`
	BirthDate sql.NullTime   `db:"birth_date"`
	Gender    sql.NullString `db:"gender"`
	Phone     sql.NullString `db:"phone"`
	Address   sql.NullString `db:"address"`
	CreatedAt time.Time      `db:"created_at"`
}

func (m Patient) ToDomain() domain.Patient {
	return domain.Patient{
		ID:        m.ID,
		UserID:    m.UserID,
		FullName:  m.FullName,
		BirthDate: timePtr(m.BirthDate),
		Gender:    stringPtr(m.Gender),
		Phone:     stringPtr(m.Phone),
		Address:   stringPtr(m.Address),
		CreatedAt: m.CreatedAt,
	}
}

type Doctor struct {
	ID            int64          `db:"id"`
	UserID        int64          `db:"user_id"`
	FullName      string         `db:"full_name"`
	Office        sql.NullString `db:"office"`
	SpecialtyID   int64          `db:"specialty_id"`
	SpecialtyName string         `db:"specialty_name"`
	CreatedAt     time.Time      `db:"created_at"`
}

func (m Doctor) ToDomain() domain.Doctor {
	return domain.Doctor{
		ID:        m.ID,
		UserID:    m.UserID,
		FullName:  m.FullName,
		Office:    stringPtr(m.Office),
		Specialty: domain.Specialty{ID: m.SpecialtyID, Name: m.SpecialtyName},
		CreatedAt: m.CreatedAt,
	}
}

type MedicalService struct {
	ID          int64         `db:"id"`
	Name        string        `db:"name"`
	Price       float64       `db:"price"`
	DurationMin sql.NullInt32 `db:"duration_min"`
}

func (m MedicalService) ToDomain() domain.MedicalService {
	s := domain.MedicalService{ID: m.ID, Name: m.Name, Price: m.Price}
	if m.DurationMin.Valid {
		v := m.DurationMin.Int32
		s.DurationMin = &v
	}
	return s
}

type Medication struct {
	ID       int64          `db:"id"`
	Name     string         `db:"name"`
	Form     sql.NullString `db:"form"`
	Unit     string         `db:"unit"`
	StockQty float64        `db:"stock_qty"`
}

func (m Medication) ToDomain() domain.Medication {
	return domain.Medication{
		ID:       m.ID,
		Name:     m.Name,
		Form:     stringPtr(m.Form),
		Unit:     m.Unit,
		StockQty: m.StockQty,
	}
}

// ToDomainList переводит строки БД в доменные модели.
func ToDomainList[M interface{ ToDomain() D }, D any](rows []M) []D {
	out := make([]D, 0, len(rows))
	for _, row := range rows {
		out = append(out, row.ToDomain())
	}
	return out
}

func stringPtr(v sql.NullString) *string {
	if !v.Valid {
		return nil
	}
	s := v.String
	return &s
}

func timePtr(v sql.NullTime) *time.Time {
	if !v.Valid {
		return nil
	}
	t := v.Time
	return &t
}
