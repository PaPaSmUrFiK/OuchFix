package service

import (
	"context"
	"fmt"

	"github.com/PaPaSmUrFiK/OuchFix/internal/notification/domain"
	"github.com/PaPaSmUrFiK/OuchFix/pkg/client/iam"
	pkgerrors "github.com/PaPaSmUrFiK/OuchFix/pkg/errors"
)

type NotificationRepository interface {
	ListByUser(ctx context.Context, userID int64, status *domain.Status) ([]domain.Notification, error)
}

// IAMClient — то, что Notification спрашивает у IAM (IamGrpc).
type IAMClient interface {
	GetUser(ctx context.Context, userID int64) (*iam.UserReply, error)
}

type NotificationService struct {
	repo NotificationRepository
	iam  IAMClient
}

func NewNotificationService(repo NotificationRepository, iam IAMClient) *NotificationService {
	return &NotificationService{repo: repo, iam: iam}
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

	if _, err := s.iam.GetUser(ctx, userID); err != nil {
		return nil, fmt.Errorf("check user %d: %w", userID, err)
	}

	return s.repo.ListByUser(ctx, userID, status)
}
