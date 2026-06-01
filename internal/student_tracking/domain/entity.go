package domain

import (
	"time"

	"github.com/google/uuid"
)

// StudentStudy is a student's enrollment in one classroom for one academic year.
type StudentStudy struct {
	ID                uuid.UUID
	SchoolID          uuid.UUID
	StudentID         uuid.UUID
	ClassroomID       uuid.UUID
	AcademicYearID    uuid.UUID
	ClassID           *uuid.UUID
	SubClassID        *uuid.UUID
	HomeroomTeacherID *uuid.UUID
	Status            string // active, inactive, graduated, transferred
	EnrollmentDate    time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// StudyView is an enriched read model for the tracking screens — it carries the
// human-readable names the UI needs without forcing the client to resolve IDs.
type StudyView struct {
	ID               uuid.UUID  `gorm:"column:id"`
	StudentID        uuid.UUID  `gorm:"column:student_id"`
	StudentName      string     `gorm:"column:student_name"`
	NIS              string     `gorm:"column:nis"`
	ClassroomID      uuid.UUID  `gorm:"column:school_classroom_id"`
	ClassroomName    string     `gorm:"column:classroom_name"`
	ClassID          *uuid.UUID `gorm:"column:school_class_id"`
	ClassName        string     `gorm:"column:class_name"`
	SubClassName     string     `gorm:"column:sub_class_name"`
	AcademicYearID   uuid.UUID  `gorm:"column:school_academic_year_id"`
	AcademicYearName string     `gorm:"column:academic_year_name"`
	Status           string     `gorm:"column:status"`
	EnrollmentDate   time.Time  `gorm:"column:enrollment_date"`
}
