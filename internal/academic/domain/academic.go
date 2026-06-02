package domain

import (
	"time"

	"github.com/google/uuid"
)

// EducationLevel (e.g. SD, SMP, SMA) scoped to a school.
type EducationLevel struct {
	ID        uuid.UUID
	SchoolID  uuid.UUID
	Name      string
	DeletedAt *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Class belongs to an EducationLevel (e.g. Kelas 7, Kelas 8).
type Class struct {
	ID               uuid.UUID
	SchoolID         uuid.UUID
	EducationLevelID uuid.UUID
	Name             string
	DeletedAt        *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// SubClass belongs to a Class (e.g. 7A, 7B).
type SubClass struct {
	ID        uuid.UUID
	SchoolID  uuid.UUID
	ClassID   uuid.UUID
	Name      string
	DeletedAt *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

// AcademicYear represents a school academic year (e.g. 2024/2025).
type AcademicYear struct {
	ID          uuid.UUID
	SchoolID    uuid.UUID
	Name        string
	StartDate   time.Time
	EndDate     time.Time
	IsActive    bool
	Description string
	DeletedAt   *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Subject represents a school subject (Mata Pelajaran).
type Subject struct {
	ID               uuid.UUID
	SchoolID         uuid.UUID
	EducationLevelID *uuid.UUID // optional: scope subject to a specific education level
	Name             string
	Code             *string
	Category         string // core, elective, extracurricular, specialized, vocational
	DeletedAt        *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// Classroom represents a physical classroom (Ruang Kelas).
type Classroom struct {
	ID                uuid.UUID
	SchoolID          uuid.UUID
	ClassID           *uuid.UUID // FK school_classes
	SubClassID        *uuid.UUID // FK school_sub_classes
	AcademicYearID    *uuid.UUID // FK school_academic_years
	HomeroomTeacherID *uuid.UUID // FK users (wali kelas)
	Name              string
	CodeRoom          string
	Capacity          int
	Floor             string
	Building          string
	RoomType          string
	Status            string // unknown, available, occupied, maintenance
	Facilities        string
	DeletedAt         *time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type Schedule struct {
	ID           uuid.UUID
	SchoolID     uuid.UUID
	DayOfWeek    string // monday, tuesday, wednesday, thursday, friday, saturday, sunday
	PeriodNumber int
	Label        string // "Jam 1", "Istirahat", "Jam 4"
	StartTime    string // "07:00"
	EndTime      string // "07:45"
	IsBreak      bool
	DeletedAt    *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
