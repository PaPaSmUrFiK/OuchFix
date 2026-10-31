package transport

import (
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/PaPaSmUrFiK/OuchFix/internal/notification/domain"
	"github.com/PaPaSmUrFiK/OuchFix/internal/notification/service"
	httpserver "github.com/PaPaSmUrFiK/OuchFix/pkg/transport/http/server"
)

// New — внутренний REST API уведомлений, к которому gateway проксирует запросы.
// loc — пояс медцентра, в нём отдаётся время.
func New(svc *service.NotificationService, loc *time.Location) func(fiber.Router) {
	return func(router fiber.Router) {
		router.Get("/health", httpserver.Health)

		router.Get("/users/:userId/notifications", func(c fiber.Ctx) error {
			userID, err := httpserver.ParamID(c, "userId")
			if err != nil {
				return httpserver.WriteError(c, err)
			}
			status := httpserver.QueryEnum[domain.Status](c, "status")
			items, err := svc.ListUserNotifications(c.Context(), userID, status)
			return httpserver.Reply(c, httpserver.MapList(items, toNotification(loc)), err)
		})
		router.Post("/users/:userId/notifications/:id/read", func(c fiber.Ctx) error {
			userID, err := httpserver.ParamID(c, "userId")
			if err != nil {
				return httpserver.WriteError(c, err)
			}
			id, err := httpserver.ParamID(c, "id")
			if err != nil {
				return httpserver.WriteError(c, err)
			}
			n, err := svc.MarkRead(c.Context(), userID, id)
			if err != nil {
				return httpserver.WriteError(c, err)
			}
			return httpserver.Reply(c, toNotification(loc)(n), nil)
		})
	}
}
