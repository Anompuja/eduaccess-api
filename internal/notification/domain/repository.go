package domain

import (
	"context"

	"github.com/google/uuid"
)

type NotificationRepository interface {
	Create(ctx context.Context, n *Notification) error
	ListByUserID(ctx context.Context, userID uuid.UUID, unreadOnly bool) ([]*Notification, error)
	MarkRead(ctx context.Context, id, userID uuid.UUID) error
	MarkAllRead(ctx context.Context, userID uuid.UUID) error
	// FindParentUserIDsByStudentUserID returns user_id of all parents linked to the
	// student identified by their auth.users UUID (class_schedule_students.student_id).
	FindParentUserIDsByStudentUserID(ctx context.Context, studentUserID uuid.UUID) ([]uuid.UUID, error)
}
