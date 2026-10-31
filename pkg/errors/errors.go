package errors

import "errors"

// Доменные ошибки. Доезжают до клиента: транспорт переводит их в свой формат
// только на границе (GraphQL-ответ), внутри оборачиваются через %w.
var (
	ErrInvalidArgument = errors.New("некорректный аргумент")
	ErrNotFound        = errors.New("не найдено")
	ErrNotImplemented  = errors.New("не реализовано")
	// ErrConflict — операция противоречит уже существующим данным (слот занят,
	// приём уже оплачен).
	ErrConflict = errors.New("конфликт")
	// ErrFailedPrecondition — объект в состоянии, в котором операция запрещена
	// (отмена завершённого приёма, нехватка остатка).
	ErrFailedPrecondition = errors.New("операция недопустима в текущем состоянии")
)
