package transport

import (
	"github.com/gofiber/fiber/v3"

	"github.com/PaPaSmUrFiK/OuchFix/internal/notification/domain"
	"github.com/PaPaSmUrFiK/OuchFix/internal/notification/service"
	httpserver "github.com/PaPaSmUrFiK/OuchFix/pkg/transport/http/server"
)

// New — внутренний REST API уведомлений для gateway. В ЛР6 заменится на gRPC.
func New(svc *service.NotificationService) func(fiber.Router) {
	return func(router fiber.Router) {
		router.Get("/health", httpserver.Health)

		router.Get("/users/:userId/notifications", func(c fiber.Ctx) error {
			userID, err := httpserver.ParamID(c, "userId")
			if err != nil {
				return httpserver.WriteError(c, err)
			}
			status := (*domain.Status)(httpserver.QueryString(c, "status"))
			items, err := svc.ListUserNotifications(c.Context(), userID, status)
			return httpserver.Reply(c, items, err)
		})
	}
}
