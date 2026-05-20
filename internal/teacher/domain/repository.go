package domain

import (
	"context"

	"github.com/google/uuid"
)

// TeacherRepository defines the contract for teacher persistence.
type TeacherRepository interface {
	CreateTeacherProfile(ctx context.Context, teacher *TeacherProfile) error
	FindTeacherByID(ctx context.Context, id uuid.UUID) (*TeacherProfile, error)
	UpdateTeacherProfile(ctx context.Context, teacher *TeacherProfile) error
	SoftDeleteTeacher(ctx context.Context, id uuid.UUID) error
	ListTeachers(ctx context.Context, filter TeacherFilter) ([]*TeacherProfile, int64, error)
}

// TeacherFilter is used for filtering and pagination in list operations.
type TeacherFilter struct {
	SchoolID uuid.UUID
	Search   string // searches by name, email, or username
	Offset   int
	Limit    int
}
