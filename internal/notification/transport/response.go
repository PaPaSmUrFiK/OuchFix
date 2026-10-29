package transport

import (
	"time"

	"github.com/PaPaSmUrFiK/OuchFix/internal/notification/domain"
	httpserver "github.com/PaPaSmUrFiK/OuchFix/pkg/transport/http/server"
)

// notificationResponse — в форме GraphQL-модели Notification: gateway декодирует её без преобразований.
type notificationResponse struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Body        string  `json:"body"`
	Status      string  `json:"status"`
	RelatedType *string `json:"relatedType"`
	RelatedID   *string `json:"relatedId"`
	CreatedAt   string  `json:"createdAt"`
}

func toNotification(loc *time.Location) func(*domain.Notification) notificationResponse {
	return func(n *domain.Notification) notificationResponse {
		return notificationResponse{
			ID:          httpserver.FormatID(n.ID),
			Title:       n.Title,
			Body:        n.Body,
			Status:      httpserver.FormatEnum(n.Status),
			RelatedType: n.RelatedType,
			RelatedID:   httpserver.FormatOptionalID(n.RelatedID),
			CreatedAt:   httpserver.FormatTime(n.CreatedAt, loc),
		}
	}
}
