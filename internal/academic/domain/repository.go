package domain

import (
	"context"

	"github.com/google/uuid"
)

// AcademicRepository handles education levels, classes, and sub-classes.
type AcademicRepository interface {
	// Education levels
	CreateLevel(ctx context.Context, level *EducationLevel) error
	FindLevelByID(ctx context.Context, id uuid.UUID) (*EducationLevel, error)
	ListLevels(ctx context.Context, schoolID uuid.UUID) ([]*EducationLevel, error)
	UpdateLevel(ctx context.Context, level *EducationLevel) error
	SoftDeleteLevel(ctx context.Context, id uuid.UUID) error

	// Classes
	CreateClass(ctx context.Context, class *Class) error
	FindClassByID(ctx context.Context, id uuid.UUID) (*Class, error)
	ListClasses(ctx context.Context, schoolID uuid.UUID, levelID *uuid.UUID) ([]*Class, error)
	UpdateClass(ctx context.Context, class *Class) error
	SoftDeleteClass(ctx context.Context, id uuid.UUID) error

	// Sub-classes
	CreateSubClass(ctx context.Context, sub *SubClass) error
	FindSubClassByID(ctx context.Context, id uuid.UUID) (*SubClass, error)
	ListSubClasses(ctx context.Context, schoolID uuid.UUID, classID *uuid.UUID) ([]*SubClass, error)
	UpdateSubClass(ctx context.Context, sub *SubClass) error
	SoftDeleteSubClass(ctx context.Context, id uuid.UUID) error
}
