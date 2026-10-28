package app

import (
	"context"

	"github.com/PaPaSmUrFiK/OuchFix/internal/notification/container"
	"github.com/PaPaSmUrFiK/OuchFix/internal/notification/transport"
	httpserver "github.com/PaPaSmUrFiK/OuchFix/pkg/transport/http/server"
)

func Run(ctx context.Context, ctr *container.Container) error {
	srv := httpserver.New(ctr.Config.Server.HTTP(), transport.New(ctr.Service()))

	return srv.Run(ctx)
}
