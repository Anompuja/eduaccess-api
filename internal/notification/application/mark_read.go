package application

import (
	"context"

	"github.com/eduaccess/eduaccess-api/internal/notification/domain"
	"github.com/google/uuid"
)

type MarkReadCommand struct {
	NotificationID uuid.UUID
	UserID         uuid.UUID
}

type MarkReadHandler struct{ repo domain.NotificationRepository }

func NewMarkReadHandler(repo domain.NotificationRepository) *MarkReadHandler {
	return &MarkReadHandler{repo: repo}
}

func (h *MarkReadHandler) Handle(ctx context.Context, cmd MarkReadCommand) error {
	return h.repo.MarkRead(ctx, cmd.NotificationID, cmd.UserID)
}

type MarkAllReadCommand struct {
	UserID uuid.UUID
}

type MarkAllReadHandler struct{ repo domain.NotificationRepository }

func NewMarkAllReadHandler(repo domain.NotificationRepository) *MarkAllReadHandler {
	return &MarkAllReadHandler{repo: repo}
}

func (h *MarkAllReadHandler) Handle(ctx context.Context, cmd MarkAllReadCommand) error {
	return h.repo.MarkAllRead(ctx, cmd.UserID)
}
