package client

import (
	"context"
	"net/url"
	"strconv"

	"github.com/PaPaSmUrFiK/OuchFix/internal/notification/domain"
	httpclient "github.com/PaPaSmUrFiK/OuchFix/pkg/transport/http/client"
)

type Notification struct {
	http *httpclient.Client
}

func NewNotification(http *httpclient.Client) *Notification {
	return &Notification{http: http}
}

func (c *Notification) ListUserNotifications(ctx context.Context, userID int64, status *domain.Status) ([]domain.Notification, error) {
	query := url.Values{}
	if status != nil {
		query.Set("status", string(*status))
	}
	var out []domain.Notification
	path := "/users/" + strconv.FormatInt(userID, 10) + "/notifications"
	return out, c.http.Get(ctx, path, query, &out)
}
