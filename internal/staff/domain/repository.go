package domain

import (
	"context"

	"github.com/google/uuid"
)

// StaffRepository defines the contract for staff persistence.
type StaffRepository interface {
	CreateStaffProfile(ctx context.Context, staff *StaffProfile) error
	FindStaffByID(ctx context.Context, id uuid.UUID) (*StaffProfile, error)
	UpdateStaffProfile(ctx context.Context, staff *StaffProfile) error
	SoftDeleteStaff(ctx context.Context, id uuid.UUID) error
	ListStaff(ctx context.Context, filter StaffFilter) ([]*StaffProfile, int64, error)
}

// StaffFilter is used for filtering and pagination in list operations.
type StaffFilter struct {
	SchoolID *uuid.UUID
	Search   string // searches by name, email, or username
	Offset   int
	Limit    int
}
