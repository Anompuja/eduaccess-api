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

// ParentRepository handles parent profile persistence.
type ParentRepository interface {
	ListParents(ctx context.Context, f ParentFilter) ([]*ParentProfile, int64, error)
}
