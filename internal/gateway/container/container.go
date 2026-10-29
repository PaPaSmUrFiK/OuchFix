package container

import (
	"net/http"

	"github.com/PaPaSmUrFiK/OuchFix/internal/gateway/config"
	gql "github.com/PaPaSmUrFiK/OuchFix/internal/gateway/transport/graphql"
	"github.com/PaPaSmUrFiK/OuchFix/internal/gateway/transport/graphql/resolver"
	httpclient "github.com/PaPaSmUrFiK/OuchFix/pkg/transport/http/client"
)

// Clients — HTTP-прокси gateway к сервисам: каждый знает только базовый адрес
// сервиса. Своей БД у gateway нет.
type Clients struct {
	iam          *httpclient.Client
	catalog      *httpclient.Client
	appointment  *httpclient.Client
	payment      *httpclient.Client
	notification *httpclient.Client
}

func (c *Container) NewClients() *Clients {
	s := c.Config.Services
	return &Clients{
		iam:          httpclient.New(s.IAMBaseURL, s.Timeout),
		catalog:      httpclient.New(s.CatalogBaseURL, s.Timeout),
		appointment:  httpclient.New(s.AppointmentBaseURL, s.Timeout),
		payment:      httpclient.New(s.PaymentBaseURL, s.Timeout),
		notification: httpclient.New(s.NotificationBaseURL, s.Timeout),
	}
}

// Container — composition root gateway: прокси к сервисам и транспорт.
type Container struct {
	Config  *config.Config
	clients *Clients
}

func New(cfg *config.Config) *Container {
	c := &Container{Config: cfg}
	c.clients = c.NewClients()

	return c
}

// GraphQLHandler — внешний API gateway.
func (c *Container) GraphQLHandler() http.Handler {
	return gql.NewHandler(resolver.New(resolver.Services{
		IAM:          c.clients.iam,
		Catalog:      c.clients.catalog,
		Appointment:  c.clients.appointment,
		Payment:      c.clients.payment,
		Notification: c.clients.notification,
	}, c.Config.TestUserID))
}
