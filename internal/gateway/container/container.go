package container

import (
	"context"

	"github.com/PaPaSmUrFiK/OuchFix/internal/gateway/config"
)

// Container — единственный composition root gateway: здесь собираются БД,
// репозитории, сервисы, клиенты и транспорт.
type Container struct {
	Config *config.Config
}

func New(_ context.Context, cfg *config.Config) (*Container, error) {
	return &Container{Config: cfg}, nil
}

func (c *Container) Close() error {
	return nil
}
