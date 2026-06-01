package domain

import (
	"context"

	"github.com/google/uuid"
)


type HeadmasterFilter struct {
	SchoolID *uuid.UUID
	Search   string 
	Offset   int
	Limit    int
}


type HeadmasterRepository interface {
	CreateHeadmasterProfile(ctx context.Context, profile *HeadmasterProfile) error
	FindHeadmasterByID(ctx context.Context, id uuid.UUID) (*HeadmasterProfile, error)
	FindHeadmasterByUserID(ctx context.Context, userID uuid.UUID) (*HeadmasterProfile, error)
	ListHeadmasters(ctx context.Context, f HeadmasterFilter) ([]*HeadmasterProfile, int64, error)
	UpdateHeadmasterProfile(ctx context.Context, profile *HeadmasterProfile) error
	SoftDeleteHeadmaster(ctx context.Context, id uuid.UUID) error
}
