package transport

import (
	"github.com/gofiber/fiber/v2"

	"github.com/PaPaSmUrFiK/OuchFix/internal/payment/service"
	httpserver "github.com/PaPaSmUrFiK/OuchFix/pkg/transport/http/server"
)

// New — внутренний REST API платежей для gateway. В ЛР6 заменится на gRPC.
func New(svc *service.PaymentService) func(fiber.Router) {
	return func(router fiber.Router) {
		router.Get("/health", httpserver.Health)

		router.Get("/payments", func(c *fiber.Ctx) error {
			patientID, err := httpserver.QueryID(c, "patientId")
			if err != nil {
				return httpserver.WriteError(c, err)
			}
			items, err := svc.ListPayments(c.UserContext(), patientID)
			return httpserver.Reply(c, items, err)
		})
	}
}
