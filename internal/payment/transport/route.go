package transport

import (
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/PaPaSmUrFiK/OuchFix/internal/payment/service"
	httpserver "github.com/PaPaSmUrFiK/OuchFix/pkg/transport/http/server"
)

// New — внутренний REST API платежей, к которому gateway проксирует запросы.
// loc — пояс медцентра, в нём отдаётся время.
func New(svc *service.PaymentService, loc *time.Location) func(fiber.Router) {
	return func(router fiber.Router) {
		router.Get("/health", httpserver.Health)

		router.Get("/payments", func(c fiber.Ctx) error {
			patientID, err := httpserver.QueryID(c, "patientId")
			if err != nil {
				return httpserver.WriteError(c, err)
			}
			items, err := svc.ListPayments(c.Context(), patientID)
			return httpserver.Reply(c, httpserver.MapList(items, toPayment(loc)), err)
		})
		router.Post("/payments", func(c fiber.Ctx) error {
			var req createPaymentRequest
			if err := httpserver.BindJSON(c, &req); err != nil {
				return httpserver.WriteError(c, err)
			}
			in, err := req.toInput()
			if err != nil {
				return httpserver.WriteError(c, err)
			}
			p, err := svc.CreatePayment(c.Context(), in)
			if err != nil {
				return httpserver.WriteError(c, err)
			}
			return httpserver.Reply(c, toPayment(loc)(p), nil)
		})
		router.Get("/payments/summary", func(c fiber.Ctx) error {
			sum, err := svc.PaymentSummary(c.Context(), c.Query("from"), c.Query("to"))
			if err != nil {
				return httpserver.WriteError(c, err)
			}
			return httpserver.Reply(c, summaryResponse{Count: sum.Count, Total: sum.Total}, nil)
		})
	}
}
