package errors

import "fmt"

// ValidateID проверяет идентификатор сущности: в БД это BIGSERIAL, значит > 0.
func ValidateID(name string, id int64) error {
	if id <= 0 {
		return fmt.Errorf("%w: %s должен быть положительным, получено %d", ErrInvalidArgument, name, id)
	}
	return nil
}
