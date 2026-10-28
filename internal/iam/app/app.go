package app

import (
	"context"

	"github.com/PaPaSmUrFiK/OuchFix/internal/iam/container"
	"github.com/PaPaSmUrFiK/OuchFix/internal/iam/transport"
	httpserver "github.com/PaPaSmUrFiK/OuchFix/pkg/transport/http/server"
)

func Run(ctx context.Context, ctr *container.Container) error {
	srv := httpserver.New(ctr.Config.Server, transport.New())

	return srv.Run(ctx)
}
