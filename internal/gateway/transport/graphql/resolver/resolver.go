package resolver

import (
	"context"
	"net/url"
)

// Service — HTTP-прокси к внутреннему REST API одного сервиса: GET (чтение) и
// POST (бизнес-операции) по пути относительно базового адреса. Сервисы отвечают JSON в форме GraphQL-моделей,
// поэтому ответ декодируется сразу в model.* без преобразований, а ошибки
// сервиса (NOT_FOUND, BAD_USER_INPUT, ...) проходят как есть.
type Service interface {
	Get(ctx context.Context, path string, query url.Values, out any) error
	Post(ctx context.Context, path string, body any, out any) error
}

// Services — прокси ко всем сервисам. IAM пока не используется: me, login и
// регистрация появятся в ЛР5.
type Services struct {
	IAM          Service
	Catalog      Service
	Appointment  Service
	Payment      Service
	Notification Service
}

type Resolver struct {
	services Services
	// currentUserID — ВРЕМЕННО до ЛР5: «текущий пользователь» для my*-запросов
	// берётся из конфига (GATEWAY_TEST_USER_ID), а не из JWT.
	currentUserID int64
}

func New(services Services, currentUserID int64) *Resolver {
	return &Resolver{services: services, currentUserID: currentUserID}
}
