package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type ClassScheduleFilter struct {
	SchoolID    *uuid.UUID
	ClassroomID *uuid.UUID
	TeacherID   *uuid.UUID
	SubjectID   *uuid.UUID
	Date        *time.Time
	Status      *string
}

type ClassScheduleRepository interface {
	CreateClassSchedule(ctx context.Context, cs *ClassSchedule) error
	FindClassScheduleByID(ctx context.Context, id uuid.UUID) (*ClassSchedule, error)
	ListClassSchedules(ctx context.Context, filter ClassScheduleFilter) ([]*ClassSchedule, error)
	UpdateClassSchedule(ctx context.Context, cs *ClassSchedule) error
	SoftDeleteClassSchedule(ctx context.Context, id uuid.UUID) error

	StartClassSchedule(ctx context.Context, id uuid.UUID, teacherTime time.Time) error
	CompleteClassSchedule(ctx context.Context, id uuid.UUID) error
	CancelClassSchedule(ctx context.Context, id uuid.UUID) error

	AutoPopulateStudents(ctx context.Context, scheduleID uuid.UUID, schoolID uuid.UUID, classroomID uuid.UUID) error
	SyncStudents(ctx context.Context, scheduleID uuid.UUID, schoolID uuid.UUID, classroomID uuid.UUID) error
	ListAttendances(ctx context.Context, scheduleID uuid.UUID) ([]*ClassScheduleStudent, error)
	FindAttendance(ctx context.Context, scheduleID uuid.UUID, studentID uuid.UUID) (*ClassScheduleStudent, error)
	UpdateAttendance(ctx context.Context, att *ClassScheduleStudent) error
}
