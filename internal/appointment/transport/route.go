package transport

import (
	"github.com/gofiber/fiber/v2"

	"github.com/PaPaSmUrFiK/OuchFix/internal/appointment/domain"
	"github.com/PaPaSmUrFiK/OuchFix/internal/appointment/service"
	httpserver "github.com/PaPaSmUrFiK/OuchFix/pkg/transport/http/server"
)

// New — внутренний REST API записи на приём для gateway. В ЛР6 заменится на gRPC.
func New(svc *service.AppointmentService) func(fiber.Router) {
	return func(router fiber.Router) {
		router.Get("/health", httpserver.Health)

		router.Get("/doctors/:doctorId/free-slots", func(c *fiber.Ctx) error {
			doctorID, err := httpserver.ParamID(c, "doctorId")
			if err != nil {
				return httpserver.WriteError(c, err)
			}
			slots, err := svc.FreeSlots(c.UserContext(), doctorID, c.Query("date"))
			return httpserver.Reply(c, slots, err)
		})
		router.Get("/appointments/:id", func(c *fiber.Ctx) error {
			id, err := httpserver.ParamID(c, "id")
			if err != nil {
				return httpserver.WriteError(c, err)
			}
			appointment, err := svc.GetAppointment(c.UserContext(), id)
			return httpserver.Reply(c, appointment, err)
		})
		router.Get("/patients/:patientId/appointments", func(c *fiber.Ctx) error {
			patientID, err := httpserver.ParamID(c, "patientId")
			if err != nil {
				return httpserver.WriteError(c, err)
			}
			status := (*domain.Status)(httpserver.QueryString(c, "status"))
			items, err := svc.ListPatientAppointments(c.UserContext(), patientID, status)
			return httpserver.Reply(c, items, err)
		})
	}
}
