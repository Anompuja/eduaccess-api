package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/eduaccess/eduaccess-api/internal/auth/domain"
	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// userJoinQuery is the reusable JOIN SQL for loading a user with their role and school.
//
// Strategy for users linked to multiple schools:
// - Prefer schools with status = 'active'
// - If more than one remains, pick the most recently linked membership
//
// This keeps school context deterministic for JWT generation.
const userJoinQuery = `
SELECT
    u.*,
    r.id   AS role_id,
    r.name AS role_name,
		preferred_school.school_id
FROM users u
LEFT JOIN model_has_roles mhr ON mhr.user_id = u.id
LEFT JOIN roles r              ON r.id = mhr.role_id
LEFT JOIN LATERAL (
		SELECT su.school_id
		FROM school_users su
		LEFT JOIN schools s ON s.id = su.school_id
		WHERE su.user_id = u.id
			AND su.deleted_at IS NULL
			AND (s.id IS NULL OR s.deleted_at IS NULL)
		ORDER BY
			CASE WHEN s.status = 'active' THEN 0 ELSE 1 END,
			su.created_at DESC,
			su.school_id
		LIMIT 1
) preferred_school ON TRUE
WHERE u.deleted_at IS NULL
`

// GormUserRepository implements domain.UserRepository using GORM.
type GormUserRepository struct {
	db *gorm.DB
}

func NewGormUserRepository(db *gorm.DB) *GormUserRepository {
	return &GormUserRepository{db: db}
}

func (r *GormUserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var row userWithRole
	sql := fmt.Sprintf("%s AND u.email = ? LIMIT 1", userJoinQuery)
	if err := r.db.WithContext(ctx).Raw(sql, email).Scan(&row).Error; err != nil {
		return nil, err
	}
	if row.ID == (uuid.UUID{}) {
		return nil, apperror.ErrNotFound
	}
	return row.toDomain(), nil
}

func (r *GormUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	var row userWithRole
	sql := fmt.Sprintf("%s AND u.id = ? LIMIT 1", userJoinQuery)
	if err := r.db.WithContext(ctx).Raw(sql, id).Scan(&row).Error; err != nil {
		return nil, err
	}
	if row.ID == (uuid.UUID{}) {
		return nil, apperror.ErrNotFound
	}
	return row.toDomain(), nil
}

// Create inserts the user row, assigns the role, and links to a school — all in one transaction.
func (r *GormUserRepository) Create(ctx context.Context, user *domain.User) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Insert user row
		m := &userModel{
			ID:        user.ID,
			Name:      user.Name,
			Username:  user.Username,
			Email:     user.Email,
			Password:  user.Password,
			Avatar:    user.Avatar,
			Verified:  user.Verified,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		}
		if err := tx.Create(m).Error; err != nil {
			return err
		}

		// 2. Look up the role by name
		var role roleModel
		if err := tx.Where("name = ?", user.Role).First(&role).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return apperror.New(apperror.ErrBadRequest, "unknown role: "+user.Role)
			}
			return err
		}
		user.RoleID = &role.ID

		// 3. Assign role
		if err := tx.Create(&modelHasRole{UserID: user.ID, RoleID: role.ID}).Error; err != nil {
			return err
		}

		// 4. Link to school (skip for superadmin)
		if user.SchoolID != nil {
			su := &schoolUserModel{
				ID:        uuid.New(),
				UserID:    user.ID,
				SchoolID:  *user.SchoolID,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			if err := tx.Create(su).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *GormUserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&userModel{}).
		Where("email = ? AND deleted_at IS NULL", email).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *GormUserRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&userModel{}).
		Where("username = ? AND deleted_at IS NULL", username).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
