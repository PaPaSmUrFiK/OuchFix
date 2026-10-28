package client

import (
	"context"
	"net/url"
	"strconv"

	"github.com/PaPaSmUrFiK/OuchFix/internal/appointment/domain"
	httpclient "github.com/PaPaSmUrFiK/OuchFix/pkg/transport/http/client"
)

type Appointment struct {
	http *httpclient.Client
}

func NewAppointment(http *httpclient.Client) *Appointment {
	return &Appointment{http: http}
}

func (c *Appointment) FreeSlots(ctx context.Context, doctorID int64, date string) ([]domain.Slot, error) {
	var out []domain.Slot
	path := "/doctors/" + strconv.FormatInt(doctorID, 10) + "/free-slots"
	return out, c.http.Get(ctx, path, url.Values{"date": {date}}, &out)
}

func (c *Appointment) GetAppointment(ctx context.Context, id int64) (*domain.Appointment, error) {
	var out domain.Appointment
	if err := c.http.Get(ctx, "/appointments/"+strconv.FormatInt(id, 10), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Appointment) ListPatientAppointments(ctx context.Context, patientID int64, status *domain.Status) ([]domain.Appointment, error) {
	query := url.Values{}
	if status != nil {
		query.Set("status", string(*status))
	}
	var out []domain.Appointment
	path := "/patients/" + strconv.FormatInt(patientID, 10) + "/appointments"
	return out, c.http.Get(ctx, path, query, &out)
}
