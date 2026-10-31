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

		router.Post("/appointments", func(c fiber.Ctx) error {
			var req bookRequest
			if err := httpserver.BindJSON(c, &req); err != nil {
				return httpserver.WriteError(c, err)
			}
			in, err := req.toInput()
			if err != nil {
				return httpserver.WriteError(c, err)
			}
			return replyAppointment(c, loc)(svc.Book(c.Context(), in))
		})
		router.Post("/appointments/walk-in", func(c fiber.Ctx) error {
			var req walkInRequest
			if err := httpserver.BindJSON(c, &req); err != nil {
				return httpserver.WriteError(c, err)
			}
			in, err := req.toInput()
			if err != nil {
				return httpserver.WriteError(c, err)
			}
			return replyAppointment(c, loc)(svc.WalkIn(c.Context(), in))
		})
		router.Post("/appointments/:id/cancel", func(c fiber.Ctx) error {
			id, err := httpserver.ParamID(c, "id")
			if err != nil {
				return httpserver.WriteError(c, err)
			}
			return replyAppointment(c, loc)(svc.Cancel(c.Context(), id))
		})
		router.Post("/appointments/:id/reschedule", func(c fiber.Ctx) error {
			id, err := httpserver.ParamID(c, "id")
			if err != nil {
				return httpserver.WriteError(c, err)
			}
			var req rescheduleRequest
			if err := httpserver.BindJSON(c, &req); err != nil {
				return httpserver.WriteError(c, err)
			}
			slotID, err := httpserver.ParseID("newSlotId", req.NewSlotID)
			if err != nil {
				return httpserver.WriteError(c, err)
			}
			return replyAppointment(c, loc)(svc.Reschedule(c.Context(), id, slotID))
		})
		router.Post("/appointments/:id/complete", func(c fiber.Ctx) error {
			id, err := httpserver.ParamID(c, "id")
			if err != nil {
				return httpserver.WriteError(c, err)
			}
			var req completeRequest
			if err := httpserver.BindJSON(c, &req); err != nil {
				return httpserver.WriteError(c, err)
			}
			serviceID, err := httpserver.ParseOptionalID("serviceId", req.ServiceID)
			if err != nil {
				return httpserver.WriteError(c, err)
			}
			in := service.CompleteInput{AppointmentID: id, ServiceID: serviceID, Conclusion: req.Conclusion}
			return replyAppointment(c, loc)(svc.Complete(c.Context(), in))
		})
		router.Post("/appointments/:id/diagnoses", func(c fiber.Ctx) error {
			id, err := httpserver.ParamID(c, "id")
			if err != nil {
				return httpserver.WriteError(c, err)
			}
			var req diagnosisRequest
			if err := httpserver.BindJSON(c, &req); err != nil {
				return httpserver.WriteError(c, err)
			}
			d, err := svc.AddDiagnosis(c.Context(), service.DiagnosisInput{AppointmentID: id, Code: req.Code, Description: req.Description})
			if err != nil {
				return httpserver.WriteError(c, err)
			}
			return httpserver.Reply(c, toDiagnosis(d), nil)
		})
		router.Post("/appointments/:id/prescriptions", func(c fiber.Ctx) error {
			id, err := httpserver.ParamID(c, "id")
			if err != nil {
				return httpserver.WriteError(c, err)
			}
			var req prescriptionRequest
			if err := httpserver.BindJSON(c, &req); err != nil {
				return httpserver.WriteError(c, err)
			}
			medicationID, err := httpserver.ParseOptionalID("medicationId", req.MedicationID)
			if err != nil {
				return httpserver.WriteError(c, err)
			}
			p, err := svc.AddPrescription(c.Context(), service.PrescriptionInput{AppointmentID: id, Text: req.Text, MedicationID: medicationID})
			if err != nil {
				return httpserver.WriteError(c, err)
			}
			return httpserver.Reply(c, toPrescription(p), nil)
		})
	}
}

// replyAppointment отдаёт приём, возвращённый бизнес-операцией, или её ошибку.
func replyAppointment(c fiber.Ctx, loc *time.Location) func(*domain.Appointment, error) error {
	return func(a *domain.Appointment, err error) error {
		if err != nil {
			return httpserver.WriteError(c, err)
		}
		return httpserver.Reply(c, toAppointment(loc)(a), nil)
	}
}
