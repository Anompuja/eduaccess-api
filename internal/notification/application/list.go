package application

import (
	"context"

	"github.com/eduaccess/eduaccess-api/internal/notification/domain"
	"github.com/google/uuid"
)

type ListNotificationsCommand struct {
	UserID     uuid.UUID
	UnreadOnly bool
}

type ListNotificationsHandler struct{ repo domain.NotificationRepository }

func NewListNotificationsHandler(repo domain.NotificationRepository) *ListNotificationsHandler {
	return &ListNotificationsHandler{repo: repo}
}

func (h *ListNotificationsHandler) Handle(ctx context.Context, cmd ListNotificationsCommand) ([]*domain.Notification, error) {
	return h.repo.ListByUserID(ctx, cmd.UserID, cmd.UnreadOnly)
}
