package transport

import (
	"time"

	"github.com/PaPaSmUrFiK/OuchFix/internal/payment/domain"
	httpserver "github.com/PaPaSmUrFiK/OuchFix/pkg/transport/http/server"
)

// paymentResponse — в форме GraphQL-модели Payment: gateway декодирует её без преобразований.
type paymentResponse struct {
	ID            string  `json:"id"`
	AppointmentID string  `json:"appointmentId"`
	PatientID     string  `json:"patientId"`
	ServiceID     *string `json:"serviceId"`
	Amount        float64 `json:"amount"`
	Method        string  `json:"method"`
	Status        string  `json:"status"`
	PaidAt        *string `json:"paidAt"`
}

func toPayment(loc *time.Location) func(*domain.Payment) paymentResponse {
	return func(p *domain.Payment) paymentResponse {
		return paymentResponse{
			ID:            httpserver.FormatID(p.ID),
			AppointmentID: httpserver.FormatID(p.AppointmentID),
			PatientID:     httpserver.FormatID(p.PatientID),
			ServiceID:     httpserver.FormatOptionalID(p.ServiceID),
			Amount:        p.Amount,
			Method:        httpserver.FormatEnum(p.Method),
			Status:        httpserver.FormatEnum(p.Status),
			PaidAt:        httpserver.FormatOptionalTime(p.PaidAt, loc),
		}
	}
}
