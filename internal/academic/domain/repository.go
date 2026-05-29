package domain

import (
	"context"

	"github.com/google/uuid"
)

// AcademicRepository handles education levels, classes, sub-classes, and related academic entities.
type AcademicRepository interface {
	// Education levels
	CreateLevel(ctx context.Context, level *EducationLevel) error
	FindLevelByID(ctx context.Context, id uuid.UUID) (*EducationLevel, error)
	ListLevels(ctx context.Context, schoolID *uuid.UUID) ([]*EducationLevel, error)
	UpdateLevel(ctx context.Context, level *EducationLevel) error
	SoftDeleteLevel(ctx context.Context, id uuid.UUID) error

	// Classes
	CreateClass(ctx context.Context, class *Class) error
	FindClassByID(ctx context.Context, id uuid.UUID) (*Class, error)
	ListClasses(ctx context.Context, schoolID *uuid.UUID, levelID *uuid.UUID) ([]*Class, error)
	UpdateClass(ctx context.Context, class *Class) error
	SoftDeleteClass(ctx context.Context, id uuid.UUID) error

	// Sub-classes
	CreateSubClass(ctx context.Context, sub *SubClass) error
	FindSubClassByID(ctx context.Context, id uuid.UUID) (*SubClass, error)
	ListSubClasses(ctx context.Context, schoolID *uuid.UUID, classID *uuid.UUID) ([]*SubClass, error)
	UpdateSubClass(ctx context.Context, sub *SubClass) error
	SoftDeleteSubClass(ctx context.Context, id uuid.UUID) error

	// Academic years
	CreateAcademicYear(ctx context.Context, ay *AcademicYear) error
	FindAcademicYearByID(ctx context.Context, id uuid.UUID) (*AcademicYear, error)
	ListAcademicYears(ctx context.Context, schoolID *uuid.UUID) ([]*AcademicYear, error)
	UpdateAcademicYear(ctx context.Context, ay *AcademicYear) error
	SoftDeleteAcademicYear(ctx context.Context, id uuid.UUID) error
	ActivateAcademicYear(ctx context.Context, id uuid.UUID, schoolID uuid.UUID) error

	// Subjects
	CreateSubject(ctx context.Context, s *Subject) error
	FindSubjectByID(ctx context.Context, id uuid.UUID) (*Subject, error)
	ListSubjects(ctx context.Context, schoolID *uuid.UUID) ([]*Subject, error)
	UpdateSubject(ctx context.Context, s *Subject) error
	SoftDeleteSubject(ctx context.Context, id uuid.UUID) error

	// Classrooms
	CreateClassroom(ctx context.Context, c *Classroom) error
	FindClassroomByID(ctx context.Context, id uuid.UUID) (*Classroom, error)
	ListClassrooms(ctx context.Context, schoolID *uuid.UUID) ([]*Classroom, error)
	UpdateClassroom(ctx context.Context, c *Classroom) error
	SoftDeleteClassroom(ctx context.Context, id uuid.UUID) error

	// Schedules
	CreateSchedule(ctx context.Context, s *Schedule) error
	FindScheduleByID(ctx context.Context, id uuid.UUID) (*Schedule, error)
	ListSchedules(ctx context.Context, schoolID *uuid.UUID, dayOfWeek *string) ([]*Schedule, error)
	UpdateSchedule(ctx context.Context, s *Schedule) error
	SoftDeleteSchedule(ctx context.Context, id uuid.UUID) error
}
