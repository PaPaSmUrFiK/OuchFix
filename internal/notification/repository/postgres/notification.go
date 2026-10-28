package postgres

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/PaPaSmUrFiK/OuchFix/internal/notification/domain"
	"github.com/PaPaSmUrFiK/OuchFix/internal/notification/repository/postgres/models"
	"github.com/PaPaSmUrFiK/OuchFix/pkg/storage/postgres"
)

type NotificationRepository struct {
	db *sqlx.DB
}

func NewNotificationRepository(db *sqlx.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

// ListByUser возвращает уведомления пользователя; status == nil — все.
func (r *NotificationRepository) ListByUser(ctx context.Context, userID int64, status *domain.Status) ([]domain.Notification, error) {
	const query = `
		SELECT id, user_id, channel, title, body, status, related_type, related_id, created_at, sent_at
		FROM notification.notification
		WHERE user_id = $1 AND ($2::VARCHAR IS NULL OR status = $2)
		ORDER BY created_at DESC, id DESC`

	var statusArg *string
	if status != nil {
		s := string(*status)
		statusArg = &s
	}

	rows := make([]models.Notification, 0)
	if err := r.db.SelectContext(ctx, &rows, query, userID, statusArg); err != nil {
		return nil, fmt.Errorf("list notifications of user %d: %w", userID, postgres.WrapDB(err))
	}

	return models.ToDomainList(rows), nil
}
