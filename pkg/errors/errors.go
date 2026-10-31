package errors

import "errors"

// Доменные ошибки. Доезжают до клиента: транспорт переводит их в свой формат
// только на границе (GraphQL-ответ), внутри оборачиваются через %w.
var (
	ErrInvalidArgument = errors.New("некорректный аргумент")
	ErrNotFound        = errors.New("не найдено")
	ErrNotImplemented  = errors.New("не реализовано")
)
