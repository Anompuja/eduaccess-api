package domain

import (
	"context"

	"github.com/google/uuid"
)

// ParentFilter holds list query parameters for parents.
type ParentFilter struct {
	SchoolID *uuid.UUID
	Search   string // name, email, username
	Offset   int
	Limit    int
}

// StudentRepository handles student and parent profile persistence.
type StudentRepository interface {
	StudentProfileRepository

	// Parent profiles
	CreateParentProfile(ctx context.Context, profile *ParentProfile) error
	FindParentByID(ctx context.Context, id uuid.UUID) (*ParentProfile, error)
	FindParentByUserID(ctx context.Context, userID uuid.UUID) (*ParentProfile, error)
	ListParents(ctx context.Context, f ParentFilter) ([]*ParentProfile, int64, error)
	UpdateParentProfile(ctx context.Context, profile *ParentProfile) error
	SoftDeleteParent(ctx context.Context, id uuid.UUID) error

	// Parent links
	LinkParent(ctx context.Context, link *ParentLink) error
	UnlinkParent(ctx context.Context, studentID, parentID uuid.UUID) error
	ListParentLinks(ctx context.Context, studentID uuid.UUID) ([]*ParentLink, error)
}
