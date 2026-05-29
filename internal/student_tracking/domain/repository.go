package domain

import (
	"context"

	"github.com/google/uuid"
)

type StudyFilter struct {
	SchoolID       *uuid.UUID
	ClassroomID    *uuid.UUID
	AcademicYearID *uuid.UUID
	ClassID        *uuid.UUID
	StudentID      *uuid.UUID
	Status         *string
}

type Repository interface {
	ListStudies(ctx context.Context, filter StudyFilter) ([]StudyView, error)
}
