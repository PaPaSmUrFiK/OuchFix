package service

import (
	"context"
	"fmt"

	"github.com/PaPaSmUrFiK/OuchFix/internal/notification/domain"
	pkgerrors "github.com/PaPaSmUrFiK/OuchFix/pkg/errors"
)

type NotificationRepository interface {
	ListByUser(ctx context.Context, userID int64, status *domain.Status) ([]domain.Notification, error)
}

type NotificationService struct {
	repo NotificationRepository
}

func NewNotificationService(repo NotificationRepository) *NotificationService {
	return &NotificationService{repo: repo}
}

func (s *NotificationService) ListUserNotifications(
	ctx context.Context,
	userID int64,
	status *domain.Status,
) ([]domain.Notification, error) {
	if err := pkgerrors.ValidateID("userId", userID); err != nil {
		return nil, err
	}
	if status != nil && !status.Valid() {
		return nil, fmt.Errorf("%w: неизвестный статус уведомления %q", pkgerrors.ErrInvalidArgument, *status)
	}

	return s.repo.ListByUser(ctx, userID, status)
}
