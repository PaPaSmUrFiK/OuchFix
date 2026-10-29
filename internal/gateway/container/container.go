package container

import (
	"net/http"

	"github.com/PaPaSmUrFiK/OuchFix/internal/gateway/config"
	gql "github.com/PaPaSmUrFiK/OuchFix/internal/gateway/transport/graphql"
	"github.com/PaPaSmUrFiK/OuchFix/internal/gateway/transport/graphql/resolver"
	httpclient "github.com/PaPaSmUrFiK/OuchFix/pkg/transport/http/client"
)

// Services — HTTP-прокси gateway к сервисам: каждый знает только базовый адрес
// сервиса. Своей БД у gateway нет.
type Services struct {
	proxies resolver.Services
}

func (c *Container) NewServices() *Services {
	s := c.Config.Services
	return &Services{
		proxies: resolver.Services{
			IAM:          httpclient.New(s.IAMBaseURL, s.Timeout),
			Catalog:      httpclient.New(s.CatalogBaseURL, s.Timeout),
			Appointment:  httpclient.New(s.AppointmentBaseURL, s.Timeout),
			Payment:      httpclient.New(s.PaymentBaseURL, s.Timeout),
			Notification: httpclient.New(s.NotificationBaseURL, s.Timeout),
		},
	}
}

type Handlers struct {
	GraphQL http.Handler
}

func (c *Container) NewHandlers() *Handlers {
	return &Handlers{
		GraphQL: gql.NewHandler(resolver.New(c.services.proxies, c.Config.TestUserID)),
	}
}

// Container — composition root gateway: прокси к сервисам и обработчик GraphQL.
type Container struct {
	Config   *config.Config
	services *Services
	Handlers *Handlers
}

func New(cfg *config.Config) *Container {
	c := &Container{Config: cfg}

	c.services = c.NewServices()
	c.Handlers = c.NewHandlers()

	return c
}
