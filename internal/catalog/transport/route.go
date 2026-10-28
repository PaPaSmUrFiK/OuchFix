package transport

import (
	"github.com/gofiber/fiber/v2"

	"github.com/PaPaSmUrFiK/OuchFix/internal/catalog/service"
	httpserver "github.com/PaPaSmUrFiK/OuchFix/pkg/transport/http/server"
)

// New — внутренний REST API каталога для gateway. В ЛР6 заменится на gRPC.
func New(svc *service.CatalogService) func(fiber.Router) {
	return func(router fiber.Router) {
		router.Get("/health", httpserver.Health)

		router.Get("/specialties", func(c *fiber.Ctx) error {
			items, err := svc.ListSpecialties(c.UserContext())
			return httpserver.Reply(c, items, err)
		})
		router.Get("/doctors", func(c *fiber.Ctx) error {
			specialtyID, err := httpserver.QueryID(c, "specialtyId")
			if err != nil {
				return httpserver.WriteError(c, err)
			}
			items, err := svc.ListDoctors(c.UserContext(), specialtyID)
			return httpserver.Reply(c, items, err)
		})
		router.Get("/patients/by-user/:userId", func(c *fiber.Ctx) error {
			userID, err := httpserver.ParamID(c, "userId")
			if err != nil {
				return httpserver.WriteError(c, err)
			}
			patient, err := svc.GetPatientByUserID(c.UserContext(), userID)
			return httpserver.Reply(c, patient, err)
		})
		router.Get("/medical-services", func(c *fiber.Ctx) error {
			items, err := svc.ListMedicalServices(c.UserContext())
			return httpserver.Reply(c, items, err)
		})
		router.Get("/medications", func(c *fiber.Ctx) error {
			items, err := svc.ListMedications(c.UserContext())
			return httpserver.Reply(c, items, err)
		})
	}
}
