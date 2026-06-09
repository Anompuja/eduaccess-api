package infrastructure

import (
	"context"
	"encoding/json"
	"time"

	"github.com/eduaccess/eduaccess-api/internal/notification/domain"
	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type notificationModel struct {
	ID        uuid.UUID  `gorm:"column:id;primaryKey;default:gen_random_uuid()"`
	SchoolID  *uuid.UUID `gorm:"column:school_id"`
	UserID    uuid.UUID  `gorm:"column:user_id"`
	Type      string     `gorm:"column:type"`
	Title     string     `gorm:"column:title"`
	Body      string     `gorm:"column:body"`
	Data      []byte     `gorm:"column:data;type:jsonb"`
	ReadAt    *time.Time `gorm:"column:read_at"`
	CreatedAt time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time  `gorm:"column:updated_at;autoUpdateTime"`
}

func (notificationModel) TableName() string { return "notifications" }

type GormNotificationRepository struct{ db *gorm.DB }

func NewGormNotificationRepository(db *gorm.DB) *GormNotificationRepository {
	return &GormNotificationRepository{db: db}
}

func (r *GormNotificationRepository) Create(ctx context.Context, n *domain.Notification) error {
	dataBytes, err := json.Marshal(n.Data)
	if err != nil {
		dataBytes = []byte("{}")
	}
	row := &notificationModel{
		SchoolID: n.SchoolID,
		UserID:   n.UserID,
		Type:     n.Type,
		Title:    n.Title,
		Body:     n.Body,
		Data:     dataBytes,
	}
	if err := r.db.WithContext(ctx).Create(row).Error; err != nil {
		return err
	}
	n.ID = row.ID
	n.CreatedAt = row.CreatedAt
	n.UpdatedAt = row.UpdatedAt
	return nil
}

func (r *GormNotificationRepository) ListByUserID(ctx context.Context, userID uuid.UUID, unreadOnly bool) ([]*domain.Notification, error) {
	q := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(50)

	if unreadOnly {
		q = q.Where("read_at IS NULL")
	}

	var rows []notificationModel
	if err := q.Find(&rows).Error; err != nil {
		return nil, err
	}

	result := make([]*domain.Notification, 0, len(rows))
	for _, row := range rows {
		n := rowToDomain(row)
		result = append(result, n)
	}
	return result, nil
}

func (r *GormNotificationRepository) MarkRead(ctx context.Context, id, userID uuid.UUID) error {
	res := r.db.WithContext(ctx).
		Model(&notificationModel{}).
		Where("id = ? AND user_id = ? AND read_at IS NULL", id, userID).
		Update("read_at", time.Now())
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return apperror.New(apperror.ErrNotFound, "notification not found or already read")
	}
	return nil
}

func (r *GormNotificationRepository) MarkAllRead(ctx context.Context, userID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&notificationModel{}).
		Where("user_id = ? AND read_at IS NULL", userID).
		Update("read_at", time.Now()).Error
}

// FindParentUserIDsByStudentUserID resolves auth.users UUID → student_profiles.id
// → student_parent_links → parent_profiles.user_id in a single query.
func (r *GormNotificationRepository) FindParentUserIDsByStudentUserID(ctx context.Context, studentUserID uuid.UUID) ([]uuid.UUID, error) {
	type row struct {
		UserID uuid.UUID `gorm:"column:user_id"`
	}

	var rows []row
	err := r.db.WithContext(ctx).Raw(`
		SELECT pp.user_id
		FROM student_profiles sp
		JOIN student_parent_links spl ON spl.student_id = sp.id
		JOIN parent_profiles pp       ON pp.id = spl.parent_id
		WHERE sp.user_id = ?
		  AND sp.deleted_at IS NULL
		  AND pp.deleted_at IS NULL
	`, studentUserID).Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	ids := make([]uuid.UUID, len(rows))
	for i, r := range rows {
		ids[i] = r.UserID
	}
	return ids, nil
}

func rowToDomain(row notificationModel) *domain.Notification {
	var data map[string]any
	if len(row.Data) > 0 {
		_ = json.Unmarshal(row.Data, &data)
	}
	if data == nil {
		data = map[string]any{}
	}
	return &domain.Notification{
		ID:        row.ID,
		SchoolID:  row.SchoolID,
		UserID:    row.UserID,
		Type:      row.Type,
		Title:     row.Title,
		Body:      row.Body,
		Data:      data,
		ReadAt:    row.ReadAt,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}
