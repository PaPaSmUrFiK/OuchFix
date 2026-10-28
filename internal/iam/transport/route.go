package transport

import (
	"github.com/gofiber/fiber/v2"

	httpserver "github.com/PaPaSmUrFiK/OuchFix/pkg/transport/http/server"
)

func New() func(fiber.Router) {
	return func(router fiber.Router) {
		router.Get("/health", httpserver.Health)
	}
}
