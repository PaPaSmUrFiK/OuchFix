package unit

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	gql "github.com/PaPaSmUrFiK/OuchFix/internal/gateway/transport/graphql"
	"github.com/PaPaSmUrFiK/OuchFix/internal/gateway/transport/graphql/resolver"
	pkgerrors "github.com/PaPaSmUrFiK/OuchFix/pkg/errors"
)

// fakeProxy — сервис за прокси gateway: отдаёт готовый JSON по пути и
// запоминает, какие запросы пришли.
type fakeProxy struct {
	responses map[string]string
	errs      map[string]error
	calls     []string
}

func (f *fakeProxy) Get(_ context.Context, path string, query url.Values, out any) error {
	call := path
	if len(query) > 0 {
		call += "?" + query.Encode()
	}
	f.calls = append(f.calls, call)
	if err, ok := f.errs[path]; ok {
		return err
	}
	body, ok := f.responses[path]
	if !ok {
		return fmt.Errorf("%s: %w", path, pkgerrors.ErrNotFound)
	}
	return json.Unmarshal([]byte(body), out)
}

// Post — бизнес-операция: запрос запоминается как "POST path body".
func (f *fakeProxy) Post(_ context.Context, path string, body any, out any) error {
	raw, err := json.Marshal(body)
	if err != nil {
		return err
	}
	f.calls = append(f.calls, "POST "+path+" "+string(raw))
	if err, ok := f.errs[path]; ok {
		return err
	}
	resp, ok := f.responses[path]
	if !ok {
		return fmt.Errorf("%s: %w", path, pkgerrors.ErrNotFound)
	}
	return json.Unmarshal([]byte(resp), out)
}

type fakeGateway struct {
	catalog, appointment, payment, notification *fakeProxy
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

func newGatewayHandler(t *testing.T, userID int64) (http.Handler, *fakeGateway) {
	t.Helper()
	f := &fakeGateway{
		catalog: &fakeProxy{responses: map[string]string{
			"/doctors":            `[{"id":"2","fullName":"Иванов","office":null,"specialty":{"id":"1","name":"Терапия"}}]`,
			"/patients/by-user/5": `{"id":"1"}`,
		}},
		appointment: &fakeProxy{responses: map[string]string{
			"/appointments/1":          `{"id":"1","patientId":"1","doctorId":"2","status":"NO_SHOW","visitType":"WALK_IN","diagnoses":[{"id":"1","description":"ОРВИ"}],"prescriptions":[]}`,
			"/doctors/2/free-slots":    `[{"id":"1","doctorId":"2","startsAt":"2026-11-02T09:00:00+03:00","endsAt":"2026-11-02T09:30:00+03:00","status":"FREE"}]`,
			"/patients/1/appointments": `[]`,
		}},
		payment:      &fakeProxy{errs: map[string]error{"/payments": errors.New("pq: connection refused to 10.0.0.1")}},
		notification: &fakeProxy{responses: map[string]string{"/users/5/notifications": `[]`}},
	}
	return gql.NewHandler(resolver.New(resolver.Services{
		Catalog:      f.catalog,
		Appointment:  f.appointment,
		Payment:      f.payment,
		Notification: f.notification,
	}, userID)), f
}

func TestGraphQLDoctorsWithSpecialty(t *testing.T) {
	h, f := newGatewayHandler(t, 5)
	resp := doGraphQL(t, h, `{ doctors(specialtyId: "1") { id fullName specialty { id name } } }`)
	if len(resp.Errors) != 0 {
		t.Fatalf("errors: %+v", resp.Errors)
	}
	if len(f.catalog.calls) != 1 || f.catalog.calls[0] != "/doctors?specialtyId=1" {
		t.Fatalf("catalog calls = %v", f.catalog.calls)
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
	if len(f.appointment.calls) != 0 {
		t.Fatal("service must not be called with invalid id")
	}
}

// Gateway ничего не преобразует: JSON сервиса уходит клиенту как есть.
func TestGraphQLProxiesServiceJSON(t *testing.T) {
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
	if got := f.appointment.calls; len(got) != 1 || got[0] != "/patients/1/appointments?status=NO_SHOW" {
		t.Fatalf("myAppointments calls = %v", got)
	}
	if got := f.notification.calls; len(got) != 1 || got[0] != "/users/5/notifications?status=READ" {
		t.Fatalf("myNotifications calls = %v", got)
	}
}

func TestGraphQLNotImplementedOperations(t *testing.T) {
	h, _ := newGatewayHandler(t, 5)
	for _, q := range []string{
		`{ me { id } }`,
		`mutation { login(input: {email: "a", password: "b"}) { token } }`,
	} {
		if resp := doGraphQL(t, h, q); len(resp.Errors) != 1 {
			t.Fatalf("%s: want 1 error, got %+v", q, resp.Errors)
		}
	}
}

func TestGraphQLErrorCodes(t *testing.T) {
	h, _ := newGatewayHandler(t, 5)
	for _, tc := range []struct {
		query string
		code  string
	}{
		{`{ appointment(id: "99") { id } }`, "NOT_FOUND"},
		{`{ appointment(id: "abc") { id } }`, "BAD_USER_INPUT"},
		{`{ me { id } }`, "NOT_IMPLEMENTED"},
		{`{ payments { id } }`, "INTERNAL"},
		{`{ unknownField }`, "GRAPHQL_VALIDATION_FAILED"},
	} {
		resp := doGraphQL(t, h, tc.query)
		if len(resp.Errors) != 1 {
			t.Fatalf("%s: want 1 error, got %+v", tc.query, resp.Errors)
		}
		if got := resp.Errors[0].Extensions["code"]; got != tc.code {
			t.Fatalf("%s: code = %v, want %s", tc.query, got, tc.code)
		}
		if tc.code == "INTERNAL" && strings.Contains(resp.Errors[0].Message, "10.0.0.1") {
			t.Fatalf("internal details leaked: %s", resp.Errors[0].Message)
		}
	}
}

func TestGraphQLUnknownUserIsNotFound(t *testing.T) {
	h, _ := newGatewayHandler(t, 42)
	resp := doGraphQL(t, h, `{ myAppointments { id } }`)
	if len(resp.Errors) != 1 || resp.Errors[0].Extensions["code"] != "NOT_FOUND" {
		t.Fatalf("got %+v", resp.Errors)
	}
}

// Mutation только пересылает input сервису и отдаёт его ответ; бизнес-ошибки
// сервиса доходят до клиента со своими кодами.
func TestGraphQLMutationsProxyInput(t *testing.T) {
	h, f := newGatewayHandler(t, 5)
	f.appointment.responses["/appointments"] = `{"id":"6","patientId":"1","doctorId":"1","slotId":"2","status":"SCHEDULED","visitType":"SCHEDULED","diagnoses":[],"prescriptions":[]}`
	f.appointment.errs = map[string]error{
		"/appointments/2/cancel": fmt.Errorf("%w: нельзя отменить приём в статусе completed", pkgerrors.ErrFailedPrecondition),
		"/slots":                 fmt.Errorf("%w: время пересекается со слотом 3", pkgerrors.ErrConflict),
	}

	resp := doGraphQL(t, h, `mutation { bookAppointment(input: {patientId: "1", doctorId: "1", slotId: "2"}) { id status } }`)
	if len(resp.Errors) != 0 || !strings.Contains(string(resp.Data["bookAppointment"]), `"status":"SCHEDULED"`) {
		t.Fatalf("book: %+v %s", resp.Errors, resp.Data["bookAppointment"])
	}
	if got := f.appointment.calls[0]; got != `POST /appointments {"patientId":"1","doctorId":"1","slotId":"2"}` {
		t.Fatalf("book call = %s", got)
	}

	for _, tc := range []struct{ query, code string }{
		{`mutation { cancelAppointment(id: "2") { id } }`, "FAILED_PRECONDITION"},
		{`mutation { createSlots(input: {doctorId: "1", startsAt: "a", endsAt: "b"}) { id } }`, "CONFLICT"},
		{`mutation { cancelAppointment(id: "x") { id } }`, "BAD_USER_INPUT"},
	} {
		resp := doGraphQL(t, h, tc.query)
		if len(resp.Errors) != 1 || resp.Errors[0].Extensions["code"] != tc.code {
			t.Fatalf("%s: want %s, got %+v", tc.query, tc.code, resp.Errors)
		}
	}
}

func TestGraphQLAppointmentQueriesProxy(t *testing.T) {
	h, f := newGatewayHandler(t, 5)
	f.appointment.responses["/appointments"] = `[]`
	f.appointment.responses["/stats/doctor-load"] = `[{"doctorId":"1","fullName":"Петров","appointmentCount":2}]`
	f.appointment.responses["/stats/popular-services"] = `[]`
	q := `{ appointments(doctorId: "1", date: "2030-03-01", status: SCHEDULED, sort: SCHEDULED_AT_DESC) { id }
	        doctorLoad(from: "2030-03-01", to: "2030-03-31") { fullName appointmentCount }
	        popularServices(from: "2030-03-01", to: "2030-03-31", limit: 3) { name } }`
	if resp := doGraphQL(t, h, q); len(resp.Errors) != 0 {
		t.Fatalf("errors: %+v", resp.Errors)
	}
	want := map[string]bool{
		"/appointments?date=2030-03-01&doctorId=1&sort=SCHEDULED_AT_DESC&status=SCHEDULED": true,
		"/stats/doctor-load?from=2030-03-01&to=2030-03-31":                                 true,
		"/stats/popular-services?from=2030-03-01&limit=3&to=2030-03-31":                    true,
	}
	for _, c := range f.appointment.calls {
		delete(want, c)
	}
	if len(want) != 0 {
		t.Fatalf("missing calls %v, got %v", want, f.appointment.calls)
	}
}

func TestGraphQLMedicationMutationsProxy(t *testing.T) {
	h, f := newGatewayHandler(t, 5)
	med := `{"id":"1","name":"Парацетамол","form":null,"unit":"шт","stockQty":5}`
	f.catalog.responses["/medications"] = med
	f.catalog.responses["/medications/dispense"] = med
	q := `mutation {
	  a: upsertMedication(name: "Парацетамол", unit: "шт", stockQty: 5) { id }
	  b: dispenseMedication(input: {medicationId: "1", patientId: "2", quantity: 1.5}) { stockQty } }`
	if resp := doGraphQL(t, h, q); len(resp.Errors) != 0 {
		t.Fatalf("errors: %+v", resp.Errors)
	}
	want := []string{
		`POST /medications {"form":null,"name":"Парацетамол","stockQty":5,"unit":"шт"}`,
		`POST /medications/dispense {"medicationId":"1","patientId":"2","quantity":1.5}`,
	}
	if len(f.catalog.calls) != 2 || f.catalog.calls[0] != want[0] || f.catalog.calls[1] != want[1] {
		t.Fatalf("calls = %q", f.catalog.calls)
	}
}
