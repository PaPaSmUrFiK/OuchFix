package transport

import (
	"github.com/gofiber/fiber/v3"

	"github.com/PaPaSmUrFiK/OuchFix/internal/catalog/domain"
	"github.com/PaPaSmUrFiK/OuchFix/internal/catalog/service"
	httpserver "github.com/PaPaSmUrFiK/OuchFix/pkg/transport/http/server"
)

// New — внутренний REST API каталога, к которому gateway проксирует запросы.
func New(svc *service.CatalogService) func(fiber.Router) {
	return func(router fiber.Router) {
		router.Get("/health", httpserver.Health)

		router.Get("/specialties", func(c fiber.Ctx) error {
			items, err := svc.ListSpecialties(c.Context())
			return httpserver.Reply(c, httpserver.MapList(items, toSpecialty), err)
		})
		router.Get("/doctors", func(c fiber.Ctx) error {
			specialtyID, err := httpserver.QueryID(c, "specialtyId")
			if err != nil {
				return httpserver.WriteError(c, err)
			}
			items, err := svc.ListDoctors(c.Context(), specialtyID)
			return httpserver.Reply(c, httpserver.MapList(items, toDoctor), err)
		})
		router.Get("/patients/by-user/:userId", func(c fiber.Ctx) error {
			userID, err := httpserver.ParamID(c, "userId")
			if err != nil {
				return httpserver.WriteError(c, err)
			}
			patient, err := svc.GetPatientByUserID(c.Context(), userID)
			if err != nil {
				return httpserver.WriteError(c, err)
			}
			return httpserver.Reply(c, toPatient(patient), nil)
		})
		router.Get("/medical-services", func(c fiber.Ctx) error {
			items, err := svc.ListMedicalServices(c.Context())
			return httpserver.Reply(c, httpserver.MapList(items, toMedicalService), err)
		})
		router.Get("/medications", func(c fiber.Ctx) error {
			items, err := svc.ListMedications(c.Context())
			return httpserver.Reply(c, httpserver.MapList(items, toMedication), err)
		})
		router.Post("/medications", func(c fiber.Ctx) error {
			var req upsertMedicationRequest
			if err := httpserver.BindJSON(c, &req); err != nil {
				return httpserver.WriteError(c, err)
			}
			m, err := svc.UpsertMedication(c.Context(), req.toDomain())
			return replyMedication(c, m, err)
		})
		router.Post("/medications/dispense", func(c fiber.Ctx) error {
			var req dispenseRequest
			if err := httpserver.BindJSON(c, &req); err != nil {
				return httpserver.WriteError(c, err)
			}
			d, err := req.toDomain()
			if err != nil {
				return httpserver.WriteError(c, err)
			}
			m, err := svc.DispenseMedication(c.Context(), d)
			return replyMedication(c, m, err)
		})
	}
}

func replyMedication(c fiber.Ctx, m *domain.Medication, err error) error {
	if err != nil {
		return httpserver.WriteError(c, err)
	}
	return httpserver.Reply(c, toMedication(m), nil)
}
