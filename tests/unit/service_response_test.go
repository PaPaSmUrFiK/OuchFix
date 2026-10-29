package unit

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/PaPaSmUrFiK/OuchFix/internal/appointment/domain"
	"github.com/PaPaSmUrFiK/OuchFix/internal/appointment/service"
	"github.com/PaPaSmUrFiK/OuchFix/internal/appointment/transport"
	"github.com/PaPaSmUrFiK/OuchFix/pkg/client/catalog"
)

// Сервис отвечает в форме GraphQL-модели: camelCase, id строкой, enum в
// верхнем регистре, время RFC 3339 в поясе медцентра, пустые списки — [].
func TestAppointmentRESTIsGraphQLShaped(t *testing.T) {
	at := time.Date(2026, 11, 2, 6, 0, 0, 0, time.UTC)
	repo := &fakeAppointmentRepo{appointments: map[int64]domain.Appointment{
		1: {ID: 1, PatientID: 1, DoctorID: 2, Status: domain.StatusNoShow, VisitType: domain.VisitWalkIn, ScheduledAt: &at},
	}}
	loc := minsk(t)
	app := fiber.New()
	transport.New(service.NewAppointmentService(repo, catalog.NewStub(), loc), loc)(app)

	resp, err := app.Test(httptest.NewRequest("GET", "/appointments/1", nil))
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(resp.Body)
	want := `{"id":"1","patientId":"1","doctorId":"2","slotId":null,"serviceId":null,"status":"NO_SHOW","visitType":"WALK_IN","scheduledAt":"2026-11-02T09:00:00+03:00","conclusion":null,"diagnoses":[],"prescriptions":[]}`
	if string(body) != want {
		t.Fatalf("body = %s", body)
	}

	resp, err = app.Test(httptest.NewRequest("GET", "/patients/1/appointments?status=NO_SHOW", nil))
	if err != nil {
		t.Fatal(err)
	}
	var list []json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil || len(list) != 1 {
		t.Fatalf("list = %v, %v", list, err)
	}
	if repo.lastStatus == nil || *repo.lastStatus != domain.StatusNoShow {
		t.Fatalf("status filter = %v", repo.lastStatus)
	}
}
