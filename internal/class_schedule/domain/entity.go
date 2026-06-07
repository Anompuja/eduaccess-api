package domain

import (
	"time"

	"github.com/google/uuid"
)

type ClassSchedule struct {
	ID                    uuid.UUID
	SchoolID              uuid.UUID
	ClassroomID           uuid.UUID
	SubjectID             uuid.UUID
	TeacherID             uuid.UUID
	StartPeriodID         *uuid.UUID
	EndPeriodID           *uuid.UUID
	Date                  time.Time
	StartTime             string
	EndTime               string
	TeacherAttendanceTime *time.Time
	Status                string // scheduled, ongoing, completed, cancelled
	DeletedAt             *time.Time
	CreatedAt             time.Time
	UpdatedAt             time.Time

	// Enriched via JOIN — not stored in class_schedules
	ClassroomName     string
	SubjectName       string
	TeacherName       string
	StartPeriodNumber *int
	StartPeriodLabel  *string
	EndPeriodNumber   *int
}

type ClassScheduleStudent struct {
	ID                    uuid.UUID
	SchoolID              uuid.UUID
	ClassScheduleID       uuid.UUID
	StudentID             uuid.UUID
	Type                  string
	PhotoPath             string
	Note                  string
	StudentAttendanceTime *time.Time
	Status                string // scheduled, present, sick, permission, absent
	DeletedAt             *time.Time
	CreatedAt             time.Time
	UpdatedAt             time.Time

	// Enriched via JOIN — not stored in class_schedule_students
	StudentName string
}
