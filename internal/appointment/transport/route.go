package transport

import (
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/PaPaSmUrFiK/OuchFix/internal/appointment/domain"
	"github.com/PaPaSmUrFiK/OuchFix/internal/appointment/service"
	httpserver "github.com/PaPaSmUrFiK/OuchFix/pkg/transport/http/server"
)

// New — внутренний REST API записи на приём, к которому gateway проксирует запросы.
// loc — пояс медцентра, в нём отдаётся время.
func New(svc *service.AppointmentService, loc *time.Location) func(fiber.Router) {
	return func(router fiber.Router) {
		router.Get("/health", httpserver.Health)

		router.Get("/doctors/:doctorId/free-slots", func(c fiber.Ctx) error {
			doctorID, err := httpserver.ParamID(c, "doctorId")
			if err != nil {
				return httpserver.WriteError(c, err)
			}
			slots, err := svc.FreeSlots(c.Context(), doctorID, c.Query("date"))
			return httpserver.Reply(c, httpserver.MapList(slots, toSlot(loc)), err)
		})
		router.Get("/appointments/:id", func(c fiber.Ctx) error {
			id, err := httpserver.ParamID(c, "id")
			if err != nil {
				return httpserver.WriteError(c, err)
			}
			appointment, err := svc.GetAppointment(c.Context(), id)
			if err != nil {
				return httpserver.WriteError(c, err)
			}
			return httpserver.Reply(c, toAppointment(loc)(appointment), nil)
		})
		router.Get("/patients/:patientId/appointments", func(c fiber.Ctx) error {
			patientID, err := httpserver.ParamID(c, "patientId")
			if err != nil {
				return httpserver.WriteError(c, err)
			}
			status := httpserver.QueryEnum[domain.Status](c, "status")
			items, err := svc.ListPatientAppointments(c.Context(), patientID, status)
			return httpserver.Reply(c, httpserver.MapList(items, toAppointment(loc)), err)
		})

		router.Post("/slots", func(c fiber.Ctx) error {
			var req createSlotsRequest
			if err := httpserver.BindJSON(c, &req); err != nil {
				return httpserver.WriteError(c, err)
			}
			in, err := req.toInput()
			if err != nil {
				return httpserver.WriteError(c, err)
			}
			slots, err := svc.CreateSlots(c.Context(), in)
			return httpserver.Reply(c, httpserver.MapList(slots, toSlot(loc)), err)
		})
		router.Post("/slots/update", func(c fiber.Ctx) error {
			var req updateSlotsRequest
			if err := httpserver.BindJSON(c, &req); err != nil {
				return httpserver.WriteError(c, err)
			}
			in, err := req.toInput()
			if err != nil {
				return httpserver.WriteError(c, err)
			}
			slots, err := svc.UpdateSlots(c.Context(), in)
			return httpserver.Reply(c, httpserver.MapList(slots, toSlot(loc)), err)
		})
		router.Post("/slots/cancel", func(c fiber.Ctx) error {
			var req cancelSlotsRequest
			if err := httpserver.BindJSON(c, &req); err != nil {
				return httpserver.WriteError(c, err)
			}
			ids, err := httpserver.ParseIDs("slotIds", req.SlotIDs)
			if err != nil {
				return httpserver.WriteError(c, err)
			}
			slots, err := svc.CancelSlots(c.Context(), ids)
			return httpserver.Reply(c, httpserver.MapList(slots, toSlot(loc)), err)
		})
	}
}
