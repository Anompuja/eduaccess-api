package domain

import (
	"time"

	"github.com/google/uuid"
)

// Role name constants match the seeded rows in the roles table.
const (
	RoleSuperadmin    = "superadmin"
	RoleAdminSekolah  = "admin_sekolah"
	RoleKepalaSekolah = "kepala_sekolah"
	RoleGuru          = "guru"
	RoleStaff         = "staff"
	RoleOrangTua      = "orangtua"
	RoleSiswa         = "siswa"
)

// User is the domain aggregate for authentication and user management.
// Role and SchoolID are denormalized from the join tables model_has_roles and school_users.
type User struct {
	ID       uuid.UUID
	SchoolID *uuid.UUID // nil for superadmin; loaded from school_users
	RoleID   *uuid.UUID // loaded from model_has_roles
	Role     string     // loaded from roles.name

	Name     string
	Username string
	Email    string
	Password string // bcrypt hash
	Avatar   string
	Verified bool

	DeletedAt *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

// IsActive returns true when the user has not been soft-deleted.
func (u *User) IsActive() bool {
	return u.DeletedAt == nil
}

// RefreshToken is the domain entity for the refresh_tokens table.
type RefreshToken struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Token     string
	ExpiresAt time.Time
	CreatedAt time.Time
}
