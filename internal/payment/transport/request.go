package transport

import (
	"strings"

	"github.com/PaPaSmUrFiK/OuchFix/internal/payment/domain"
	"github.com/PaPaSmUrFiK/OuchFix/internal/payment/service"
	httpserver "github.com/PaPaSmUrFiK/OuchFix/pkg/transport/http/server"
)

// createPaymentRequest — GraphQL CreatePaymentInput как есть: id строками, method UPPERCASE.
type createPaymentRequest struct {
	AppointmentID string  `json:"appointmentId"`
	PatientID     string  `json:"patientId"`
	ServiceID     *string `json:"serviceId"`
	Amount        float64 `json:"amount"`
	Method        string  `json:"method"`
}

func (r createPaymentRequest) toInput() (service.CreateInput, error) {
	var (
		in  = service.CreateInput{Amount: r.Amount, Method: domain.Method(strings.ToLower(r.Method))}
		err error
	)
	if in.AppointmentID, err = httpserver.ParseID("appointmentId", r.AppointmentID); err != nil {
		return in, err
	}
	if in.PatientID, err = httpserver.ParseID("patientId", r.PatientID); err != nil {
		return in, err
	}
	in.ServiceID, err = httpserver.ParseOptionalID("serviceId", r.ServiceID)
	return in, err
}
