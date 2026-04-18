package domain

import "context"

import "github.com/google/uuid"

// AdminFilter holds list query parameters for admins.
type AdminFilter struct {
	SchoolID *uuid.UUID
	Search   string // name, email, username
	Offset   int
	Limit    int
}

// AdminRepository handles admin profile persistence.
type AdminRepository interface {
	CreateAdminProfile(ctx context.Context, profile *AdminProfile) error
	FindAdminByID(ctx context.Context, id uuid.UUID) (*AdminProfile, error)
	UpdateAdminProfile(ctx context.Context, profile *AdminProfile) error
	SoftDeleteAdmin(ctx context.Context, id uuid.UUID) error
	ListAdmins(ctx context.Context, f AdminFilter) ([]*AdminProfile, int64, error)
}
