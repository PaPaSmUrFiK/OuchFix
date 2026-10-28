package transport

import (
	"net/http"

	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"

	httpserver "github.com/PaPaSmUrFiK/OuchFix/pkg/transport/http/server"
)

func New(graphql http.Handler) func(fiber.Router) {
	return func(router fiber.Router) {
		router.Get("/health", httpserver.Health)

		gql := adaptor.HTTPHandler(graphql)
		router.Post("/graphql", gql)
		router.Get("/graphql", gql)
		router.Get("/playground", adaptor.HTTPHandlerFunc(playground.Handler("OuchFix", "/graphql")))
	}
}
