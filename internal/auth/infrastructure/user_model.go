package infrastructure

import (
	"time"

	"github.com/eduaccess/eduaccess-api/internal/auth/domain"
	"github.com/google/uuid"
)

// userModel maps to the users table.
type userModel struct {
	ID               uuid.UUID  `gorm:"type:uuid;primaryKey"`
	Name             string
	Username         string
	Email            string
	Password         string
	Avatar           string
	QrCode           *string
	EmailVerifiedAt  *time.Time
	VerificationCode *string
	Verified         bool
	DeletedAt        *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func (userModel) TableName() string { return "users" }

// userWithRole is a scan target for the JOIN query used by FindByEmail/FindByID.
type userWithRole struct {
	userModel
	RoleID   *uuid.UUID `gorm:"column:role_id"`
	RoleName *string    `gorm:"column:role_name"`
	SchoolID *uuid.UUID `gorm:"column:school_id"`
}

func (r *userWithRole) toDomain() *domain.User {
	u := &domain.User{
		ID:        r.ID,
		SchoolID:  r.SchoolID,
		RoleID:    r.RoleID,
		Name:      r.Name,
		Username:  r.Username,
		Email:     r.Email,
		Password:  r.Password,
		Avatar:    r.Avatar,
		Verified:  r.Verified,
		DeletedAt: r.DeletedAt,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
	if r.RoleName != nil {
		u.Role = *r.RoleName
	}
	return u
}

// roleModel maps to the roles table (used for role_id lookup).
type roleModel struct {
	ID   uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name string
}

func (roleModel) TableName() string { return "roles" }

// modelHasRole maps to the model_has_roles join table.
type modelHasRole struct {
	UserID uuid.UUID `gorm:"type:uuid;column:user_id"`
	RoleID uuid.UUID `gorm:"type:uuid;column:role_id"`
}

func (modelHasRole) TableName() string { return "model_has_roles" }

// schoolUserModel maps to the school_users join table.
type schoolUserModel struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID  `gorm:"type:uuid;column:user_id"`
	SchoolID  uuid.UUID  `gorm:"type:uuid;column:school_id"`
	DeletedAt *time.Time `gorm:"column:deleted_at"`
	CreatedAt time.Time  `gorm:"column:created_at"`
	UpdatedAt time.Time  `gorm:"column:updated_at"`
}

func (schoolUserModel) TableName() string { return "school_users" }

// refreshTokenModel maps to the refresh_tokens table.
type refreshTokenModel struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID `gorm:"type:uuid"`
	Token     string
	ExpiresAt time.Time
	CreatedAt time.Time
}

func (refreshTokenModel) TableName() string { return "refresh_tokens" }

func refreshTokenModelFromDomain(rt *domain.RefreshToken) *refreshTokenModel {
	return &refreshTokenModel{
		ID:        rt.ID,
		UserID:    rt.UserID,
		Token:     rt.Token,
		ExpiresAt: rt.ExpiresAt,
		CreatedAt: rt.CreatedAt,
	}
}

func (m *refreshTokenModel) toDomain() *domain.RefreshToken {
	return &domain.RefreshToken{
		ID:        m.ID,
		UserID:    m.UserID,
		Token:     m.Token,
		ExpiresAt: m.ExpiresAt,
		CreatedAt: m.CreatedAt,
	}
}
