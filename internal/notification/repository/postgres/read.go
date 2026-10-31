package postgres

import (
	"context"
	"fmt"

	"github.com/PaPaSmUrFiK/OuchFix/internal/notification/domain"
	"github.com/PaPaSmUrFiK/OuchFix/internal/notification/repository/postgres/models"
	"github.com/PaPaSmUrFiK/OuchFix/pkg/storage/postgres"
)

const notificationColumns = `id, user_id, channel, title, body, status, related_type, related_id, created_at, sent_at`

// MarkRead переводит доставленное или ожидающее уведомление пользователя в read
// (read → read не меняет ничего). ErrNotFound — уведомления нет, оно чужое
// или не доставлено (failed): различает сервис через Get.
func (r *NotificationRepository) MarkRead(ctx context.Context, userID, id int64) (*domain.Notification, error) {
	const query = `
		UPDATE notification.notification SET status = 'read'
		WHERE id = $1 AND user_id = $2 AND status IN ('pending', 'sent', 'read')
		RETURNING ` + notificationColumns

	var row models.Notification
	if err := r.db.GetContext(ctx, &row, query, id, userID); err != nil {
		return nil, fmt.Errorf("mark notification %d read: %w", id, postgres.WrapDB(err))
	}
	out := row.ToDomain()
	return &out, nil
}

// Get — уведомление пользователя; чужое — ErrNotFound.
func (r *NotificationRepository) Get(ctx context.Context, userID, id int64) (*domain.Notification, error) {
	const query = `SELECT ` + notificationColumns + ` FROM notification.notification WHERE id = $1 AND user_id = $2`

	var row models.Notification
	if err := r.db.GetContext(ctx, &row, query, id, userID); err != nil {
		return nil, fmt.Errorf("get notification %d: %w", id, postgres.WrapDB(err))
	}
	out := row.ToDomain()
	return &out, nil
}
