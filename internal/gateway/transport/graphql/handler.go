package graphql

import (
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"

	"github.com/PaPaSmUrFiK/OuchFix/internal/gateway/transport/graphql/generated"
	"github.com/PaPaSmUrFiK/OuchFix/internal/gateway/transport/graphql/resolver"
)

// NewHandler собирает HTTP-обработчик GraphQL: POST и GET-запросы, интроспекция
// для playground. Подписок и загрузки файлов в схеме нет.
func NewHandler(r *resolver.Resolver) http.Handler {
	srv := handler.New(generated.NewExecutableSchema(generated.Config{Resolvers: r}))
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})
	srv.Use(extension.Introspection{})

	return srv
}
