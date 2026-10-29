package app

import (
	"context"

	"github.com/PaPaSmUrFiK/OuchFix/internal/payment/container"
	httpserver "github.com/PaPaSmUrFiK/OuchFix/pkg/transport/http/server"
)

func Run(ctx context.Context, ctr *container.Container) error {
	srv := httpserver.New(ctr.Config.Server.HTTP(), ctr.Handlers.Router)

	return srv.Run(ctx)
}
