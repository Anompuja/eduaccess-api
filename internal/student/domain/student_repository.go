package domain

import (
	"context"

	"github.com/google/uuid"
)

// StudentFilter holds list query parameters for students.
type StudentFilter struct {
	SchoolID         *uuid.UUID
	EducationLevelID *uuid.UUID
	ClassID          *uuid.UUID
	SubClassID       *uuid.UUID
	Search           string // name, email, username, NIS, NISN
	Offset           int
	Limit            int
}

// StudentProfileRepository handles student profile persistence.
type StudentProfileRepository interface {
	CreateStudentProfile(ctx context.Context, profile *StudentProfile) error
	FindStudentByID(ctx context.Context, id uuid.UUID) (*StudentProfile, error)
	FindStudentByUserID(ctx context.Context, userID uuid.UUID) (*StudentProfile, error)
	ListStudents(ctx context.Context, f StudentFilter) ([]*StudentProfile, int64, error)
	UpdateStudentProfile(ctx context.Context, profile *StudentProfile) error
	SoftDeleteStudent(ctx context.Context, id uuid.UUID) error

	// AutoEnrollStudent opens an active student_studies record for a freshly
	// created student by matching their class/sub-class to a classroom in the
	// school's active academic year. Best-effort: returns nil (no-op) when no
	// active year or matching classroom exists.
	AutoEnrollStudent(ctx context.Context, schoolID, studentUserID uuid.UUID, classID, subClassID *uuid.UUID) error
}
