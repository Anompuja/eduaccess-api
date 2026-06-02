package domain

import (
	"time"

	"github.com/google/uuid"
)

// StudentPromotion is an audit record of a student moving between classrooms.
type StudentPromotion struct {
	ID              uuid.UUID
	SchoolID        uuid.UUID
	StudentID       uuid.UUID
	FromClassroomID uuid.UUID
	ToClassroomID   uuid.UUID
	AcademicYearID  uuid.UUID
	PromotionDate   time.Time
	Status          string // promoted, retained, transferred
	Notes           string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// PromotionView is an enriched read model for promotion history.
type PromotionView struct {
	ID                uuid.UUID `gorm:"column:id"`
	StudentID         uuid.UUID `gorm:"column:student_id"`
	StudentName       string    `gorm:"column:student_name"`
	NIS               string    `gorm:"column:nis"`
	FromClassroomID   uuid.UUID `gorm:"column:from_classroom_id"`
	FromClassroomName string    `gorm:"column:from_classroom_name"`
	ToClassroomID     uuid.UUID `gorm:"column:to_classroom_id"`
	ToClassroomName   string    `gorm:"column:to_classroom_name"`
	AcademicYearID    uuid.UUID `gorm:"column:school_academic_year_id"`
	AcademicYearName  string    `gorm:"column:academic_year_name"`
	PromotionDate     time.Time `gorm:"column:promotion_date"`
	Status            string    `gorm:"column:status"`
	Notes             string    `gorm:"column:notes"`
}

// ClassroomTarget describes the destination classroom of a promotion, resolved
// to the class/sub-class/academic-year/homeroom it implies.
type ClassroomTarget struct {
	ClassroomID       uuid.UUID
	SchoolID          uuid.UUID
	ClassID           *uuid.UUID
	SubClassID        *uuid.UUID
	AcademicYearID    *uuid.UUID
	HomeroomTeacherID *uuid.UUID
	EducationLevelID  *uuid.UUID
}

// PromotionInput is the resolved data needed to promote a single student.
type PromotionInput struct {
	SchoolID      uuid.UUID
	StudentID     uuid.UUID
	Target        ClassroomTarget
	PromotionDate time.Time
	Status        string
	Notes         string
}
