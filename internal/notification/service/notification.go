package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/PaPaSmUrFiK/OuchFix/internal/notification/domain"
	"github.com/PaPaSmUrFiK/OuchFix/pkg/client/iam"
	pkgerrors "github.com/PaPaSmUrFiK/OuchFix/pkg/errors"
)

type NotificationRepository interface {
	ListByUser(ctx context.Context, userID int64, status *domain.Status) ([]domain.Notification, error)
	MarkRead(ctx context.Context, userID, id int64) (*domain.Notification, error)
	Get(ctx context.Context, userID, id int64) (*domain.Notification, error)
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

// MarkRead отмечает уведомление пользователя прочитанным; повторный вызов ничего
// не меняет. Чужое уведомление — NOT_FOUND (не раскрываем, что оно существует),
// недоставленное (failed) — FAILED_PRECONDITION.
func (s *NotificationService) MarkRead(ctx context.Context, userID, id int64) (*domain.Notification, error) {
	if err := pkgerrors.ValidateID("userId", userID); err != nil {
		return nil, err
	}
	if err := pkgerrors.ValidateID("id", id); err != nil {
		return nil, err
	}
	n, err := s.repo.MarkRead(ctx, userID, id)
	if !errors.Is(err, pkgerrors.ErrNotFound) {
		return n, err
	}
	current, getErr := s.repo.Get(ctx, userID, id)
	if getErr != nil {
		return nil, getErr
	}
	return nil, fmt.Errorf("%w: уведомление в статусе %q нельзя отметить прочитанным", pkgerrors.ErrFailedPrecondition, current.Status)
}
