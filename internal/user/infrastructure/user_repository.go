package infrastructure

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/eduaccess/eduaccess-api/internal/auth/domain"
	authInfra "github.com/eduaccess/eduaccess-api/internal/auth/infrastructure"
	"github.com/eduaccess/eduaccess-api/internal/user/application"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GormUserManagementRepository implements both UserReadRepository and UserWriteRepository.
// It re-uses the auth infrastructure's GormUserRepository for FindByID and adds List,
// Update, and SoftDelete on top.
type GormUserManagementRepository struct {
	db   *gorm.DB
	auth *authInfra.GormUserRepository
}

func NewGormUserManagementRepository(db *gorm.DB) *GormUserManagementRepository {
	return &GormUserManagementRepository{
		db:   db,
		auth: authInfra.NewGormUserRepository(db),
	}
}

func (r *GormUserManagementRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return r.auth.FindByID(ctx, id)
}

func (r *GormUserManagementRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	return r.auth.ExistsByEmail(ctx, email)
}

func (r *GormUserManagementRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	return r.auth.ExistsByUsername(ctx, username)
}

func (r *GormUserManagementRepository) List(ctx context.Context, f application.ListFilter) ([]*domain.User, int64, error) {
	base := `
FROM users u
LEFT JOIN model_has_roles mhr ON mhr.user_id = u.id
LEFT JOIN roles r              ON r.id = mhr.role_id
LEFT JOIN school_users su      ON su.user_id = u.id AND su.deleted_at IS NULL
WHERE u.deleted_at IS NULL`

	args := []interface{}{}
	conditions := []string{}

	if f.SchoolID != nil {
		conditions = append(conditions, "su.school_id = ?")
		args = append(args, *f.SchoolID)
	}
	if f.Role != "" {
		conditions = append(conditions, "r.name = ?")
		args = append(args, f.Role)
	}
	if f.Search != "" {
		conditions = append(conditions, "(u.name ILIKE ? OR u.email ILIKE ? OR u.username ILIKE ?)")
		like := "%" + f.Search + "%"
		args = append(args, like, like, like)
	}

	where := ""
	if len(conditions) > 0 {
		where = " AND " + strings.Join(conditions, " AND ")
	}

	// Count
	var total int64
	countSQL := fmt.Sprintf("SELECT COUNT(DISTINCT u.id) %s%s", base, where)
	if err := r.db.WithContext(ctx).Raw(countSQL, args...).Scan(&total).Error; err != nil {
		return nil, 0, err
	}

	// Data
	type row struct {
		ID        uuid.UUID  `gorm:"column:id"`
		SchoolID  *uuid.UUID `gorm:"column:school_id"`
		RoleID    *uuid.UUID `gorm:"column:role_id"`
		RoleName  *string    `gorm:"column:role_name"`
		Name      string     `gorm:"column:name"`
		Username  string     `gorm:"column:username"`
		Email     string     `gorm:"column:email"`
		Password  string     `gorm:"column:password"`
		Avatar    string     `gorm:"column:avatar"`
		Verified  bool       `gorm:"column:verified"`
		DeletedAt *time.Time `gorm:"column:deleted_at"`
		CreatedAt time.Time  `gorm:"column:created_at"`
		UpdatedAt time.Time  `gorm:"column:updated_at"`
	}

	dataSQL := fmt.Sprintf(`
SELECT DISTINCT ON (u.id)
    u.id, u.name, u.username, u.email, u.password, u.avatar, u.verified,
    u.deleted_at, u.created_at, u.updated_at,
    r.id AS role_id, r.name AS role_name, su.school_id
%s%s
ORDER BY u.id, u.created_at DESC
LIMIT ? OFFSET ?`, base, where)

	queryArgs := append(args, f.Limit, f.Offset)
	var rows []row
	if err := r.db.WithContext(ctx).Raw(dataSQL, queryArgs...).Scan(&rows).Error; err != nil {
		return nil, 0, err
	}

	users := make([]*domain.User, 0, len(rows))
	for _, rw := range rows {
		u := &domain.User{
			ID:        rw.ID,
			SchoolID:  rw.SchoolID,
			RoleID:    rw.RoleID,
			Name:      rw.Name,
			Username:  rw.Username,
			Email:     rw.Email,
			Password:  rw.Password,
			Avatar:    rw.Avatar,
			Verified:  rw.Verified,
			DeletedAt: rw.DeletedAt,
			CreatedAt: rw.CreatedAt,
			UpdatedAt: rw.UpdatedAt,
		}
		if rw.RoleName != nil {
			u.Role = *rw.RoleName
		}
		users = append(users, u)
	}

	return users, total, nil
}

func (r *GormUserManagementRepository) Update(ctx context.Context, user *domain.User) error {
	return r.db.WithContext(ctx).
		Table("users").
		Where("id = ?", user.ID).
		Updates(map[string]interface{}{
			"name":       user.Name,
			"avatar":     user.Avatar,
			"password":   user.Password,
			"updated_at": user.UpdatedAt,
			"username":   user.Username,
			"email":      user.Email,
		}).Error
}

func (r *GormUserManagementRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Table("users").
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"deleted_at": now,
			"updated_at": now,
		}).Error
}
