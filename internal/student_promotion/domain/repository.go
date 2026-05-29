package domain

import (
	"context"

	"github.com/google/uuid"
)

type PromotionFilter struct {
	SchoolID       *uuid.UUID
	StudentID      *uuid.UUID
	AcademicYearID *uuid.UUID
}

type Repository interface {
	ListPromotions(ctx context.Context, filter PromotionFilter) ([]PromotionView, error)
	FindClassroomTarget(ctx context.Context, classroomID uuid.UUID) (*ClassroomTarget, error)
	PromoteStudent(ctx context.Context, in PromotionInput) error
}
