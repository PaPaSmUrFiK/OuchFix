package container

import (
	"net/http"

	"github.com/PaPaSmUrFiK/OuchFix/internal/gateway/client"
	"github.com/PaPaSmUrFiK/OuchFix/internal/gateway/config"
	gql "github.com/PaPaSmUrFiK/OuchFix/internal/gateway/transport/graphql"
	"github.com/PaPaSmUrFiK/OuchFix/internal/gateway/transport/graphql/resolver"
	httpclient "github.com/PaPaSmUrFiK/OuchFix/pkg/transport/http/client"
)

// Clients — клиенты gateway к сервисам. Своей БД у gateway нет: данные только
// через внутренние REST API сервисов (в ЛР6 — gRPC).
type Clients struct {
	catalog      *client.Catalog
	appointment  *client.Appointment
	payment      *client.Payment
	notification *client.Notification
}

func (c *Container) NewClients() *Clients {
	s := c.Config.Services
	return &Clients{
		catalog:      client.NewCatalog(httpclient.New(s.CatalogURL, s.Timeout)),
		appointment:  client.NewAppointment(httpclient.New(s.AppointmentURL, s.Timeout)),
		payment:      client.NewPayment(httpclient.New(s.PaymentURL, s.Timeout)),
		notification: client.NewNotification(httpclient.New(s.NotificationURL, s.Timeout)),
	}
}

// Container — composition root gateway: клиенты сервисов и транспорт.
type Container struct {
	Config  *config.Config
	clients *Clients
}

func New(cfg *config.Config) *Container {
	c := &Container{Config: cfg}
	c.clients = c.NewClients()

	return c
}

// GraphQLHandler — внешний API gateway. Резолверы получают сервисы как
// интерфейсы: в ЛР3 это HTTP-клиенты, в ЛР6 — gRPC-клиенты.
func (c *Container) GraphQLHandler() http.Handler {
	return gql.NewHandler(resolver.New(
		c.clients.catalog,
		c.clients.appointment,
		c.clients.payment,
		c.clients.notification,
		c.Config.Location,
		c.Config.TestUserID,
	))
}
