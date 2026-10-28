package unit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	appointmentdomain "github.com/PaPaSmUrFiK/OuchFix/internal/appointment/domain"
	catalogdomain "github.com/PaPaSmUrFiK/OuchFix/internal/catalog/domain"
	gql "github.com/PaPaSmUrFiK/OuchFix/internal/gateway/transport/graphql"
	"github.com/PaPaSmUrFiK/OuchFix/internal/gateway/transport/graphql/resolver"
	notificationdomain "github.com/PaPaSmUrFiK/OuchFix/internal/notification/domain"
	paymentdomain "github.com/PaPaSmUrFiK/OuchFix/internal/payment/domain"
	pkgerrors "github.com/PaPaSmUrFiK/OuchFix/pkg/errors"
)

type fakeGatewayServices struct {
	gotSpecialtyID   *int64
	gotPatientID     int64
	gotApptStatus    *appointmentdomain.Status
	gotUserID        int64
	gotNotifStatus   *notificationdomain.Status
	appointmentCalls int
}

func (f *fakeGatewayServices) ListSpecialties(context.Context) ([]catalogdomain.Specialty, error) {
	return []catalogdomain.Specialty{{ID: 1, Name: "Терапия"}}, nil
}

func (f *fakeGatewayServices) ListDoctors(_ context.Context, specialtyID *int64) ([]catalogdomain.Doctor, error) {
	f.gotSpecialtyID = specialtyID
	return []catalogdomain.Doctor{{ID: 2, FullName: "Иванов", Specialty: catalogdomain.Specialty{ID: 1, Name: "Терапия"}}}, nil
}

func (f *fakeGatewayServices) GetPatientByUserID(_ context.Context, userID int64) (*catalogdomain.Patient, error) {
	if userID != 5 {
		return nil, fmt.Errorf("patient by user %d: %w", userID, pkgerrors.ErrNotFound)
	}
	return &catalogdomain.Patient{ID: 1, UserID: 5}, nil
}

func (f *fakeGatewayServices) ListMedicalServices(context.Context) ([]catalogdomain.MedicalService, error) {
	return nil, nil
}

func (f *fakeGatewayServices) ListMedications(context.Context) ([]catalogdomain.Medication, error) {
	return nil, nil
}

func (f *fakeGatewayServices) FreeSlots(_ context.Context, doctorID int64, _ string) ([]appointmentdomain.Slot, error) {
	start := time.Date(2026, 11, 2, 6, 0, 0, 0, time.UTC)
	return []appointmentdomain.Slot{{ID: 1, DoctorID: doctorID, StartsAt: start, EndsAt: start.Add(30 * time.Minute), Status: appointmentdomain.SlotFree}}, nil
}

func (f *fakeGatewayServices) GetAppointment(_ context.Context, id int64) (*appointmentdomain.Appointment, error) {
	f.appointmentCalls++
	if id != 1 {
		return nil, fmt.Errorf("appointment %d: %w", id, pkgerrors.ErrNotFound)
	}
	return &appointmentdomain.Appointment{
		ID: 1, PatientID: 1, DoctorID: 2,
		Status: appointmentdomain.StatusNoShow, VisitType: appointmentdomain.VisitWalkIn,
		Diagnoses: []appointmentdomain.Diagnosis{{ID: 1, Description: "ОРВИ"}},
	}, nil
}

func (f *fakeGatewayServices) ListPatientAppointments(_ context.Context, patientID int64, status *appointmentdomain.Status) ([]appointmentdomain.Appointment, error) {
	f.gotPatientID, f.gotApptStatus = patientID, status
	return nil, nil
}

func (f *fakeGatewayServices) ListPayments(context.Context, *int64) ([]paymentdomain.Payment, error) {
	return nil, nil
}

func (f *fakeGatewayServices) ListUserNotifications(_ context.Context, userID int64, status *notificationdomain.Status) ([]notificationdomain.Notification, error) {
	f.gotUserID, f.gotNotifStatus = userID, status
	return nil, nil
}

type gqlResponse struct {
	Data   map[string]json.RawMessage `json:"data"`
	Errors []struct {
		Message    string         `json:"message"`
		Extensions map[string]any `json:"extensions"`
	} `json:"errors"`
}

func doGraphQL(t *testing.T, h http.Handler, query string) gqlResponse {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"query": query})
	req := httptest.NewRequest(http.MethodPost, "/graphql", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	var resp gqlResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response %q: %v", rec.Body.String(), err)
	}
	return resp
}

func newGatewayHandler(t *testing.T, userID int64) (http.Handler, *fakeGatewayServices) {
	t.Helper()
	loc, err := time.LoadLocation("Europe/Minsk")
	if err != nil {
		t.Fatal(err)
	}
	f := &fakeGatewayServices{}
	return gql.NewHandler(resolver.New(f, f, f, f, loc, userID)), f
}

func TestGraphQLDoctorsWithSpecialty(t *testing.T) {
	h, f := newGatewayHandler(t, 5)
	resp := doGraphQL(t, h, `{ doctors(specialtyId: "1") { id fullName specialty { id name } } }`)
	if len(resp.Errors) != 0 {
		t.Fatalf("errors: %+v", resp.Errors)
	}
	if f.gotSpecialtyID == nil || *f.gotSpecialtyID != 1 {
		t.Fatalf("specialtyId not passed: %v", f.gotSpecialtyID)
	}
	if got := string(resp.Data["doctors"]); !strings.Contains(got, `"specialty":{"id":"1","name":"Терапия"}`) {
		t.Fatalf("doctors = %s", got)
	}
}

func TestGraphQLInvalidIDIsNotPassedToService(t *testing.T) {
	h, f := newGatewayHandler(t, 5)
	resp := doGraphQL(t, h, `{ appointment(id: "abc") { id } }`)
	if len(resp.Errors) != 1 {
		t.Fatalf("want 1 error, got %+v", resp.Errors)
	}
	if f.appointmentCalls != 0 {
		t.Fatal("service must not be called with invalid id")
	}
}

func TestGraphQLEnumsAndTimeFormat(t *testing.T) {
	h, _ := newGatewayHandler(t, 5)
	resp := doGraphQL(t, h, `{ appointment(id: "1") { status visitType diagnoses { description } } freeSlots(doctorId: "2", date: "2026-11-02") { startsAt status } }`)
	if len(resp.Errors) != 0 {
		t.Fatalf("errors: %+v", resp.Errors)
	}
	if got := string(resp.Data["appointment"]); !strings.Contains(got, `"status":"NO_SHOW","visitType":"WALK_IN"`) {
		t.Fatalf("appointment = %s", got)
	}
	if got := string(resp.Data["freeSlots"]); !strings.Contains(got, `"startsAt":"2026-11-02T09:00:00+03:00","status":"FREE"`) {
		t.Fatalf("freeSlots = %s", got)
	}
}

func TestGraphQLMyQueriesUseConfiguredUser(t *testing.T) {
	h, f := newGatewayHandler(t, 5)
	resp := doGraphQL(t, h, `{ myAppointments(status: NO_SHOW) { id } myNotifications(status: READ) { id } }`)
	if len(resp.Errors) != 0 {
		t.Fatalf("errors: %+v", resp.Errors)
	}
	if f.gotPatientID != 1 || f.gotApptStatus == nil || *f.gotApptStatus != appointmentdomain.StatusNoShow {
		t.Fatalf("myAppointments: patient=%d status=%v", f.gotPatientID, f.gotApptStatus)
	}
	if f.gotUserID != 5 || f.gotNotifStatus == nil || *f.gotNotifStatus != notificationdomain.StatusRead {
		t.Fatalf("myNotifications: user=%d status=%v", f.gotUserID, f.gotNotifStatus)
	}
}

func TestGraphQLNotImplementedOperations(t *testing.T) {
	h, _ := newGatewayHandler(t, 5)
	for _, q := range []string{
		`{ me { id } }`,
		`{ appointments { id } }`,
		`mutation { login(input: {email: "a", password: "b"}) { token } }`,
	} {
		if resp := doGraphQL(t, h, q); len(resp.Errors) != 1 {
			t.Fatalf("%s: want 1 error, got %+v", q, resp.Errors)
		}
	}
}
